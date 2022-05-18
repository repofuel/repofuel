import {applyMiddleware, createStore} from 'redux';
import thunkMiddleware from 'redux-thunk';
import {AppState} from './types';
import {rootReducer} from './reducers';

//fixme: remove any's for when we finish testing, or specify the correct types
export const store = createStore<AppState, any, any, any>(
  rootReducer,
  applyMiddleware(thunkMiddleware)
);
