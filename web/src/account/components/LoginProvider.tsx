import React, {useEffect, useState} from 'react';
import {Redirect} from 'react-router';
import {LoginScreen} from './LoginScreen';
import {useParams} from 'react-router-dom';
import {Credentials} from '../types';
import {notify} from '../../ui/snackbar';
import tokenStore, {useIsAuthenticated} from '../token';

const LoginProvider: React.FC<any> = (props) => {
  const isAuthenticated = useIsAuthenticated();
  const [isFetching, setFetching] = useState(true);
  const {provider}: any = useParams();
  useEffect(() => {
    const searchParams = new URL(window.location.href).searchParams;
    const cred = credFromSearchParams(searchParams, provider);
    if (!cred) {
      //todo: raise an error message
      return;
    }

    tokenStore.login(cred).catch((err) => {
      setFetching(false);
      notify({
        title: 'Error in the login process',
        actions: [
          {
            title: 'Dismiss',
          },
        ],
      });
    });
  }, [provider]);

  if (isAuthenticated) {
    return <Redirect to={{pathname: '/'}} />;
  }

  return <LoginScreen isFetching={isFetching} />;
};

function credFromSearchParams(
  params: URLSearchParams,
  provider: string
): Credentials | void {
  if (params.get('code')) {
    return {
      grant_type: 'authorization_code',
      provider: provider,
      code: params.get('code') || '',
      state: params.get('state') || '',
    };
  }

  if (params.get('oauth_token')) {
    return {
      grant_type: 'oauth_token',
      provider: provider,
      oauth_token: params.get('oauth_token') || '',
      oauth_verifier: params.get('oauth_verifier') || '',
    };
  }
}

export default LoginProvider;
