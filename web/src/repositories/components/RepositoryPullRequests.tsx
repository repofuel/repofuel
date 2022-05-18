import React, {Suspense} from 'react';

import {RepositoryAddress} from '../types';

import {PullRequestsList} from './PullRequestsList';
import {RouteComponentProps, useParams} from 'react-router-dom';
import {Card, CardActionButton, CardActions} from '@rmwc/card';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Page404} from '../../ui/Page404';
import {RepositoryLayout} from './RepositoryLayout';
import {PageSpinner} from '../../ui/Layout';
import {RepositoryPullRequests_repository$key} from './__generated__/RepositoryPullRequests_repository.graphql';
import {RepositoryPullRequestsQuery} from './__generated__/RepositoryPullRequestsQuery.graphql';
import {usePaginationFragment} from 'react-relay/lib/hooks';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faSpinner} from '@fortawesome/free-solid-svg-icons';
import {useStatusTracker} from './Progress';
import {Helmet} from 'react-helmet';

const PULL_REQUEST_SCREEN_PAGE_SIZE = 30;

interface RepositoryPullRequestsProps {
  repository: RepositoryPullRequests_repository$key;
  repoAddr: RepositoryAddress; //fixme: should be removed
  pageSize: number;
}

export const RepositoryPullRequests: React.FC<RepositoryPullRequestsProps> = (
  props
) => {
  const {
    data: repository,
    hasNext,
    refetch,
    isLoadingNext,
    loadNext,
  } = usePaginationFragment(
    graphql`
      fragment RepositoryPullRequests_repository on Repository
      @refetchable(queryName: "RepositoryPullRequestsRefetchableQuery")
      @argumentDefinitions(
        pulls_count: {type: "Int"}
        pulls_cursor: {type: "String"}
      ) {
        status
        pullRequests(first: $pulls_count, after: $pulls_cursor)
          @connection(key: "RepositoryPullRequests_repository_pullRequests") {
          edges {
            node {
              id
              ...PullRequestsList_Item_pullRequest
            }
          }
        }
      }
    `,
    props.repository
  );

  function handelRefetch() {
    refetch({}, {fetchPolicy: 'store-and-network'});
  }

  useStatusTracker(repository.status, handelRefetch, 'READY', 'WATCHED');

  function handelClick() {
    if (isLoadingNext || !hasNext) return;

    loadNext(props.pageSize);
  }

  return (
    <Card outlined>
      <PullRequestsList
        repoAddr={props.repoAddr}
        pullRequests={repository.pullRequests?.edges}>
        {(repository.pullRequests?.edges?.length || 0) >= props.pageSize && (
          <CardActions fullBleed>
            <CardActionButton
              disabled={!hasNext || isLoadingNext}
              onClick={handelClick}
              label="Load more pull request"
              trailingIcon={
                isLoadingNext && <FontAwesomeIcon icon={faSpinner} spin />
              }
            />
          </CardActions>
        )}
      </PullRequestsList>
    </Card>
  );
};

interface RepositoryPullRequestsScreenProps extends RouteComponentProps {}

export const RepositoryPullRequestsScreen: React.FC<RepositoryPullRequestsScreenProps> = ({
  location,
}) => {
  const repoAddr: any = useParams();
  const {platform, owner, repo} = repoAddr;

  const {repository} = useLazyLoadQuery<RepositoryPullRequestsQuery>(
    graphql`
      query RepositoryPullRequestsQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $pulls_count: Int
      ) {
        repository(provider: $provider, owner: $owner, name: $name) {
          ...RepositoryLayout_repository
          ...RepositoryPullRequests_repository
            @arguments(pulls_count: $pulls_count)
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      pulls_count: PULL_REQUEST_SCREEN_PAGE_SIZE,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository) return <Page404 location={location} />;

  return (
    <RepositoryLayout repository={repository}>
      <Helmet>
        <title>Pull requests</title>
      </Helmet>
      <Suspense fallback={<PageSpinner />}>
        <RepositoryPullRequests
          repoAddr={repoAddr}
          repository={repository}
          pageSize={PULL_REQUEST_SCREEN_PAGE_SIZE}
        />
      </Suspense>
    </RepositoryLayout>
  );
};
