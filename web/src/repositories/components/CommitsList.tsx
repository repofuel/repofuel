import React, {useState} from 'react';

import {formatDistanceToNow} from 'date-fns';

import {
  ListItemGraphic,
  ListItemMeta,
  ListItemPrimaryText,
  ListItemSecondaryText,
  ListItemText,
} from '@rmwc/list';
import {Avatar} from '@rmwc/avatar';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faSpinner} from '@fortawesome/free-solid-svg-icons';
import './CommitsList.scss';
import {GitCommitIcon} from '@primer/octicons-react';
import {Label, LabelProps} from '@primer/components';
import Skeleton from 'react-loading-skeleton/lib';
import graphql from 'babel-plugin-relay/macro';
import {commitAddress, CommitDetailsModal} from './CommitDetails';
import {useFragment} from 'react-relay/hooks';
import {CommitsList_Item_commit$key} from './__generated__/CommitsList_Item_commit.graphql';
import {CommitsList_SlimItem_commit$key} from './__generated__/CommitsList_SlimItem_commit.graphql';
import {CardActionButton, CardActions} from '@rmwc/card';
import {LoadMoreFn} from 'react-relay/lib/relay-experimental/useLoadMoreFunction';
import {RepositoryCommits_Filtered_repository} from './__generated__/RepositoryCommits_Filtered_repository.graphql';
import {Link} from 'react-router-dom';
import {RepositoryAddress} from '../types';
import ScorePoints from './ScorePoints';

interface BlankslateProps {}

export const Blankslate: React.FC<BlankslateProps> = ({children}) => {
  return <div className="blankslate">{children}</div>;
};

export const SkeletonCommitsList: React.FC<{max?: number}> = ({max}) => {
  return (
    <div className="mdc-list--non-interactive divided-list mdc-list mdc-list--two-line mdc-list--avatar-list">
      {[...Array(!max || max > 20 ? 20 : max)].map((_, i) => (
        <div key={i} className="mdc-list-item">
          <ListItemGraphic
            icon={<Skeleton circle={true} height={40} width={40} />}
          />

          <ListItemText>
            <ListItemPrimaryText>
              <Skeleton width={160 + Math.random() * 400} />
            </ListItemPrimaryText>
            <ListItemSecondaryText>
              <strong>
                <Skeleton width={70 + Math.random() * 90} />{' '}
              </strong>
              <Skeleton width={100 + Math.random() * 40} />
            </ListItemSecondaryText>
          </ListItemText>

          <ListItemMeta className="commit-meta">
            <Skeleton width={80} /> <Skeleton width={45} />
          </ListItemMeta>
        </div>
      ))}
    </div>
  );
};

//todo: should be improved
function CommitURL(provider: string, repoURL: string, hash: string): string {
  return `${repoURL}/${provider === 'github' ? 'commit' : 'commits'}/${hash}`;
}

interface CommitsListItemProps {
  onSelect: (id: string) => void;
  repoURL: string;
  platform: any; //todo: split the provider from the platform
  commit: CommitsList_Item_commit$key;
}

export const CommitsListItem: React.FC<CommitsListItemProps> = ({
  onSelect,
  repoURL,
  platform,
  ...props
}) => {
  const commit = useFragment(
    graphql`
      fragment CommitsList_Item_commit on Commit {
        id
        hash
        message
        author {
          name
          date
        }
        analysis {
          bugPotential
        }
        fixed
      }
    `,
    props.commit
  );

  return (
    <div className="mdc-list-item">
      <ListItemGraphic
        icon={
          <Avatar
            name={commit.author.name}
            // src={commit.author.avatar} //todo: to be included
          />
        }
      />

      <ListItemText>
        <ListItemPrimaryText>
          {/*todo: should be improved ti use href*/}
          <span className={'link'} onClick={() => onSelect(commit.id)}>
            {commit.message}
          </span>
          {commit.fixed && (
            <Label m={1} dropshadow bg="green.5">
              Fixed
            </Label>
          )}
        </ListItemPrimaryText>
        <ListItemSecondaryText>
          <strong>{commit.author.name} </strong>
          committed{' '}
          {formatDistanceToNow(new Date(commit.author.date), {addSuffix: true})}
        </ListItemSecondaryText>
      </ListItemText>

      <ListItemMeta className="commit-meta">
        <CommitsHashLabel
          mr={2}
          hash={commit.hash}
          repoURL={repoURL}
          platform={platform}
        />
        <RiskPoints risk={commit.analysis?.bugPotential} />
      </ListItemMeta>
    </div>
  );
};

interface CommitsListItemSlimProps {
  repoURL: string;
  platform: any; //todo: split the provider from the platform
  commit: CommitsList_SlimItem_commit$key;
  repoAddr: RepositoryAddress;
}

export const CommitsListSlimItem: React.FC<CommitsListItemSlimProps> = ({
  repoURL,
  platform,
  ...props
}) => {
  const commit = useFragment(
    graphql`
      fragment CommitsList_SlimItem_commit on Commit {
        id
        hash
        message
        author {
          name
          date
        }
      }
    `,
    props.commit
  );

  return (
    <div className="list-item commit-overview-list-item">
      <CommitsHashLabel
        mr={2}
        variant={'small'}
        hash={commit.hash}
        repoURL={repoURL}
        platform={platform}
      />
      <Link
        to={commitAddress(props.repoAddr, commit.hash)}
        className={'mdc-theme-text-primary-on-background'}>
        {commit.message}
      </Link>
      <ListItemSecondaryText>
        <strong>{commit.author.name} </strong>
        committed{' '}
        {formatDistanceToNow(new Date(commit.author.date), {addSuffix: true})}
      </ListItemSecondaryText>
    </div>
  );
};

interface CommitsHashLabelProp extends LabelProps {
  hash: string;
  platform: string;
  repoURL: string;
}

export const CommitsHashLabel: React.FC<CommitsHashLabelProp> = ({
  hash,
  platform,
  repoURL,
  ...props
}) => {
  return (
    <a
      href={CommitURL(platform, repoURL, hash)}
      target="_blank"
      rel="noopener noreferrer">
      <Label {...props}>
        <FontAwesomeIcon icon={['fab', platform as any]} />{' '}
        <code>{hash.substr(0, 7)}</code>
      </Label>
    </a>
  );
};

const RiskPoints: React.FC<any> = ({risk}) => {
  const riskLevel = Math.ceil(risk * 10);
  const [numberMode, setMode] = useState(false);

  function handelClick() {
    setMode(!numberMode);
  }

  const strRisk = Math.round(risk * 100) + '%';
  if (numberMode) {
    return (
      <div className="risk-score" onClick={handelClick}>
        {risk !== undefined ? (
          <span className={'mdc-typography--body1'}>{strRisk}</span>
        ) : (
          'No Value'
        )}
      </div>
    );
  }

  return (
    <ScorePoints
      className="risk-score"
      level={riskLevel}
      onClick={handelClick}
    />
  );
};

interface CommitsListProps {
  commits: RepositoryCommits_Filtered_repository['commits']['edges'];
  repoAddr: any; // todo: should be removed completely from the source code
  repoURL: string;
  platform: string;
}

export const CommitsList: React.FC<CommitsListProps> = (props) => {
  const [selectedCommit, selectCommit] = useState<string>();

  if (!props.commits?.length)
    return (
      <Blankslate>
        <GitCommitIcon size={'large'} />
        <h3>There aren’t any commits.</h3>
        <p>
          We will keep watching. When new commits get analyzed, we will show
          them here.
        </p>
      </Blankslate>
    );

  // const selectCommit = (id: string) => {
  //     ShowModal(
  //         <Modal title="Commit Details">
  //
  //         </Modal>
  //     );
  // };

  function handelClose() {
    selectCommit(undefined);
  }

  return (
    <div className="no-pointer divided-list mdc-list mdc-list--two-line mdc-list--avatar-list">
      {selectedCommit && (
        <CommitDetailsModal
          handelClose={handelClose}
          commitID={selectedCommit}
          repoAddr={props.repoAddr}
          repoURL={props.repoURL}
          platform={props.platform}
        />
      )}

      {props.commits.map(
        (edge) =>
          edge &&
          edge.node && (
            <CommitsListItem
              key={edge.node.id}
              repoURL={props.repoURL}
              commit={edge.node} //fixme: should remove "any"
              platform={props.platform}
              onSelect={selectCommit}
            />
          )
      )}

      {props.children}
    </div>
  );
};

interface LoadMoreCardButtonProps {
  loadMore: LoadMoreFn<any>; //fixme: remove `any`, maybe we cannot have LoadMoreCardButton as standalone component
  isLoading: boolean;
  hasMore: boolean;
  itemsCount?: number;
  pageSize: number;
  label: string;
}

//todo: rename to ListLoadMoreButton
export const LoadMoreCardButton: React.FC<LoadMoreCardButtonProps> = (
  props
) => {
  // Control if we should render the load more button
  if (props.itemsCount !== undefined && props.itemsCount < props.pageSize)
    return null;

  function handelLoadNext() {
    if (props.isLoading || !props.hasMore) return;

    props.loadMore(props.pageSize);
  }

  return (
    <CardActions fullBleed>
      <CardActionButton
        trailingIcon={
          props.isLoading && <FontAwesomeIcon icon={faSpinner} spin />
        }
        disabled={!props.hasMore || props.isLoading}
        onClick={handelLoadNext}
        label={props.label}
      />
    </CardActions>
  );
};
