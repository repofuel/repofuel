import {
  Environment,
  FetchFunction,
  Network,
  Observable,
  RecordSource,
  Store,
  SubscribeFunction,
} from 'relay-runtime';
import {fetchGraphQL, subscriptionClient} from './fetchGraphQL';

const fetchRelay: FetchFunction = async (params, variables) => {
  if (!params.text) {
    //todo: support persisted queries
    throw Error('fetching persisted queries is not supported for now');
  }

  return fetchGraphQL(params.text, variables);
};

const subscribeRelay: SubscribeFunction = (request, variables) => {
  const subscribeObservable = subscriptionClient.request({
    query: request.text || undefined,
    operationName: request.name,
    variables,
  });

  // Convert subscriptions-transport-ws observable type to Relay's
  return Observable.from<any>(subscribeObservable); //fixme: should avoid using any
};

export default new Environment({
  network: Network.create(fetchRelay, subscribeRelay),
  store: new Store(new RecordSource(), {
    // This property tells Relay to not immediately clear its cache when the user
    // navigates around the app. Relay will hold onto the specified number of
    // query results, allowing the user to return to recently visited pages
    // and reusing cached data if its available/fresh.
    gcReleaseBufferSize: 10,
  }),
  // log: console.log,
});
