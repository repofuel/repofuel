import {Reducer} from 'redux';
import {
  FETCH_ORGANIZATION_SUCCESS,
  FETCH_ORGANIZATIONS_FAIL,
  FETCH_ORGANIZATIONS_REQUEST,
  FETCH_ORGANIZATIONS_SUCCESS,
} from './actions';
import {OrganizationsState} from './types';

export const orgs: Reducer<OrganizationsState> = (
  state = {orgsList: {}},
  action
) => {
  switch (action.type) {
    case FETCH_ORGANIZATIONS_REQUEST:
      return {
        ...state,
        isFetching: true,
      };

    case FETCH_ORGANIZATIONS_FAIL:
      return {
        ...state,
        isFetching: false,
      };

    case FETCH_ORGANIZATIONS_SUCCESS:
      return {
        ...state,
        isFetching: false,
        orgsList: action.orgs,
      };
    case FETCH_ORGANIZATION_SUCCESS:
      return {
        ...state,
        current: action.org,
      };

    default:
      return state;
  }
};
