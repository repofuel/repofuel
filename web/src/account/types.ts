export interface AuthorizationCodeCredentials {
  grant_type: 'authorization_code';
  provider: string;
  code: string;
  state: string;
}

export interface TokenCredentials {
  grant_type: 'access_code' | 'refresh_token';
  token: string;
}

export interface OauthTokenCredentials {
  grant_type: 'oauth_token';
  provider: string;
  oauth_token: string;
  oauth_verifier: string;
}

export type Credentials =
  | AuthorizationCodeCredentials
  | TokenCredentials
  | OauthTokenCredentials;
