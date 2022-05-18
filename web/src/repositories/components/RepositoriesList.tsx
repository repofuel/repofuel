import React, {useMemo} from 'react';
import {
  ListItemMeta,
  ListItemPrimaryText,
  ListItemSecondaryText,
  ListItemText,
} from '@rmwc/list';
import {Link} from 'react-router-dom';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faCircle} from '@fortawesome/free-solid-svg-icons';

import './RepositoriesList.scss';
import {RiskyCommitsDoughnut} from './RiskyCommitsDoughnut';
import {Blankslate} from './CommitsList';
import {RepoIcon} from '@primer/octicons-react';
import Skeleton from 'react-loading-skeleton/lib';
import {Stage} from './__generated__/JobsListItem_job.graphql';
import graphql from 'babel-plugin-relay/macro';
import {useFragment} from 'react-relay/hooks';
import {RepositoriesList_Item_repository$key} from './__generated__/RepositoriesList_Item_repository.graphql';
import {RepositoriesList_repositories$key} from './__generated__/RepositoriesList_repositories.graphql';
import {useProgressSubscription} from './Progress';

interface RepositoryListProps {
  repositories?: RepositoriesList_repositories$key;
}

export const RepositoriesList: React.FC<RepositoryListProps> = (props) => {
  const repositories = useFragment(
    graphql`
      fragment RepositoriesList_repositories on RepositoryConnection {
        edges {
          node {
            id
            ...RepositoriesList_Item_repository
          }
        }
      }
    `,
    props.repositories || null
  );

  const ids = useMemo(
    () => repositories?.edges?.map((edge) => edge?.node?.id as string) || [],
    [repositories]
  );

  useProgressSubscription(ids);

  //todo: to be enabled when the the React concurrent mode arrives
  // if ( isFetching === true) return <SkeletonRepositoriesList />;

  if (!repositories?.edges?.length) {
    return (
      <Blankslate>
        <RepoIcon size={'large'} />
        <h3>There aren’t any repositories.</h3>
        <p>
          We will keep watching. When repositories get added, we will show them
          here.
        </p>
      </Blankslate>
    );
  }

  return (
    <div className="no-pointer divided-list mdc-list mdc-list--two-line">
      {repositories.edges.map(
        (edge) =>
          edge?.node && (
            <RepositoriesListItem key={edge.node.id} repository={edge.node} />
          )
      )}
    </div>
  );
};

export const SkeletonRepositoriesList: React.FC = () => {
  //todo: have a max rows as in the commits count skeleton
  return (
    <div className="mdc-list--non-interactive divided-list mdc-list mdc-list--two-line">
      {[...Array(10)].map((_, i) => (
        <div key={i} className="mdc-list-item">
          <span className="margin-right">
            <Skeleton circle={true} height={45} width={45} />
          </span>
          <ListItemText>
            <ListItemPrimaryText>
              <Skeleton width={200 + Math.random() * 300} />
            </ListItemPrimaryText>
            <ListItemSecondaryText>
              <Skeleton width={100 + Math.random() * 20} />
            </ListItemSecondaryText>
          </ListItemText>

          <ListItemMeta className="repo-status">
            <span>
              <Skeleton />
            </span>
          </ListItemMeta>
        </div>
      ))}
    </div>
  );
};

interface RepositoriesListItemProps {
  repository: RepositoriesList_Item_repository$key;
}

const RepositoriesListItem: React.FC<RepositoriesListItemProps> = (props) => {
  const repository = useFragment(
    graphql`
      fragment RepositoriesList_Item_repository on Repository {
        id
        name
        owner {
          slug
        }
        providerSCM
        status
        commitsCount
        buggyCommitsCount
        PredictionStatus
      }
    `,
    props.repository
  );

  const repoName = repository.owner.slug + '/' + repository.name;
  const repoHref = `/repos/${repository.providerSCM}/${repoName}`;

  let status = repository.status;
  if (
    repository.PredictionStatus &&
    repository.PredictionStatus > 0 &&
    status === 'READY'
  ) {
    status = 'WATCHED';
  }

  return (
    <div className="mdc-list-item">
      <RiskyCommitsDoughnut
        className="risk-chart"
        commitNum={repository.commitsCount}
        riskyCommitNum={repository.buggyCommitsCount}
      />

      <ListItemText>
        <ListItemPrimaryText>
          <Link className="link" to={repoHref}>
            {repoName}
          </Link>
        </ListItemPrimaryText>
        <ListItemSecondaryText>
          {repository.commitsCount ? repository.commitsCount + ' commits' : ''}
        </ListItemSecondaryText>
      </ListItemText>

      <ListItemMeta className="repo-status">
        <span>
          <JobStatus status={status} />
        </span>
      </ListItemMeta>
    </div>
  );
};

export const JobStatus: React.FC<{status: Stage}> = ({status}) => {
  let statusClassName;
  switch (status) {
    case 'READY':
      statusClassName = 'repo-status_ready';
      break;
    case 'FAILED':
      statusClassName = 'repo-status_error';
      break;
    case 'CANCELED':
    case 'RECOVERED':
    case 'WATCHED':
      statusClassName = 'repo-status_idle';
      break;
    default:
      statusClassName = 'repo-status_pending';
  }

  return (
    <>
      <FontAwesomeIcon
        className={statusClassName}
        icon={faCircle}
        style={{fontSize: '.7em'}}
      />{' '}
      {status.charAt(0) + status.slice(1).toLowerCase()}
    </>
  );
};
