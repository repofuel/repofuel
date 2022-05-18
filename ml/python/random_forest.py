import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import RandomizedSearchCV


def build_model(X_train, y_train, params=None):
    if params is None:
        params = calculate_hyperparameters(X_train, y_train)

    model = RandomForestClassifier(n_estimators=params['n_estimators'],
                                   bootstrap=params['bootstrap'],
                                   max_depth=params['max_depth'],
                                   max_features=params['max_features'])

    model.fit(X_train, y_train)

    return model


def calculate_hyperparameters(X_train, y_train):
    param = {'n_estimators': [int(x) for x in np.linspace(start=10, stop=2000, num=10)],
             'bootstrap': [True, False],
             'max_depth': [int(x) for x in np.linspace(5, 100, num=10)],
             'max_features': ['auto', None]}

    # hyper_parameters
    return RandomizedSearchCV(estimator=RandomForestClassifier(),
                              param_distributions=param,
                              scoring='roc_auc',
                              cv=5,
                              random_state=120,
                              verbose=0,
                              n_jobs=1).fit(X_train, y_train).best_params_
