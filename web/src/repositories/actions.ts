import {Job, PullRequest, Repository, RepositoryAddress} from './types';
import {AnyAction} from 'redux';
import {ThunkAction} from 'redux-thunk';
import {AppState} from '../store/types';
import {checkStatus, parsJsonLines} from '../util/requests';
import {authRequest} from '../store/util';

export const FETCH_REPOSITORIES_REQUEST = 'FETCH_REPOSITORIES_REQUEST';
export const FETCH_REPOSITORIES_SUCCESS = 'FETCH_REPOSITORIES_SUCCESS';
export const FETCH_REPOSITORIES_FAIL = 'FETCH_REPOSITORIES_FAIL';

const fetchRepositoriesRequest = () => ({
  type: FETCH_REPOSITORIES_REQUEST,
});
const fetchRepositoriesSuccess = (repositories: Repository[]) => ({
  type: FETCH_REPOSITORIES_SUCCESS,
  repositories,
});
const fetchRepositoriesFail = (error: any) => ({
  type: FETCH_REPOSITORIES_FAIL,
  error,
});

//todo: refactor the following functions to remove duplications
export const fetchRepositories = (
  platform?: string,
  owner?: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchRepositoriesRequest());
    const url =
      platform && owner
        ? `/ingest/platforms/${platform}/users/${owner}/repos`
        : `/ingest/user/repos`;
    authRequest('GET', url)
      .then(parsJsonLines)
      .then((data: any) => dispatch(fetchRepositoriesSuccess(data)))
      .catch((e) => dispatch(fetchRepositoriesFail(e)));
  };
};

export const REFRESH_SOURCE_REQUEST = 'REFRESH_SOURCE_REQUEST';
export const REFRESH_SOURCE_SUCCESS = 'REFRESH_SOURCE_SUCCESS';
export const REFRESH_SOURCE_FAIL = 'REFRESH_SOURCE_FAIL';

const refreshSourceRequest = (addr: RepositoryAddress) => ({
  type: REFRESH_SOURCE_REQUEST,
  addr,
});
const refreshSourceSuccess = (addr: RepositoryAddress) => ({
  type: REFRESH_SOURCE_SUCCESS,
  addr,
});
const refreshSourceFail = (addr: RepositoryAddress) => ({
  type: REFRESH_SOURCE_FAIL,
  addr,
});

export const refreshSourceInfo = (
  addr: RepositoryAddress,
  repo_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(refreshSourceRequest(addr));
    return authRequest('GET', `/ingest/repositories/${repo_id}/update`)
      .then((ata) => dispatch(refreshSourceSuccess(addr)))
      .catch((e) => dispatch(refreshSourceFail(addr)));
  };
};

export const triggerProcess = (
  repo_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest('GET', `/ingest/repositories/${repo_id}/process/trigger`).then(
      checkStatus
    );
  //todo: handle the response
};

export const stopRepositoryProcess = (
  repo_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest('GET', `/ingest/repositories/${repo_id}/process/stop`).then(
      checkStatus
    );
};

export const FETCH_COMMITS_REQUEST = 'FETCH_COMMITS_REQUEST';
export const FETCH_COMMITS_SUCCESS = 'FETCH_COMMITS_SUCCESS';
export const FETCH_COMMITS_FAIL = 'FETCH_COMMITS_FAIL';

const fetchCommitsRequest = (addr: RepositoryAddress) => ({
  type: FETCH_COMMITS_REQUEST,
  addr,
});

//todo: should split in to actions (fetch commit and select branch)
const fetchCommitsSuccess = (
  addr: RepositoryAddress,
  commits: any,
  selected_branch?: string
) => ({
  type: FETCH_COMMITS_SUCCESS,
  addr,
  selected_branch,
  commits,
});
const fetchCommitsFail = (addr: RepositoryAddress, error: any) => ({
  type: FETCH_COMMITS_FAIL,
  addr,
  error,
});

export const fetchCommits = (
  addr: RepositoryAddress,
  branch?: string,
  page?: number
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchCommitsRequest(addr));

    const param = [];
    if (branch) param.push(`branch=${branch}`);
    if (page) param.push(`page=${page}`);
    const query = param.length > 0 ? '?' + param.join('&') : '';

    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/commits${query}`
    )
      .then(parsJsonLines)
      .then((data) => dispatch(fetchCommitsSuccess(addr, data, branch)))
      .catch((e) => dispatch(fetchCommitsFail(addr, e)));
  };
};

export const FETCH_REPOSITORY_REQUEST = 'FETCH_REPOSITORY_REQUEST';
export const FETCH_REPOSITORY_SUCCESS = 'FETCH_REPOSITORY_SUCCESS';
export const FETCH_REPOSITORY_FAIL = 'FETCH_REPOSITORY_FAIL';

const fetchRepositoryRequest = (addr: RepositoryAddress) => ({
  type: FETCH_REPOSITORY_REQUEST,
  addr,
});
const fetchRepositorySuccess = (
  addr: RepositoryAddress,
  repository: Repository
) => ({
  type: FETCH_REPOSITORY_SUCCESS,
  addr,
  repository,
});
const fetchRepositoryFail = (addr: RepositoryAddress, error: any) => ({
  type: FETCH_REPOSITORY_FAIL,
  addr,
  error,
});

export const fetchRepository = (
  addr: RepositoryAddress
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchRepositoryRequest(addr));
    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}`
    )
      .then((response) => response.json())
      .then((data) => dispatch(fetchRepositorySuccess(addr, data)))
      .catch((e) => dispatch(fetchRepositoryFail(addr, e)));
  };
};

export const FETCH_COMMIT_REQUEST = 'FETCH_COMMIT_REQUEST';
export const FETCH_COMMIT_SUCCESS = 'FETCH_COMMIT_SUCCESS';
export const FETCH_COMMIT_FAIL = 'FETCH_COMMIT_FAIL';

const fetchCommitRequest = () => ({
  type: FETCH_COMMIT_REQUEST,
});
const fetchCommitSuccess = (commit: Comment) => ({
  type: FETCH_COMMIT_SUCCESS,
  commit,
});
const fetchCommitFail = (error: any) => ({
  type: FETCH_COMMIT_FAIL,
  error,
});

export const fetchCommit = (
  addr: RepositoryAddress,
  hash: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchCommitRequest());
    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/commits/${hash}`
    )
      .then((response) => response.json())
      .then((data) => dispatch(fetchCommitSuccess(data)))
      .catch((e) => dispatch(fetchCommitFail(e)));
  };
};

export const FETCH_JOBS_REQUEST = 'FETCH_JOBS_REQUEST';
export const FETCH_JOBS_SUCCESS = 'FETCH_JOBS_SUCCESS';
export const FETCH_JOBS_FAIL = 'FETCH_JOBS_FAIL';

const fetchJobsRequest = (addr: RepositoryAddress) => ({
  type: FETCH_JOBS_REQUEST,
  addr,
});
const fetchJobsSuccess = (addr: RepositoryAddress, jobs: Job[]) => ({
  type: FETCH_JOBS_SUCCESS,
  addr,
  jobs,
});
const fetchJobsFail = (addr: RepositoryAddress, err: Error) => ({
  type: FETCH_JOBS_FAIL,
  addr,
  err,
});

export const fetchJobs = (
  addr: RepositoryAddress,
  repo_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchJobsRequest(addr));
    authRequest('GET', `/ingest/repositories/${repo_id}/jobs`)
      .then(parsJsonLines)
      .then((data) => dispatch(fetchJobsSuccess(addr, data)))
      .catch((err) => dispatch(fetchJobsFail(addr, err)));
  };
};

export const FETCH_MODELS_REQUEST = 'FETCH_MODELS_REQUEST';
export const FETCH_MODELS_SUCCESS = 'FETCH_MODELS_SUCCESS';
export const FETCH_MODELS_FAIL = 'FETCH_MODELS_FAIL';

const fetchModelsRequest = (addr: RepositoryAddress) => ({
  type: FETCH_MODELS_REQUEST,
  addr,
});
const fetchModelsSuccess = (addr: RepositoryAddress, models: Job[]) => ({
  type: FETCH_MODELS_SUCCESS,
  addr,
  models,
});
const fetchModelsFail = (addr: RepositoryAddress, err: Error) => ({
  type: FETCH_MODELS_FAIL,
  addr,
  err,
});

export const fetchModels = (
  addr: RepositoryAddress,
  repo_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    //todo: if do not request if the user do not have the permissions
    dispatch(fetchModelsRequest(addr));
    authRequest('GET', `/ai/repositories/${repo_id}/models`)
      .then(parsJsonLines)
      .then((data) => dispatch(fetchModelsSuccess(addr, data)))
      .catch((err) => dispatch(fetchModelsFail(addr, err)));
  };
};

export const FETCH_PULLS_REQUEST = 'FETCH_PULLS_REQUEST';
export const FETCH_PULLS_SUCCESS = 'FETCH_PULLS_SUCCESS';
export const FETCH_PULLS_FAIL = 'FETCH_PULLS_FAIL';

const fetchPullsRequest = (addr: RepositoryAddress) => ({
  type: FETCH_PULLS_REQUEST,
  addr,
});
const fetchPullsSuccess = (addr: RepositoryAddress, pulls: PullRequest[]) => ({
  type: FETCH_PULLS_SUCCESS,
  addr,
  pulls,
});
const fetchPullsFail = (addr: RepositoryAddress, err: Error) => ({
  type: FETCH_PULLS_FAIL,
  addr,
  err,
});

export const fetchPulls = (
  addr: RepositoryAddress
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchPullsRequest(addr));
    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/pulls`
    )
      .then(parsJsonLines)
      .then((data) => dispatch(fetchPullsSuccess(addr, data)))
      .catch((err) => dispatch(fetchPullsFail(addr, err)));
  };
};

export const FETCH_PULL_REQUEST = 'FETCH_PULL_REQUEST';
export const FETCH_PULL_SUCCESS = 'FETCH_PULL_SUCCESS';
export const FETCH_PULL_FAIL = 'FETCH_PULL_FAIL';

const fetchPullRequest = (addr: RepositoryAddress) => ({
  type: FETCH_PULL_REQUEST,
  addr,
});
const fetchPullSuccess = (pull: PullRequest) => ({
  type: FETCH_PULL_SUCCESS,
  pull,
});
const fetchPullFail = (err: Error) => ({
  type: FETCH_PULL_FAIL,
  err,
});

export const fetchPull = (
  addr: RepositoryAddress,
  number: number
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchPullRequest(addr));
    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/pulls/${number}`
    )
      .then((response) => response.json())
      .then((data) => dispatch(fetchPullSuccess(data)))
      .catch((err) => dispatch(fetchPullFail(err)));
  };
};

export const FETCH_PULL_COMMITS_REQUEST = 'FETCH_PULL_COMMITS_REQUEST';
export const FETCH_PULL_COMMITS_SUCCESS = 'FETCH_PULL_COMMITS_SUCCESS';
export const FETCH_PULL_COMMITS_FAIL = 'FETCH_PULL_COMMITS_FAIL';

const fetchPullCommitsRequest = () => ({
  type: FETCH_PULL_COMMITS_REQUEST,
});
const fetchPullCommitsSuccess = (commits: any) => ({
  type: FETCH_PULL_COMMITS_SUCCESS,
  commits,
});
const fetchPullCommitsFail = (error: any) => ({
  type: FETCH_PULL_COMMITS_FAIL,
  error,
});

export const fetchPullCommits = (
  addr: RepositoryAddress,
  pull: number,
  page?: number
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchPullCommitsRequest());

    const param = [];
    if (page) param.push(`page=${page}`);
    const query = param.length > 0 ? '?' + param.join('&') : '';

    authRequest(
      'GET',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/pulls/${pull}/commits${query}`
    )
      .then(parsJsonLines)
      .then((data) => dispatch(fetchPullCommitsSuccess(data)))
      .catch((e) => dispatch(fetchPullCommitsFail(e)));
  };
};

export const deleteCommitTag = (
  addr: RepositoryAddress,
  commit_hash: string,
  tag: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest(
      'DELETE',
      `/ingest/platforms/${addr.platform}/repos/${addr.owner}/${addr.repo}/commits/${commit_hash}/tags/${tag}`
    );
};
