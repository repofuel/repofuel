import tokenStore from '../account/token';
import {SubscriptionClient} from 'subscriptions-transport-ws';

export async function fetchGraphQL(text: string, variables: Object) {
  const access_token = await tokenStore.getAccessToken();

  const response = await fetch('/ingest/graphql', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${access_token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: text,
      variables,
    }),
  });

  return await response.json();
}

export const subscriptionClient = new SubscriptionClient(
  `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${
    window.location.host
  }/ingest/graphql`,
  {
    reconnect: true,
    connectionParams: async () => {
      return {
        Authorization: await tokenStore.getAccessToken(),
      };
    },
  }
);
