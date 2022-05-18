import {ThunkAction} from 'redux-thunk';
import {Organization} from './types';
import {AnyAction} from 'redux';
import {authRequest} from '../store/util';
import {parsJsonLines} from '../util/requests';
import {AppState} from '../store/types';

export const FETCH_ORGANIZATIONS_REQUEST = 'FETCH_ORGANIZATIONS_REQUEST';
export const FETCH_ORGANIZATIONS_SUCCESS = 'FETCH_ORGANIZATIONS_SUCCESS';
export const FETCH_ORGANIZATIONS_FAIL = 'FETCH_ORGANIZATIONS_FAIL';

const fetchOrganizationsRequest = () => ({
  type: FETCH_ORGANIZATIONS_REQUEST,
});
const fetchOrganizationsSuccess = (orgs: Organization[]) => ({
  type: FETCH_ORGANIZATIONS_SUCCESS,
  orgs,
});
const fetchOrganizationsFail = (error: any) => ({
  type: FETCH_ORGANIZATIONS_FAIL,
  error,
});

export const fetchMyOrganizations = (): ThunkAction<
  any,
  AppState,
  any,
  AnyAction
> => {
  return (dispatch, getState) => {
    dispatch(fetchOrganizationsRequest());
    authRequest('GET', '/ingest/user/orgs')
      .then(parsJsonLines)
      .then((data) => dispatch(fetchOrganizationsSuccess(data)))
      .catch((e) => dispatch(fetchOrganizationsFail(e)));
  };
};

export const FETCH_ORGANIZATION_REQUEST = 'FETCH_ORGANIZATION_REQUEST';
export const FETCH_ORGANIZATION_SUCCESS = 'FETCH_ORGANIZATION_SUCCESS';
export const FETCH_ORGANIZATION_FAIL = 'FETCH_ORGANIZATION_FAIL';

const fetchOrganizationRequest = () => ({
  type: FETCH_ORGANIZATION_REQUEST,
});
const fetchOrganizationSuccess = (org: Organization) => ({
  type: FETCH_ORGANIZATION_SUCCESS,
  org,
});
const fetchOrganizationFail = (error: any) => ({
  type: FETCH_ORGANIZATION_FAIL,
  error,
});

export const fetchOrganization = (
  provider?: string,
  slug?: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) => {
    dispatch(fetchOrganizationRequest());
    return authRequest('GET', `/ingest/platforms/${provider}/orgs/${slug}`)
      .then((response) => response.json())
      .then((data) => dispatch(fetchOrganizationSuccess(data)))
      .catch((e) => dispatch(fetchOrganizationFail(e)));
  };
};

export const CheckJiraUrl = (
  url: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest('POST', `/ingest/integrations/jira/check_url`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        base_url: url,
      }),
    }).then((response) => response.json());
};

export const GetOAuthUrl = (
  provider: string,
  org_id: string
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest(
      'GET',
      `/ingest/apps/${provider}/organizations/${org_id}/link`
    ).then((response) => response.json());
};

export const SendConditionals = (
  provider: string,
  org_id: string,
  form: object
): ThunkAction<any, AppState, any, AnyAction> => {
  return (dispatch, getState) =>
    authRequest(
      'POST',
      `/ingest/apps/${provider}/organizations/${org_id}/basic`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(form),
      }
    );
};
