import os

import numpy as np
import pandas as pd
from sklearn.metrics import classification_report
from sklearn.model_selection import train_test_split

import onnxutil
import random_forest as rf

experience = ['exp', 'rexp', 'sexp']
history = ['ndev', 'nuc', 'age']
size = ['la', 'ld', 'lt']
diffusion = ['ns', 'nd', 'nf', 'entropy']

features = experience + history + size + diffusion

dimensions = {
    "experience": experience,
    "history": history,
    "size": size,
    "diffusion": diffusion
}

max_training_rows = 10000
target = 'buggy'
date_column = 'author_date'
id_column = 'commit_id'
ignored_days = 90


def prepare_data(df, oldest_ignored_date):
    df = df.dropna(subset=features + [target], axis=0)
    df = df.loc[df[date_column] < oldest_ignored_date]

    if len(df) > max_training_rows:
        df = df.sort_values(date_column, ascending=False)[:max_training_rows]

    X = df.loc[:, features]
    y = df.loc[:, target].astype(str)

    # from sklearn.preprocessing import StandardScaler
    # X = StandardScaler().fit_transform(X) #??
    # X = (X - X.mean()) / X.std()#??

    # todo: reconsider the final training set size

    return X, y


def build_and_predict(collector, params, model_path):
    result = {}

    try:
        result["quantiles"] = collector.calculate_metrics_quantiles()
    except Exception as e:
        result["error"] = {
            "stage": "quantile calculation",
            "message": str(e)
        }
        return result

    df = collector.all_commit_metrics()
    latest_date = df[date_column].max()
    oldest_ignored_date = latest_date - (ignored_days * 86400)
    collector.delete_cache()

    try:
        X, y = prepare_data(df, oldest_ignored_date)
        medians = X.median()
        medians_buggy = X[y == "True"].median()
        medians_clean = X[y == "False"].median()
    except ValueError as e:
        result["error"] = {
            "stage": "data preparation",
            "message": str(e)
        }
        return result

    result["medians"] = {
        # todo: filing Nan with zeros could be problematic, if we have
        #  NaN, it means it is a crappy model and we should fail the model.
        "all": medians.fillna(0).to_dict(),
        "clean": medians_clean.fillna(0).to_dict(),
        "buggy": medians_buggy.fillna(0).to_dict(),
    }

    try:
        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.1, random_state=42)
    except ValueError as e:
        result["error"] = {
            "stage": "data splitting",
            "message": str(e)
        }
        return result

    # prediction data for newest commits (e.g, less than 90 days)
    df_predict = df.loc[df[date_column] >= oldest_ignored_date]

    result["data_points"] = {
        "all": len(df),
        "train": y_train.value_counts().to_dict(),
        "test": y_test.value_counts().to_dict(),
        "predict": len(df_predict),
    }

    try:
        model = rf.build_model(X_train, y_train, params)
    except Exception as e:
        result["error"] = {
            "stage": "model building",
            "message": str(e) + " " + str(type(e))
        }
        return result

    result["is_built"] = True

    # convert and save the model
    onx = onnxutil.convert_to_onnx(model)
    os.makedirs(os.path.dirname(model_path), exist_ok=True)
    with open(model_path, "wb") as f:
        f.write(onx.SerializeToString())

    # test the converted model
    onnx_model = onnxutil.Model(model_path)
    y_pred = onnx_model.predict(X_test)
    # convert the prediction probabilities to labels in order to test it
    y_pred = ['True' if v['True'] >= .5 else 'False' for v in y_pred]
    report = classification_report(y_test, y_pred, output_dict=True)
    medians_score = onnx_model.predict(pd.DataFrame([medians]))[0]["True"]

    result["model_report"] = {
        "format": "onnx",
        "params": model.get_params(),
        "feature_importance": match_ordered_lists(features, model.feature_importances_),
        "accuracy": report.get("accuracy", 0),
        "medians_score": medians_score,
        "buggy": report.get("True", {}),
        "clean": report.get("False", {}),
        "macro_avg": report.get("macro avg", {}),
        "weighted_avg": report.get("weighted avg", {}),
    }

    try:
        result["prediction"] = predict(onnx_model, df_predict, medians, medians_score).to_dict(orient='records')
    except Exception as e:
        result["error"] = {
            "stage": "predicting",
            "message": str(e)
        }
        return result

    result["is_predicted"] = True

    return result


def load_and_predict(collector, model, medians, medians_score):
    result = {}

    df = collector.job_commit_metrics()
    onnx_model = onnxutil.Model(model)
    medians = pd.DataFrame.from_dict([medians])

    try:
        result["prediction"] = predict(onnx_model, df, medians, medians_score).to_dict(orient='records')
    except Exception as e:
        result["error"] = {
            "stage": "predicting",
            "message": str(e)
        }
        return result

    result["is_predicted"] = True

    return result


def predict(onnx_model, df, medians, medians_score):
    df = df.reset_index()
    score = pd.DataFrame(onnx_model.predict(df.loc[:, features]))["True"]

    result = pd.DataFrame()
    for name, dimension_features in dimensions.items():
        dimension_df = df.loc[:, features]
        for feature in [f for f in features if f not in dimension_features]:
            dimension_df[feature] = medians[feature]
        result[name] = (pd.DataFrame(onnx_model.predict(dimension_df))["True"] - medians_score) / medians_score

    minimum = result.min(axis=1)
    minimum[minimum > 0] = 0  # todo: is it correct?
    minimum = minimum.abs()
    result = result.add(minimum, axis=0)
    total = result.sum(axis=1)
    result = result.div(total, axis=0).mul(score, axis=0)
    result["score"] = score
    result[id_column] = df[id_column]

    # fixme: replace "Infinity", "-Infinity", and "NaN" with Zero could be problematic
    return result.replace([np.inf, -np.inf, np.nan], 0)


def match_ordered_lists(l1, l2):
    r = dict()
    for i, v in enumerate(l1):
        r[v] = l2[i]
    return r
