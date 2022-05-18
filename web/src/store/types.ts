import {
  CommitState,
  PullRequestState,
  RepositoriesState,
} from '../repositories/types';
import {OrganizationsState} from '../organization/types';
import {ThunkAction} from 'redux-thunk';
import {AnyAction} from 'redux';

export interface AppState {
  repositories: RepositoriesState;
  commit: CommitState;
  pull: PullRequestState;
  orgs: OrganizationsState;
}

//todo: should be used in all thunk actions
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  AppState,
  unknown,
  AnyAction
>;
