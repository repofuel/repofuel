import {Redirect, Route} from 'react-router';
import React from 'react';
import {useIsAuthenticated} from '../token';

const PrivateRoute: React.FC<any> = (props) => {
  const isAuthenticated = useIsAuthenticated();

  if (isAuthenticated) {
    return <Route {...props} />;
  }
  return <Redirect to={{pathname: '/login'}} />;
};

export default PrivateRoute;
