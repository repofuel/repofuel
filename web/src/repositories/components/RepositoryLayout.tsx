import React, {Suspense} from 'react';
import {useFragment, useMutation} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import Layout, {PageSpinner} from '../../ui/Layout';
import {Grid, GridCell} from '@rmwc/grid';
import {
  IsWorking,
  RepositoryProgress,
  useProgressSubscription,
} from './Progress';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {JobStatus} from './RepositoriesList';
import {Link, useParams} from 'react-router-dom';
import {
  DrawerContent,
  DrawerHeader,
  DrawerSubtitle,
  DrawerTitle,
} from '@rmwc/drawer';
import {List, ListItemText} from '@rmwc/list';
import {
  GearIcon,
  GitCommitIcon,
  GitPullRequestIcon,
  GraphIcon,
} from '@primer/octicons-react';
import {RepositoryLayout_repository$key} from './__generated__/RepositoryLayout_repository.graphql';
import {RepositoryLayout_Menu_repository$key} from './__generated__/RepositoryLayout_Menu_repository.graphql';
import './RepositoryScreen.scss';
import {formatDistanceToNow} from 'date-fns';
import {Helmet} from 'react-helmet';

import {RepositoryLayout_MonitorMutation} from './__generated__/RepositoryLayout_MonitorMutation.graphql';
import {RepositoryLayout_StopMonitoringMutation} from './__generated__/RepositoryLayout_StopMonitoringMutation.graphql';
import {RepositoryLayout_MonitorButton_repository$key} from './__generated__/RepositoryLayout_MonitorButton_repository.graphql';
import {ButtonGroup} from '../../ui/ButtonGroup';
import {Button} from 'rmwc';
import {NavListItem} from '../../ui/List';

interface RepositoryLayoutProps {
  repository: RepositoryLayout_repository$key;
}

export const RepositoryLayout: React.FC<RepositoryLayoutProps> = (props) => {
  const repository = useFragment(
    graphql`
      fragment RepositoryLayout_repository on Repository {
        id
        name
        owner {
          slug
        }
        status
        PredictionStatus
        viewerIsMonitor
        ...RepositoryLayout_MonitorButton_repository
        ...RepositoryLayout_Menu_repository
        ...Progress_repository
      }
    `,
    props.repository
  );

  useProgressSubscription([repository.id]);

  const repoAddr: any = useParams();

  let status = repository.status;
  if (
    repository.PredictionStatus &&
    repository.PredictionStatus > 0 &&
    status === 'READY'
  ) {
    status = 'WATCHED';
  }

  const defaultTitle =
    repository.owner.slug + '/' + repository.name + ' - Repofuel';
  return (
    <Layout menuItems={<RepositoryMenu repository={repository} />}>
      <Helmet
        titleTemplate={'%s · ' + defaultTitle}
        defaultTitle={defaultTitle}
      />

      <Grid className="fixed-width-layout">
        <GridCell span={12}>
          <RepositoryHeader {...repoAddr} status={status}>
            <RepositoryMonitorButton repository={repository} />
          </RepositoryHeader>
          {IsWorking(repository.status) && (
            <RepositoryProgress repository={repository} />
          )}
        </GridCell>
        <GridCell span={12}>
          <Suspense fallback={<PageSpinner />}>{props.children}</Suspense>
        </GridCell>
      </Grid>
    </Layout>
  );
};

const RepositoryHeader: React.FC<any> = ({
  platform,
  owner,
  repo,
  status,
  ...props
}) => {
  return (
    <div className="repository-head">
      <span className="mdc-typography--headline5">
        <FontAwesomeIcon className="margin-right" icon={['fab', platform]} />
        <Link to={`/orgs/${platform}/${owner}`}>{owner}</Link>
        <PathDivider />
        <strong>
          <Link to={`/repos/${platform}/${owner}/${repo}`}>{repo}</Link>
        </strong>
      </span>
      {status && (
        <span className="repo-header-status">
          <JobStatus status={status} />
        </span>
      )}
      {props.children}
    </div>
  );
};

const RepositoryMenu: React.FC<{
  repository: RepositoryLayout_Menu_repository$key;
}> = (props) => {
  const repository = useFragment(
    graphql`
      fragment RepositoryLayout_Menu_repository on Repository {
        providerSCM
        name
        owner {
          slug
        }
        viewerCanAdminister
        createdAt
      }
    `,
    props.repository
  );

  const repoURL = `/repos/${repository.providerSCM}/${repository.owner.slug}/${repository.name}`;

  return (
    <DrawerContent>
      <DrawerHeader>
        <DrawerTitle tag={Link} to={repoURL}>
          {repository.name}
        </DrawerTitle>
        <DrawerSubtitle title={repository.createdAt as string}>
          Added
          {formatDistanceToNow(new Date(repository.createdAt), {
            addSuffix: true,
          })}
        </DrawerSubtitle>
      </DrawerHeader>
      <List>
        <NavListItem exact to={repoURL}>
          {/*<FontAwesomeIcon className="list-close-icon" icon={faChartLine}/>*/}
          <GraphIcon className="list-close-icon" />
          <ListItemText>Dashboard</ListItemText>
        </NavListItem>
        <NavListItem to={`${repoURL}/commits`}>
          <GitCommitIcon className="list-close-icon" />
          <ListItemText>Commits</ListItemText>
          {/*<ListLoadingStatus active={repository.isCommitsLoading}/>*/}
        </NavListItem>
        <NavListItem to={`${repoURL}/pulls`}>
          <GitPullRequestIcon className="list-close-icon" />
          <ListItemText>Pull Requests</ListItemText>
        </NavListItem>
        {repository.viewerCanAdminister && (
          <NavListItem to={`${repoURL}/settings`}>
            <GearIcon className="list-close-icon" />
            <ListItemText>Settings</ListItemText>
          </NavListItem>
        )}
      </List>
    </DrawerContent>
  );
};

const PathDivider: React.FC = () => {
  return <span className="path-divider">/</span>;
};

interface RepositoryMonitorButtonProps {
  repository: RepositoryLayout_MonitorButton_repository$key;
}

const RepositoryMonitorButton: React.FC<RepositoryMonitorButtonProps> = (
  props
) => {
  const repository = useFragment(
    graphql`
      fragment RepositoryLayout_MonitorButton_repository on Repository {
        id
        viewerIsMonitor
        monitorCount
      }
    `,
    props.repository
  );

  const [
    commitSubscribe,
    isSubscribeInFlight,
  ] = useMutation<RepositoryLayout_MonitorMutation>(graphql`
    mutation RepositoryLayout_MonitorMutation($repoID: ID!) {
      monitorRepository(id: $repoID) {
        repository {
          ...RepositoryLayout_MonitorButton_repository
        }
      }
    }
  `);

  const [
    commitUnsubscribe,
    isUnsubscribeInFlight,
  ] = useMutation<RepositoryLayout_StopMonitoringMutation>(graphql`
    mutation RepositoryLayout_StopMonitoringMutation($repoID: ID!) {
      stopRepositoryMonitoring(id: $repoID) {
        repository {
          ...RepositoryLayout_MonitorButton_repository
        }
      }
    }
  `);

  const commit = repository.viewerIsMonitor
    ? commitUnsubscribe
    : commitSubscribe;

  function handleClick() {
    commit({
      variables: {
        repoID: repository.id,
      },
    });
  }

  return (
    <ButtonGroup className="float-right">
      <Button
        outlined
        disabled={isSubscribeInFlight || isUnsubscribeInFlight}
        onClick={handleClick}>
        {repository.viewerIsMonitor ? 'Stop Monitoring' : 'Monitor'}
      </Button>
      <Button outlined disabled>
        {repository.monitorCount}
      </Button>
    </ButtonGroup>
  );
};
