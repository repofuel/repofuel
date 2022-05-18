import React, {useMemo} from 'react';
import {
  ListItemGraphic,
  ListItemMeta,
  ListItemPrimaryText,
  ListItemSecondaryText,
  ListItemText,
} from '@rmwc/list';
import {
  GitMergeIcon,
  GitPullRequestIcon,
  IconProps,
} from '@primer/octicons-react';
import {format} from 'date-fns';
import './PullRequest.scss';
import {Tooltip} from '@rmwc/tooltip';
import {Blankslate} from './CommitsList';
import {JobStatus} from './RepositoriesList';
import {RepositoryPullRequests_repository} from './__generated__/RepositoryPullRequests_repository.graphql';
import {useFragment} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {PullRequestsList_Item_pullRequest$key} from './__generated__/PullRequestsList_Item_pullRequest.graphql';
import {Link} from 'react-router-dom';
import {RepositoryAddress} from '../types';
import {useProgressSubscription} from './Progress';

interface PullRequestsListProps {
  pullRequests: RepositoryPullRequests_repository['pullRequests']['edges'];
  repoAddr: RepositoryAddress;
}

export const PullRequestsList: React.FC<PullRequestsListProps> = (props) => {
  const ids = useMemo(
    () => props.pullRequests?.map((edge) => edge?.node?.id || '') || [],
    [props.pullRequests]
  );

  useProgressSubscription(ids);

  if (!props?.pullRequests?.length)
    return (
      <Blankslate>
        <GitPullRequestIcon size={'large'} />
        <h3>There aren’t any pull requests.</h3>
        <p>
          We will keep watching. When new pull requests get analyzed, we will
          show them here.
        </p>
      </Blankslate>
    );

  return (
    <div className="no-pointer divided-list mdc-list mdc-list--two-line">
      {props.pullRequests?.map(
        (edge) =>
          edge &&
          edge.node && (
            <PullRequestsListItem
              key={edge.node.id}
              pull={edge.node}
              repoAddr={props.repoAddr}
            />
          )
      )}

      {props.children}
    </div>
  );
};

interface PullRequestsListItemProps {
  pull: PullRequestsList_Item_pullRequest$key;
  repoAddr: RepositoryAddress;
}

const PullRequestsListItem: React.FC<PullRequestsListItemProps> = (props) => {
  const pull = useFragment(
    graphql`
      fragment PullRequestsList_Item_pullRequest on PullRequest {
        id
        status
        source {
          number
          title
          createdAt
          merged
          mergedAt
          closed
          closedAt
        }
      }
    `,
    props.pull
  );

  return (
    <div className="mdc-list-item">
      <ListItemGraphic icon={<PullRequestIcon pull={pull.source} />} />

      <ListItemText>
        <ListItemPrimaryText
          tag={Link}
          className="link"
          to={`/repos/${props.repoAddr.platform}/${props.repoAddr.owner}/${props.repoAddr.repo}/pulls/${pull.source.number}`}>
          {pull.source.title}
        </ListItemPrimaryText>
        <ListItemSecondaryText>
          #{pull.source.number} opened on{' '}
          {format(new Date(pull.source.createdAt), 'MMM d, y')}
        </ListItemSecondaryText>
      </ListItemText>

      <ListItemMeta className="repo-status">
        <span>
          <JobStatus status={pull.status} />
        </span>
      </ListItemMeta>
    </div>
  );
};

interface PullRequestIconProps extends IconProps {
  pull: {
    merged: boolean;
    closed: boolean;
  };
}

export const PullRequestIcon: React.FC<PullRequestIconProps> = ({
  pull,
  ...props
}) => {
  if (pull.merged)
    return (
      <Tooltip content="Merge pull request">
        <GitMergeIcon {...props} className="pull merged" />
      </Tooltip>
    );

  if (pull.closed)
    return (
      <Tooltip content="Closed pull request">
        <GitPullRequestIcon {...props} className="pull closed" />
      </Tooltip>
    );

  return (
    <Tooltip content="Open pull request">
      <GitPullRequestIcon {...props} className="pull open" />
    </Tooltip>
  );
};
