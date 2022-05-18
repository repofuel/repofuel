import React from 'react';
import {RepositoryAddress} from '../types';
import {RouteComponentProps, useParams} from 'react-router-dom';
import {CommitsList, LoadMoreCardButton} from './CommitsList';
import {GridCell, GridRow} from '@rmwc/grid';
import {Card} from '@rmwc/card';
import {Typography} from '@rmwc/typography';
import 'github-markdown-css/github-markdown.css';
import ReactMarkdown from 'react-markdown';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Page404, Page404Custom} from '../../ui/Page404';
import {RepositoryLayout} from './RepositoryLayout';
import {PullRequestDetails_repository$key} from './__generated__/PullRequestDetails_repository.graphql';
import {usePaginationFragment} from 'react-relay/lib/hooks';
import {useProgressSubscription, useStatusTracker} from './Progress';
import {PullRequestDetailsQuery} from './__generated__/PullRequestDetailsQuery.graphql';
import {JobStatus} from './RepositoriesList';
import {PullRequestIcon} from './PullRequestsList';

const PULL_REQUEST_COMMITS_PAGE_SIZE = 20;

interface PullRequestDetailsProps {
  repoAddr: RepositoryAddress;
  repository: PullRequestDetails_repository$key;
  pageSize: number;
}

export const PullRequestDetails: React.FC<PullRequestDetailsProps> = (
  props
) => {
  const {
    data: repository,
    refetch,
    hasNext,
    isLoadingNext,
    loadNext,
  } = usePaginationFragment(
    graphql`
      fragment PullRequestDetails_repository on Repository
      @refetchable(queryName: "PullRequestDetailsRefreshQuery")
      @argumentDefinitions(
        commits_count: {type: "Int"}
        commits_cursor: {type: "String"}
        pull_number: {type: "Int!"}
      ) {
        source {
          url
        }
        providerSCM
        pullRequest(number: $pull_number) {
          id
          source {
            number
            title
            body
            merged
            closed
          }
          status
          commits(first: $commits_count, after: $commits_cursor)
            @connection(key: "PullRequestDetails_pullRequest_commits") {
            edges {
              node {
                id
                ...CommitsList_Item_commit
              }
            }
          }
        }
      }
    `,
    props.repository
  );

  const pull = repository.pullRequest;

  function handelRefetch() {
    refetch({}, {fetchPolicy: 'store-and-network'});
  }

  useProgressSubscription([pull?.id || '']);
  useStatusTracker(
    pull?.status,
    handelRefetch,
    'PREDICTING',
    'READY',
    'WATCHED'
  );

  if (!pull) return <Page404Custom>Cannot find the pull request</Page404Custom>;

  return (
    <GridRow>
      <GridCell span={12}>
        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              <PullRequestIcon size={'medium'} pull={pull.source} /> #
              {pull.source.number} {pull.source.title}
              <span className="float-right mdc-list-item__secondary-text">
                <JobStatus status={pull.status} />
              </span>
            </Typography>
          </div>
          {pull.source.body && (
            <div className="markdown-body" style={{margin: '30px'}}>
              <ReactMarkdown source={pull.source.body} />
            </div>
          )}
        </Card>
      </GridCell>
      <GridCell span={12}>
        <Card outlined>
          <CommitsList
            repoAddr={props.repoAddr}
            repoURL={repository.source.url}
            platform={repository.providerSCM}
            commits={pull.commits?.edges}>
            <LoadMoreCardButton
              label="Load more commits"
              pageSize={props.pageSize}
              itemsCount={pull.commits?.edges?.length || 0}
              isLoading={isLoadingNext}
              hasMore={hasNext}
              loadMore={loadNext}
            />
          </CommitsList>
        </Card>
      </GridCell>
    </GridRow>
  );
};

interface PullRequestDetailsScreenProps extends RouteComponentProps {}

export const PullRequestDetailsScreen: React.FC<PullRequestDetailsScreenProps> = ({
  location,
}) => {
  //fixme: we should not pass down the repoAddr, the data should be gotten from the query result
  const repoAddr: any = useParams();
  const {platform, owner, repo, number} = repoAddr;

  const {repository} = useLazyLoadQuery<PullRequestDetailsQuery>(
    graphql`
      query PullRequestDetailsQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $pull_number: Int!
        $commits_count: Int
      ) {
        repository(provider: $provider, owner: $owner, name: $name) {
          ...RepositoryLayout_repository
          ...PullRequestDetails_repository
            @arguments(commits_count: $commits_count, pull_number: $pull_number)
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      pull_number: number,
      commits_count: PULL_REQUEST_COMMITS_PAGE_SIZE,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository) return <Page404 location={location} />;

  return (
    <RepositoryLayout repository={repository}>
      <PullRequestDetails
        repoAddr={repoAddr}
        repository={repository}
        pageSize={PULL_REQUEST_COMMITS_PAGE_SIZE}
      />
    </RepositoryLayout>
  );
};
