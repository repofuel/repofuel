import React, {useEffect, useMemo, useRef} from 'react';
import {LinearProgress} from '@rmwc/linear-progress';
import graphql from 'babel-plugin-relay/macro';
import './Progress.scss';
import {useFragment, useSubscription} from 'react-relay/hooks';
import {Progress_repository$key} from './__generated__/Progress_repository.graphql';
import {Stage} from './__generated__/JobsListItem_job.graphql';
import {GraphQLSubscriptionConfig} from 'relay-runtime';
import {
  ProgressSubscription,
  ProgressSubscriptionResponse,
} from './__generated__/ProgressSubscription.graphql';
import {SelectorStoreUpdater} from 'relay-runtime/lib/store/RelayStoreTypes';

export const IsWorking = (status: Stage) => {
  switch (status) {
    case 'READY':
    case 'WATCHED':
    case 'FAILED':
    case 'CANCELED':
    case undefined:
      return false;
  }

  return true;
};

export const useStatusTracker = (
  status: Stage | undefined,
  cb: () => void,
  ...targets: Stage[]
) => {
  const isFirstRun = useRef(true);
  useEffect(() => {
    if (isFirstRun.current) {
      isFirstRun.current = false;
      return;
    }

    if (!targets.length) {
      cb();
      return;
    }

    for (let i = 0, len = targets.length; i < len; i++) {
      if (status === targets[i]) {
        cb();
        return;
      }
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);

  // todo: replace with this implementation when adopt concurrent mood
  // const [lastStatus, setStatus] = useState(status);
  // if (status !== lastStatus) {
  //     setStatus(status)
  //
  //     if (!targets.length) {
  //         cb() // todo: we should call it inside startTransition()
  //         return
  //     }
  //
  //     for (let i = 0, len = targets.length; i < len; i++) {
  //         if (status === targets[i]) {
  //             cb()
  //             return
  //         }
  //     }
  // }
};

const progressUpdater: SelectorStoreUpdater<ProgressSubscriptionResponse> = (
  store,
  data
) => {
  const node = store.get(data.changeProgress.target);
  if (!node) return;

  node.setValue(data.changeProgress.progress.status, 'status');
  node
    .getOrCreateLinkedRecord('progress', 'Progress')
    .copyFieldsFrom(
      store.getRootField('changeProgress').getLinkedRecord('progress')
    );
};

export const useProgressSubscription = (ids: string[]) => {
  const subscriptionConfig = useMemo<
    GraphQLSubscriptionConfig<ProgressSubscription>
  >(
    () => ({
      subscription: graphql`
        subscription ProgressSubscription($ids: [ID!]!) {
          changeProgress(ids: $ids) {
            target
            progress {
              total
              current
              status
            }
          }
        }
      `,
      variables: {ids: ids},
      updater: progressUpdater,
    }),
    // Disabling the linter should be avoided, this is an
    // exception since it is necessary.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    ids
  );

  useSubscription(subscriptionConfig);
};

interface ProgressProps {
  repository: Progress_repository$key;
}

export const RepositoryProgress: React.FC<ProgressProps> = (props) => {
  const repository = useFragment(
    graphql`
      fragment Progress_repository on Repository {
        progress {
          current
          total
        }
      }
    `,
    props.repository
  );

  if (!repository.progress) return null;

  const {total, current} = repository.progress;
  if (!total || !current) {
    return (
      <span className="progress-group">
        <LinearProgress />
        <span className="progress-caption">In progress...</span>
      </span>
    );
  }

  const percentage = current / total;
  return (
    <span className="progress-group">
      <LinearProgress progress={percentage} />
      <span className="progress-caption">
        {current} out of {total} ({Math.floor(percentage * 100)}%)
      </span>
    </span>
  );
};
