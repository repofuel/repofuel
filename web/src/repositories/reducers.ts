import {Reducer} from 'redux';
import {
  FETCH_COMMIT_FAIL,
  FETCH_COMMIT_REQUEST,
  FETCH_COMMIT_SUCCESS,
  FETCH_COMMITS_FAIL,
  FETCH_COMMITS_REQUEST,
  FETCH_COMMITS_SUCCESS,
  FETCH_JOBS_FAIL,
  FETCH_JOBS_REQUEST,
  FETCH_JOBS_SUCCESS,
  FETCH_MODELS_FAIL,
  FETCH_MODELS_REQUEST,
  FETCH_MODELS_SUCCESS,
  FETCH_PULL_COMMITS_FAIL,
  FETCH_PULL_COMMITS_REQUEST,
  FETCH_PULL_COMMITS_SUCCESS,
  FETCH_PULL_FAIL,
  FETCH_PULL_REQUEST,
  FETCH_PULL_SUCCESS,
  FETCH_PULLS_FAIL,
  FETCH_PULLS_REQUEST,
  FETCH_PULLS_SUCCESS,
  FETCH_REPOSITORIES_FAIL,
  FETCH_REPOSITORIES_REQUEST,
  FETCH_REPOSITORIES_SUCCESS,
  FETCH_REPOSITORY_FAIL,
  FETCH_REPOSITORY_REQUEST,
  FETCH_REPOSITORY_SUCCESS,
  REFRESH_SOURCE_FAIL,
  REFRESH_SOURCE_REQUEST,
} from './actions';
import {
  CommitState,
  PullRequestState,
  RepositoriesMap,
  RepositoriesState,
  Repository,
  RepositoryAddress,
} from './types';

export function toStrAddr(addr: RepositoryAddress) {
  return `${addr.platform}/${addr.owner}/${addr.repo}`;
}

function updateRepository(
  state: RepositoriesState,
  addr: RepositoryAddress,
  repository: any
) {
  const strAddr = toStrAddr(addr);

  return {
    ...state,
    reposList: {
      ...state.reposList,
      [strAddr]: {
        ...state.reposList[strAddr],
        ...repository,
      },
    },
  };
}

function toRepositoriesMap(repos: Repository[]) {
  const map: RepositoriesMap = {};
  for (let i = 0, len = repos.length; i < len; i++) {
    const repo = repos[i];
    repo.listed = true;
    map[`${repo.provider}/${repo.owner}/${repo.name}`] = repo;
  }
  return map;
}

export const repositories: Reducer<RepositoriesState> = (
  state = {reposList: {}},
  action
) => {
  switch (action.type) {
    case FETCH_REPOSITORIES_REQUEST:
      return {
        ...state,
        isFetching: true,
      };
    case FETCH_REPOSITORIES_SUCCESS:
      // this will delete the changes e.g., loaded commits from the repository objects
      return {
        isFetching: false,
        reposList: toRepositoriesMap(action.repositories),
      };
    case FETCH_REPOSITORIES_FAIL:
      return {
        ...state,
        isFetching: false,
      };

    case FETCH_REPOSITORY_REQUEST:
    case REFRESH_SOURCE_REQUEST:
      return updateRepository(state, action.addr, {isFetching: true});

    case FETCH_REPOSITORY_SUCCESS:
      action.repository.isFetching = false;
      return updateRepository(state, action.addr, action.repository);

    case REFRESH_SOURCE_FAIL:
    case FETCH_REPOSITORY_FAIL:
      return updateRepository(state, action.addr, {
        error: action.error,
        isFetching: false,
      });

    case FETCH_COMMITS_REQUEST:
      return updateRepository(state, action.addr, {isCommitsLoading: true});
    case FETCH_COMMITS_SUCCESS:
      return updateRepository(state, action.addr, {
        commits: action.commits,
        selected_branch: action.selected_branch,
        isCommitsLoading: false,
      });
    case FETCH_COMMITS_FAIL:
      return updateRepository(state, action.addr, {isCommitsLoading: false});

    case FETCH_JOBS_REQUEST:
      return updateRepository(state, action.addr, {isJobsLoading: true});
    case FETCH_JOBS_SUCCESS:
      return updateRepository(state, action.addr, {
        jobs: action.jobs,
        isJobsLoading: false,
      });
    case FETCH_JOBS_FAIL:
      return updateRepository(state, action.addr, {isJobsLoading: false});

    case FETCH_MODELS_REQUEST:
      return updateRepository(state, action.addr, {isModelsLoading: true});
    case FETCH_MODELS_SUCCESS:
      return updateRepository(state, action.addr, {
        models: action.models,
        isModelsLoading: false,
      });
    case FETCH_MODELS_FAIL:
      return updateRepository(state, action.addr, {isModelsLoading: false});

    case FETCH_PULLS_REQUEST:
      return updateRepository(state, action.addr, {isPullsLoading: true});
    case FETCH_PULLS_SUCCESS:
      return updateRepository(state, action.addr, {
        pulls: action.pulls,
        isPullsLoading: false,
      });
    case FETCH_PULLS_FAIL:
      return updateRepository(state, action.addr, {isPullsLoading: false});

    default:
      return state;
  }
};

export const commit: Reducer<CommitState> = (state = {}, action) => {
  switch (action.type) {
    case FETCH_COMMIT_REQUEST:
      return {
        ...state,
        isFetching: true,
      };
    case FETCH_COMMIT_FAIL:
      return {
        ...state,
        isFetching: false,
        //todo: specify the text based on the error
        error: 'error',
      };
    case FETCH_COMMIT_SUCCESS:
      return {
        ...state,
        isFetching: false,
        commit: action.commit,
      };
    default:
      return state;
  }
};

export const pull: Reducer<PullRequestState> = (state = {}, action) => {
  switch (action.type) {
    case FETCH_PULL_REQUEST:
      return {
        ...state,
        isFetching: true,
      };
    case FETCH_PULL_FAIL:
      return {
        ...state,
        isFetching: false,
        //todo: specify the text based on the error
        error: 'error',
      };
    case FETCH_PULL_SUCCESS:
      return {
        ...state,
        isFetching: false,
        pull: action.pull,
      };

    case FETCH_PULL_COMMITS_REQUEST:
      return {
        ...state,
        isCommitsFetching: true,
      };
    case FETCH_PULL_COMMITS_FAIL:
      return {
        ...state,
        isCommitsFetching: false,
        //todo: specify the text based on the error
        error: 'error',
      };
    case FETCH_PULL_COMMITS_SUCCESS:
      return {
        ...state,
        isCommitsFetching: false,
        commits: action.commits,
      };

    default:
      return state;
  }
};
