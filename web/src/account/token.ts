import {MouseEvent} from 'react';
import {Credentials} from './types';
import {useEffect, useState} from 'react';

const expiryDelta = 10000;

interface TokenResponse {
  expires_in: number;
  access_token: string;
  refresh_token?: string;
  token_type: string;
}

class TokenStore {
  private refreshRequestInfo?: RequestInfo;
  private accessTokenPromise?: Promise<string>;
  private tokenType: string;
  private expiry: number;
  private isFetching: boolean;
  private subscribers: Set<(isAuthenticated: boolean) => void>;

  constructor() {
    this.expiry = 0;
    this.isFetching = false;
    this.tokenType = 'bearer';
    this.subscribers = new Set();

    const refreshToken = localStorage.getItem('refresh_token');
    if (refreshToken) {
      this.refreshRequestInfo = this.createRefreshRequestInfo(refreshToken);
      this.accessTokenPromise = this.getAccessToken();
    }
    // commitUserAccountLocally(RelayEnvironment, {isAuthenticated: !!refreshToken})
  }

  isAuthenticated(): boolean {
    return !!this.refreshRequestInfo;
  }

  getExpiryDate() {
    return new Date(this.expiry);
  }

  getAccessToken(): Promise<string> {
    if (
      !this.accessTokenPromise ||
      (!this.hasValidAccessToken() && !this.isFetching)
    ) {
      this.accessTokenPromise = this.newAccessToken();
    }

    return this.accessTokenPromise;
  }

  subscribe(sub: (isAuthenticated: boolean) => void) {
    this.subscribers.add(sub);

    return () => {
      this.subscribers.delete(sub);
    };
  }

  login(creds: Credentials) {
    let url = authorizationUrl(creds);
    if (!url) return Promise.reject('unsupported credentials grunt type');

    this.isFetching = true;
    this.accessTokenPromise = fetch(url)
      .then((response): any => {
        if (!response.ok) return Promise.reject(response);

        return response.json();
      })
      .then((token: TokenResponse) => {
        this.applyTokenResponse(token);
        return token.access_token;
      })
      .catch((err) => {
        this.isFetching = false;
        return Promise.reject(err);
      });

    return this.accessTokenPromise.then(() => this.notify(true));
  }

  logout() {
    localStorage.removeItem('refresh_token');
    this.refreshRequestInfo = undefined;
    this.accessTokenPromise = undefined;
    this.isFetching = false;
    this.expiry = 0;
    this.notify(false);
  }

  private hasValidAccessToken(): boolean {
    return this.expiry - expiryDelta > Date.now();
  }

  private newAccessToken(): Promise<string> {
    if (!this.refreshRequestInfo) return Promise.reject('need to login first');

    this.isFetching = true;
    return fetch(this.refreshRequestInfo)
      .then((response) => {
        if (!response.ok) return Promise.reject(response);

        return response.json();
      })
      .then((token: TokenResponse) => {
        this.applyTokenResponse(token);
        return token.access_token;
      })
      .catch((err) => {
        this.isFetching = false;
        if (err.status === 401) {
          tokenStore.logout();
        }

        return Promise.reject(err);
      });
  }

  private createRefreshRequestInfo(refreshToken: string) {
    return new Request('/accounts/auth/access_token', {
      method: 'POST',
      headers: new Headers({
        'Content-Type': 'application/x-www-form-urlencoded',
      }),
      body: new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
      }),
    });
  }

  private applyTokenResponse(token: TokenResponse) {
    this.expiry = Date.now() + token.expires_in * 1000;

    if (token.refresh_token) {
      this.refreshRequestInfo = this.createRefreshRequestInfo(
        token.refresh_token
      );
      localStorage.setItem('refresh_token', token.refresh_token);
    }

    if (token.token_type) {
      this.tokenType = token.token_type;
    }

    this.isFetching = false;
  }

  private notify(isAuthenticated: boolean) {
    this.subscribers.forEach((sub) => sub(isAuthenticated));
  }
}

function authorizationUrl(creds: Credentials) {
  switch (creds.grant_type) {
    case 'authorization_code':
      return `/accounts/login/${creds.provider}/callback?code=${creds.code}&state=${creds.state}`;
    case 'oauth_token':
      return `/accounts/login/${creds.provider}/callback?oauth_token=${creds.oauth_token}&oauth_verifier=${creds.oauth_verifier}`;
  }
}

const tokenStore = new TokenStore();
export default tokenStore;

export const useIsAuthenticated = (): boolean => {
  const [isAuthenticated, setAuthenticated] = useState(
    tokenStore.isAuthenticated()
  );
  useEffect(() => tokenStore.subscribe(setAuthenticated), [setAuthenticated]);
  return isAuthenticated;
};

export const authenticateDownload = (url: string) => {
  return (e: MouseEvent) => {
    tokenStore.getAccessToken().then((access_token) => {
      document.cookie = `jwt=${access_token}; expires=${tokenStore
        .getExpiryDate()
        .toUTCString()}; path=/ingest/download`;
      // eslint-disable-next-line no-restricted-globals
      location.href = url;
    });
  };
};
