import React, {Suspense, useEffect, useState} from 'react';
import {connect, useDispatch} from 'react-redux';
import {
  fetchModels,
  fetchRepository,
  refreshSourceInfo,
  stopRepositoryProcess,
  triggerProcess,
} from '../actions';
import {Card} from '@rmwc/card';
import {Typography} from '@rmwc/typography';
import {ListDivider} from '@rmwc/list';
import {GridCell, GridRow} from '@rmwc/grid';
import {format} from 'date-fns';
import {faSync} from '@fortawesome/free-solid-svg-icons';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {RepositoryAddress} from '../types';
import {Button} from '@rmwc/button';
import {
  useFragment,
  useLazyLoadQuery,
  useMutation,
  useRefetchableFragment,
} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {RepositorySettings_repository$key} from './__generated__/RepositorySettings_repository.graphql';
import {RepositorySettings_Info_repository$key} from './__generated__/RepositorySettings_Info_repository.graphql';
import {JobsList} from './JobsList';
import {ModelsList} from './ModelsList';
import {RepositoryLayout} from './RepositoryLayout';
import {Page404} from '../../ui/Page404';
import {PageSpinner} from '../../ui/Layout';
import {RouteComponentProps, useHistory, useParams} from 'react-router-dom';
import {IsWorking} from './Progress';
import {faPlayCircle, faStopCircle} from '@fortawesome/free-regular-svg-icons';
import {RepositorySettings_Jobs_repository$key} from './__generated__/RepositorySettings_Jobs_repository.graphql';
import {RepositorySettingsQuery} from './__generated__/RepositorySettingsQuery.graphql';
import {AppState} from '../../store/types';
import {toStrAddr} from '../reducers';
import {RepositoryChecksConfig} from './RepositoryChecksConfig';
import {Helmet} from 'react-helmet';
import {dialogs} from '../../ui/dialogs';
import {notify} from '../../ui/snackbar';
import {RepositorySettingsMutation} from './__generated__/RepositorySettingsMutation.graphql';

const SETTINGS_SCREEN_JOBS_PAGE_SIZE = 10;

interface RepositorySettingsProps {
  viewerRole?: string | null;
  repoAddr: RepositoryAddress;
  repository: RepositorySettings_repository$key;
  pageSize: number;
}

export const RepositorySettings: React.FC<RepositorySettingsProps> = (
  props
) => {
  const repository = useFragment(
    graphql`
      fragment RepositorySettings_repository on Repository
      @argumentDefinitions(jobs_count: {type: "Int"}) {
        id
        databaseId
        name
        ...RepositorySettings_Info_repository
        ...RepositoryChecksConfig_repository
        ...RepositorySettings_Jobs_repository @arguments(count: $jobs_count)
      }
    `,
    props.repository
  );

  return (
    <GridRow>
      <GridCell span={12}>
        <RepositoryInfoCard repository={repository} />
      </GridCell>

      <GridCell span={12}>
        <RepositoryChecksConfig repository={repository} />
      </GridCell>

      <GridCell span={12}>
        <JobsCard
          repoAddr={props.repoAddr}
          repository={repository}
          pageSize={props.pageSize}
        />
      </GridCell>

      {props.viewerRole === 'SITE_ADMIN' && (
        <>
          <GridCell span={12}>
            <ModelsCardContainer
              repoAddr={props.repoAddr}
              repository={repository}
            />
          </GridCell>
          <GridCell span={12}>
            <DeleteRepositoryCard
              repositoryID={repository.id}
              name={repository.name}
            />
          </GridCell>
        </>
      )}
    </GridRow>
  );
};

interface RepositoryInfoCardProps {
  repository: RepositorySettings_Info_repository$key;
}

const RepositoryInfoCard: React.FC<RepositoryInfoCardProps> = (props) => {
  const [repository, refetch] = useRefetchableFragment(
    graphql`
      fragment RepositorySettings_Info_repository on Repository
      @refetchable(queryName: "RepositorySettingsInfoRefreshQuery") {
        name
        databaseId
        providerSCM
        source {
          description
          defaultBranch
          createdAt
          url
        }
        owner {
          slug
        }

        collaboratorsCount
      }
    `,
    props.repository
  );

  const dispatch = useDispatch<any>();
  const [isFetching, setFetching] = useState(false);
  return (
    <Card outlined>
      <div style={{margin: '15px'}}>
        <Typography use="headline5" style={{marginBottom: '5px'}}>
          Repository Info
        </Typography>
        <Button
          className="float-right"
          disabled={isFetching}
          icon={
            isFetching ? (
              <FontAwesomeIcon icon={faSync} spin />
            ) : (
              <FontAwesomeIcon icon={['fab', repository.providerSCM as any]} />
            )
          }
          onClick={() => {
            setFetching(true);
            dispatch(
              refreshSourceInfo(
                {
                  platform: repository.providerSCM as any,
                  owner: repository.owner.slug as any,
                  repo: repository.name,
                },
                repository.databaseId
              )
            ).finally(() => {
              refetch({}, {fetchPolicy: 'store-and-network'});
              setFetching(false);
            });
          }}>
          Sync
        </Button>

        {repository.source.description ? (
          <>
            <p>Description: {repository.source.description}</p>
            <ListDivider />
          </>
        ) : (
          ''
        )}
        <p>Default Branch: {repository.source.defaultBranch}</p>
        <ListDivider />
        <p>Collaborators Count: {repository.collaboratorsCount}</p>
        <ListDivider />
        <p>
          Created on:{' '}
          {format(new Date(repository.source.createdAt), 'MMMM d, yyyy')}
        </p>
        <ListDivider />
        <p>
          URL:{' '}
          <a
            href={repository.source.url}
            target="_blank"
            rel="noopener noreferrer">
            {repository.source.url}
          </a>
        </p>
      </div>
    </Card>
  );
};

interface JobsCardProps {
  repository: RepositorySettings_Jobs_repository$key;
  repoAddr: any;
  pageSize: number;
}

const JobsCard: React.FC<JobsCardProps> = ({repoAddr, ...props}) => {
  const repository = useFragment(
    graphql`
      fragment RepositorySettings_Jobs_repository on Repository
      @argumentDefinitions(count: {type: "Int"}) {
        databaseId
        status
        ...JobsList_repository @arguments(count: $count)
      }
    `,
    props.repository
  );

  const [isLoading, setLoading] = useState(false);
  const dispatch = useDispatch<any>();

  return (
    <Card outlined>
      <div style={{margin: '15px 15px 0'}}>
        <Typography use="headline5" style={{marginBottom: '5px'}}>
          Processes
        </Typography>

        {IsWorking(repository.status) ? (
          <Button
            className="float-right"
            disabled={isLoading}
            icon={<FontAwesomeIcon icon={faStopCircle} />}
            onClick={() => {
              setLoading(true);
              dispatch(stopRepositoryProcess(repository.databaseId))
                .then(() => dispatch(fetchRepository(repoAddr)))
                .finally(() => setLoading(false));
            }}>
            Stop
          </Button>
        ) : (
          <Button
            className="float-right"
            disabled={isLoading}
            icon={<FontAwesomeIcon icon={faPlayCircle} />}
            onClick={() => {
              setLoading(true);
              dispatch(triggerProcess(repository.databaseId))
                .then(() => dispatch(fetchRepository(repoAddr)))
                .finally(() => setLoading(false));
            }}>
            Trigger
          </Button>
        )}

        <JobsList repository={repository} pageSize={props.pageSize} />
      </div>
    </Card>
  );
};

interface ModelsCardProps {
  repoAddr: any;
  repository: any;
}

const ModelsCard: React.FC<ModelsCardProps> = ({repoAddr, repository}) => {
  const dispatch = useDispatch();

  useEffect(
    function () {
      repository.id && dispatch(fetchModels(repoAddr, repository.id));
    },
    [dispatch, repoAddr, repository.id]
  );

  if (!Array.isArray(repository.models) || repository.models.length < 1)
    return null;

  return (
    <Card outlined>
      <div style={{margin: '15px'}}>
        <Typography use="headline5" style={{marginBottom: '5px'}}>
          Models
        </Typography>
        <ModelsList models={repository.models} />
      </div>
    </Card>
  );
};

const mapStateToProps = (state: AppState, ownProps: any) => {
  return {
    repository: state.repositories.reposList[toStrAddr(ownProps.repoAddr)] || {
      id: ownProps.repository.databaseId,
    },
  };
};

const ModelsCardContainer = connect(mapStateToProps)(ModelsCard);

interface RepositorySettingsScreenProps extends RouteComponentProps {}

export const RepositorySettingsScreen: React.FC<RepositorySettingsScreenProps> = ({
  location,
}) => {
  const repoAddr: any = useParams();
  const {platform, owner, repo} = repoAddr;

  const {repository, viewer} = useLazyLoadQuery<RepositorySettingsQuery>(
    graphql`
      query RepositorySettingsQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $jobs_count: Int
      ) {
        viewer {
          role
        }
        repository(provider: $provider, owner: $owner, name: $name) {
          ...RepositoryLayout_repository
          ...RepositorySettings_repository @arguments(jobs_count: $jobs_count)
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      jobs_count: SETTINGS_SCREEN_JOBS_PAGE_SIZE,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository) return <Page404 location={location} />;

  return (
    <RepositoryLayout repository={repository}>
      <Helmet>
        <title>Settings</title>
      </Helmet>
      <Suspense fallback={<PageSpinner />}>
        <RepositorySettings
          viewerRole={viewer?.role}
          repoAddr={repoAddr}
          repository={repository}
          pageSize={SETTINGS_SCREEN_JOBS_PAGE_SIZE}
        />
      </Suspense>
    </RepositoryLayout>
  );
};

interface DeleteRepositoryCardProps {
  repositoryID: string;
  name: string;
}

const DeleteRepositoryCard: React.FC<DeleteRepositoryCardProps> = ({
  repositoryID,
  name,
}) => {
  const history = useHistory();
  const [commit, isInFlight] = useMutation<RepositorySettingsMutation>(graphql`
    mutation RepositorySettingsMutation($id: ID!) {
      deleteRepository(id: $id) {
        repository {
          id
        }
      }
    }
  `);

  function deleteRepository() {
    commit({
      variables: {
        id: repositoryID,
      },
      updater: (store, data) => {
        const repoID = data?.deleteRepository?.repository?.id;
        if (!repoID) return;

        const repository = store.get(repoID);
        repository?.invalidateRecord();
      },
      onCompleted: (data) => {
        const repoID = data?.deleteRepository?.repository?.id;
        if (!repoID) return;
        notify({
          title: 'Repository deleted from Repofuel.',
        });
        history.push(`/repos/`);
      },
    });
  }

  function confirmDialog(value?: string) {
    dialogs
      .prompt({
        title: 'Are you absolutely sure?',
        body: (
          <div>
            This action cannot be undone. This will permanently delete the{' '}
            <b>{name}</b> repository from Repofuel, including all of its
            historical data.
            <p>
              Please type <b>{name}</b> to confirm.
            </p>
          </div>
        ),
        acceptLabel: 'Delete this repository',
        inputProps: {
          outlined: true,
          invalid: value !== undefined,
          placeholder: value,
        },
      })
      .then((res) => {
        if (res === null) return;
        if (res === name) return deleteRepository();
        confirmDialog(res);
      });
  }

  return (
    <Card
      outlined
      style={{
        borderColor: 'var(--mdc-theme-error)',
      }}>
      <div style={{margin: '15px'}}>
        <Typography use="headline5" style={{marginBottom: '5px'}}>
          Delete repository
        </Typography>
        <p>
          Deleting a repository will remove it entirely from Repofuel, including
          all of its historical data.
        </p>
        <Button
          danger
          unelevated
          disabled={isInFlight}
          onClick={() => confirmDialog()}>
          Delete repository
        </Button>
      </div>
    </Card>
  );
};
