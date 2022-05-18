import io

import pandas as pd
import requests


class Collector:
    def __init__(self, ingest_url, auth, repo_id, start_job_id=None, last_job_id=None):
        if auth is None:
            raise Exception("messing the authentication")

        if ingest_url.endswith('/'):
            ingest_url = ingest_url[:-1]

        self.repo_url = f'{ingest_url}/repositories/{repo_id}'
        self.auth = auth
        self.repo_id = repo_id
        self.start_job_id = start_job_id
        self.last_job_id = last_job_id
        self._all_commit_metrics = None
        self._quantile_steps = [.5, .75, .9]

    def calculate_metrics_quantiles(self):
        return {
            "commit": self.calculate_commit_metrics_quantiles(),
            "file": self.calculate_file_metrics_quantiles(),
            "developer": self.calculate_developer_metrics_quantiles(),
        }

    def calculate_file_metrics_quantiles(self):
        df = self.file_aggregated_metrics()
        df = df.quantile(self._quantile_steps)
        return df.to_dict(orient="index")

    def calculate_developer_metrics_quantiles(self):
        df = self.developer_aggregated_metrics()
        df = df.quantile(self._quantile_steps)
        return df.to_dict(orient="index")

    def calculate_commit_metrics_quantiles(self):
        df = self.all_commit_metrics(cache=True)
        df = df.iloc[:, 3:].quantile(self._quantile_steps)
        return df.to_dict(orient="index")

    def file_aggregated_metrics(self):
        url = f'{self.repo_url}/file_aggregated_metrics.csv'
        return self._read_data(url)

    def developer_aggregated_metrics(self):
        url = f'{self.repo_url}/developer_aggregated_metrics.csv'
        return self._read_data(url)

    def all_commit_metrics(self, cache=True):
        if self._all_commit_metrics is not None:
            return self._all_commit_metrics

        params = {}
        if self.last_job_id:
            params["last_job"] = self.last_job_id

        url = f'{self.repo_url}/metrics.csv'
        df = self._read_data(url, params)
        if cache:
            self._all_commit_metrics = df
        return df

    def job_commit_metrics(self):
        params = {}
        if self.start_job_id:
            params["start_job"] = self.start_job_id

        if self.last_job_id:
            params["last_job"] = self.last_job_id

        url = f'{self.repo_url}/metrics.csv'
        return self._read_data(url, params)

    def _read_data(self, url, params=None):
        headers = {"Authorization": self.auth}
        resp = requests.get(url, params=params, headers=headers)
        resp.raise_for_status()
        return pd.read_csv(io.StringIO(resp.text))

    def delete_cache(self):
        self._all_commit_metrics = None
