import {Reducer} from 'redux';
import {commit, pull, repositories} from '../repositories/reducers';
import {orgs} from '../organization/reducers';

export const rootReducer: Reducer = (state = {}, action) => {
  return {
    repositories: repositories(state.repositories, action),
    commit: commit(state.commit, action),
    pull: pull(state.pull, action),
    orgs: orgs(state.orgs, action),
  };
};
