import React, {useState} from 'react';
import {
  CollapsibleList,
  List,
  ListItem,
  ListItemMeta,
  ListItemText,
} from '@rmwc/list';
import {format, formatDistanceStrict, formatDistanceToNow} from 'date-fns';
import {JobStatus} from './RepositoriesList';
import {Tooltip} from '@rmwc/tooltip';
import './JobsList.scss';
import {
  faChevronDown,
  faChevronRight,
  faSpinner,
} from '@fortawesome/free-solid-svg-icons';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {Flash} from '@primer/components';
import {usePaginationFragment} from 'react-relay/lib/hooks';
import graphql from 'babel-plugin-relay/macro';
import {JobsList_repository$key} from './__generated__/JobsList_repository.graphql';
import {useFragment} from 'react-relay/hooks';
import {
  JobsListItem_job,
  JobsListItem_job$key,
} from './__generated__/JobsListItem_job.graphql';
import {CardActionButton, CardActions} from '@rmwc/card';
import {useStatusTracker} from './Progress';

interface JobsListProps {
  repository: JobsList_repository$key;
  pageSize: number;
}

export const JobsList: React.FC<JobsListProps> = (props) => {
  const {
    data: repository,
    refetch,
    hasNext,
    isLoadingNext,
    loadNext,
  } = usePaginationFragment(
    graphql`
      fragment JobsList_repository on Repository
      @refetchable(queryName: "JobsListRefetchableQuery")
      @argumentDefinitions(count: {type: "Int"}, cursor: {type: "String"}) {
        status
        jobs(first: $count, after: $cursor)
          @connection(key: "JobsList_repository_jobs") {
          edges {
            node {
              id
              ...JobsListItem_job
            }
          }
        }
      }
    `,
    props.repository
  );

  function handelClick() {
    if (isLoadingNext || !hasNext) return;

    loadNext(props.pageSize);
  }

  function handelRefetch() {
    refetch({}, {fetchPolicy: 'store-and-network'});
  }

  useStatusTracker(repository.status, handelRefetch, 'READY');

  return (
    <List className="divided-collapsible-list">
      {repository.jobs?.edges?.map((edge, i) => {
        if (!edge?.node) return null;
        return (
          <JobsListItem
            key={edge.node.id}
            job={edge.node}
            defaultOpen={i === 0}
          />
        );
      })}

      {(repository.jobs?.edges?.length || 0) >= props.pageSize && (
        <CardActions fullBleed>
          <CardActionButton
            disabled={!hasNext || isLoadingNext}
            onClick={handelClick}
            label="Load more process"
            trailingIcon={
              isLoadingNext && <FontAwesomeIcon icon={faSpinner} spin />
            }
          />
        </CardActions>
      )}
    </List>
  );
};

interface JobsListItemProps {
  job: JobsListItem_job$key;
  defaultOpen: boolean;
}

const JobsListItem: React.FC<JobsListItemProps> = (props) => {
  const job = useFragment(
    graphql`
      fragment JobsListItem_job on Job {
        id
        statusLog {
          status
          statusText
          startedAt
        }
        createdAt
        error
      }
    `,
    props.job
  );

  const [isOpen, setOpen] = useState(props.defaultOpen);
  let {statusLog: log, error} = job;
  if (!log) log = [];

  const isExpandable = log.length > 1;
  const lastStage = log[log.length - 1] || {status: 'Not started'};
  return (
    <CollapsibleList
      handle={
        <ListItem
          disabled={!isExpandable}
          onClick={() => isExpandable && setOpen(!isOpen)}>
          <ListItemText>
            <JobStatus status={lastStage.status} />
            {lastStage.startedAt && (
              <Tooltip
                content={format(
                  new Date(lastStage.startedAt),
                  "'Started' MMM d 'at' h:m a"
                )}
                showArrow
                align="right">
                <span className="process-time">
                  {' '}
                  {formatDistanceToNow(new Date(lastStage.startedAt), {
                    addSuffix: true,
                  })}
                </span>
              </Tooltip>
            )}
          </ListItemText>
          <ListItemMeta>
            {isExpandable && (
              <FontAwesomeIcon icon={isOpen ? faChevronDown : faChevronRight} />
            )}
          </ListItemMeta>
        </ListItem>
      }>
      {isOpen && <JobsListItemDetails log={log} error={error} />}
    </CollapsibleList>
  );
};

interface JobsListItemDetailsProps {
  log: JobsListItem_job['statusLog'];
  error: string | null;
}

const JobsListItemDetails: React.FC<JobsListItemDetailsProps> = ({
  log = [],
  error,
}) => {
  if (!log) return null;

  const timeDiff = (stage: typeof log[0], i: number) => {
    const nextStage = log[i + 1];
    if (nextStage) {
      return formatDistanceStrict(
        new Date(stage.startedAt),
        new Date(nextStage.startedAt)
      );
    }

    return 'In Progress';
  };

  return (
    <>
      {error && (
        <Flash m={2} variant="danger">
          {error}
        </Flash>
      )}
      {log.map((stage, i) =>
        // do not show the last stage
        i + 1 === log.length ? null : (
          <div key={stage.status} className="sub-process">
            <span className="process-status">{stage.statusText}</span>
            {/*todo: refactor duplications with JobsListIteam*/}
            <Tooltip
              showArrow
              align="left"
              content={format(
                new Date(stage.startedAt),
                "'Started' MMM d 'at' h:m a"
              )}>
              <span className="process-time">{timeDiff(stage, i)}</span>
            </Tooltip>
          </div>
        )
      )}
    </>
  );
};
