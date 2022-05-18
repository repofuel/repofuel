import argparse
import io
import json
import sys

import pandas as pd
import requests  # todo: reconsider, maybe it worth to use urllib3 instead to remove an extra dependency

import data
import ml


def main():
    args = get_args()

    collector = data.Collector(args.ingest_url, args.auth, args.repo_id, args.start_job_id, args.last_job_id)

    if args.action == "build":
        output = ml.build_and_predict(collector, args.params, args.model)

    elif args.action == "predict":
        output = ml.load_and_predict(collector, args.model, args.medians, args.medians_score)

    else:
        raise Exception("unknown action")

    print(json.dumps(output))


def read_data(args):
    if args.auth is not None:
        headers = {"Authorization": args.auth}
        url = args.data
        content = requests.get(url, headers=headers).text
        return pd.read_csv(io.StringIO(content))

    if args.data is not None:
        return pd.read_csv(args.data)

    return pd.read_csv(sys.stdin)


def get_args():
    """Parse commandline."""
    parser = argparse.ArgumentParser()
    parser.add_argument("--ingest-url")
    parser.add_argument("--repo-id")
    parser.add_argument("--start-job-id")
    parser.add_argument("--last-job-id")
    parser.add_argument("--auth", help="HTTP Authorization header to fitch the data")

    action = parser.add_subparsers(dest="action")

    build = action.add_parser("build", help="build a model")
    build.add_argument("--params", type=json.loads, help="hyper-parameters for the model. "
                                                         "It should be in JSON format, if omitted, it will be "
                                                         "calculated using Random Search Cross Validation")
    build.add_argument("--model", help="the path where the model will be stored")

    predict = action.add_parser("predict", help="predict the riskiness of given commits")
    predict.add_argument("--model")
    predict.add_argument("--medians-score", type=float)
    predict.add_argument("--medians", type=json.loads)

    args = parser.parse_args()
    return args


if __name__ == '__main__':
    main()
