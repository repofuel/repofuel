import tokenStore from '../account/token';
import {checkStatus} from '../util/requests';

// todo: refactor to match the seigneur of fetch
export function authRequest(method: string, url: string, init?: RequestInit) {
  return tokenStore
    .getAccessToken()
    .then((access_token) => {
      return fetch(url, {
        ...init,
        method: method,
        headers: {
          ...init?.headers,
          Authorization: `Bearer ${access_token}`,
        },
      });
    })
    .then(checkStatus);
}
