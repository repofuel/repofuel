import React, {Suspense, useEffect, useRef, useState} from 'react';
import {RouteComponentProps, useParams} from 'react-router-dom';
import {LineRipple} from '@rmwc/line-ripple';
import {
  CommitsList,
  LoadMoreCardButton,
  SkeletonCommitsList,
} from './CommitsList';
import {RepositoryAddress} from '../types';
import {Select} from '@rmwc/select';
import {FloatingLabel} from '@rmwc/floating-label';
import {Card} from '@rmwc/card';
import {ListDivider} from '@rmwc/list';
import {MenuItem} from '@rmwc/menu';

import {Slider, SliderOnChangeEventT} from '@rmwc/slider';
import {GitBranchIcon, PersonIcon} from '@primer/octicons-react';
import './RepositoryCommits.scss';

import {useFragment, useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Page404} from '../../ui/Page404';
import {RepositoryLayout} from './RepositoryLayout';
import {RepositoryCommits_repository$key} from './__generated__/RepositoryCommits_repository.graphql';
import {RepositoryCommitsQuery} from './__generated__/RepositoryCommitsQuery.graphql';
import {useStatusTracker} from './Progress';
import {usePaginationFragment} from 'react-relay/lib/hooks';
import {CommitFilters} from './__generated__/RepositoryCommitsRefreshQuery.graphql';
import {RepositoryCommits_Filtered_repository$key} from './__generated__/RepositoryCommits_Filtered_repository.graphql';
import {RepositoryCommits_SelectBranchItems_repository$key} from './__generated__/RepositoryCommits_SelectBranchItems_repository.graphql';
import {RepositoryCommits_SelectDeveloperItems_repository$key} from './__generated__/RepositoryCommits_SelectDeveloperItems_repository.graphql';
import styled from 'styled-components';
import {Helmet} from 'react-helmet';

const COMMIT_SCREEN_PAGE_SIZE = 20;

// function useQuery() {
//     return new URLSearchParams(useLocation().search);
// }

interface RepositoryCommitsScreenProps extends RouteComponentProps {}

export const RepositoryCommitsScreen: React.FC<RepositoryCommitsScreenProps> = ({
  location,
  ...props
}) => {
  const repoAddr: any = useParams();
  const {platform, owner, repo} = repoAddr;

  const {repository} = useLazyLoadQuery<RepositoryCommitsQuery>(
    graphql`
      query RepositoryCommitsQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $commits_count: Int
      ) {
        repository(provider: $provider, owner: $owner, name: $name) {
          ...RepositoryLayout_repository
          ...RepositoryCommits_repository
            @arguments(commits_count: $commits_count)
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      commits_count: COMMIT_SCREEN_PAGE_SIZE,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository) return <Page404 location={location} />;

  return (
    <RepositoryLayout repository={repository}>
      <Helmet>
        <title>Commits</title>
      </Helmet>
      <RepositoryCommits
        repoAddr={repoAddr}
        repository={repository}
        pageSize={COMMIT_SCREEN_PAGE_SIZE}
      />
    </RepositoryLayout>
  );
};

interface RepositoryCommitsProps {
  repository: RepositoryCommits_repository$key;
  repoAddr: RepositoryAddress; //fixme: should be removed
  pageSize: number;
}

export const RepositoryCommits: React.FC<RepositoryCommitsProps> = (props) => {
  const [filters, setFilters] = useState<CommitFilters>({});
  const repository = useFragment(
    graphql`
      fragment RepositoryCommits_repository on Repository
      @refetchable(queryName: "RepositoryCommitsRefreshQuery")
      @argumentDefinitions(
        commits_count: {type: "Int"}
        commits_cursor: {type: "String"}
        filters: {type: "CommitFilters"}
      ) {
        id
        source {
          defaultBranch
          url
        }
        providerSCM
        status
        ...RepositoryCommits_SelectDeveloperItems_repository
        ...RepositoryCommits_SelectBranchItems_repository
        ...RepositoryCommits_Filtered_repository
          @arguments(commits_count: $commits_count)
      }
    `,
    props.repository
  );

  function handelSelectBranch(evt: React.ChangeEvent<HTMLSelectElement>) {
    setFilters({...filters, branch: evt.currentTarget.value});
  }

  function handelSelectDeveloper(evt: React.ChangeEvent<HTMLSelectElement>) {
    setFilters({
      ...filters,
      developerName: evt.currentTarget.value || undefined,
    });
  }

  function handelSelectMinRisk(evt: SliderOnChangeEventT) {
    setFilters({
      ...filters,
      minRisk: evt.detail.value ? evt.detail.value / 100 : undefined,
    });
  }

  return (
    <Card outlined>
      <div className="commit-list-select-container">
        <Select
          label="Branch"
          enhanced
          icon={<GitBranchIcon />}
          value={filters.branch || repository.source.defaultBranch}
          onChange={handelSelectBranch}>
          <Suspense fallback={<span>Loading...</span>}>
            <SelectBranchItems
              selected={filters.branch}
              // repositoryId={repository.id}
              repository={repository}
            />
          </Suspense>
        </Select>

        <Select
          label="Developer"
          enhanced
          placeholder="All Developers"
          icon={<PersonIcon />}
          value={filters.developerName || undefined}
          onChange={handelSelectDeveloper}>
          <Suspense fallback={<span>Loading...</span>}>
            <SelectDeveloperItems
              selected={filters.developerName}
              // repositoryId={repository.id}
              repository={repository}
            />
          </Suspense>
        </Select>

        <div className="slider-box mdc-select">
          <div className="slider-box-inner">
            <Slider //className={"mdc-select"}
              value={(filters.minRisk && filters.minRisk * 100) || undefined}
              onChange={handelSelectMinRisk}
              step={1}
              max={99}>
              <FloatingLabel float shake>
                Risk
              </FloatingLabel>
            </Slider>
          </div>
          <LineRipple />
        </div>
      </div>
      <Suspense fallback={<SkeletonCommitsList />}>
        <RepositoryCommitsFiltered
          filters={filters}
          pageSize={props.pageSize}
          repository={repository}
          repoAddr={props.repoAddr}
        />
      </Suspense>
    </Card>
  );
};

const CentralSmallText = styled.div`
  padding: 16px;
`;

interface SelectDeveloperItemsProps {
  repository: RepositoryCommits_SelectDeveloperItems_repository$key;
  // repositoryId: string
  selected?: string | null;
}

export const SelectDeveloperItems: React.FC<SelectDeveloperItemsProps> = (
  props
) => {
  //todo: move the query to a different file and rename the query
  //todo: return the developers from graphql as a connection
  // const {node: repository} = useLazyLoadQuery<RepositoryCommitsDevelopersQuery>(graphql`
  //     query RepositoryCommitsDevelopersQuery($id:ID!) {
  //         node(id:$id) {
  //             ... on Repository{
  //                 developerNames
  //             }
  //         }
  //     }
  // `, {
  //     id: props.repositoryId
  // }, {fetchPolicy: "store-and-network"})
  const repository = useFragment(
    graphql`
      fragment RepositoryCommits_SelectDeveloperItems_repository on Repository {
        developerNames
      }
    `,
    props.repository
  );

  if (!repository?.developerNames?.length) {
    return (
      <>
        <ListDivider />
        <CentralSmallText>Nothing to show</CentralSmallText>
      </>
    );
  }

  const selected = props.selected;
  return (
    <>
      {repository.developerNames.map((dev) => (
        <MenuItem key={dev} value={dev} activated={selected === dev}>
          {dev}
        </MenuItem>
      ))}
    </>
  );
};

interface SelectBranchItemsProps {
  repository: RepositoryCommits_SelectBranchItems_repository$key;
  // repositoryId: string
  selected?: string | null;
}

export const SelectBranchItems: React.FC<SelectBranchItemsProps> = (props) => {
  //todo: move the query to a different file and rename the query
  //todo: return the branches from graphql as a connection
  // const {node: repository} = useLazyLoadQuery<RepositoryCommitsBranchesQuery>(graphql`
  //     query RepositoryCommitsBranchesQuery($id:ID!) {
  //         node(id:$id) {
  //             ... on Repository{
  //                 branches{
  //                     name
  //                 }
  //             }
  //         }
  //     }
  // `, {
  //     id: props.repositoryId
  // }, {fetchPolicy: "store-and-network"})
  const repository = useFragment(
    graphql`
      fragment RepositoryCommits_SelectBranchItems_repository on Repository {
        branches {
          name
        }
      }
    `,
    props.repository
  );

  if (!repository?.branches?.length) {
    return <CentralSmallText>Nothing to show</CentralSmallText>;
  }

  const selected = props.selected;
  return (
    <>
      {repository.branches.map(({name}) => (
        <MenuItem key={name} value={name} activated={selected === name}>
          {name}
        </MenuItem>
      ))}
    </>
  );
};

interface RepositoryCommitsFilteredProps {
  repository: RepositoryCommits_Filtered_repository$key;
  repoAddr: RepositoryAddress; //fixme: should be removed
  pageSize: number;
  filters: CommitFilters;
}

export const RepositoryCommitsFiltered: React.FC<RepositoryCommitsFilteredProps> = (
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
      fragment RepositoryCommits_Filtered_repository on Repository
      @refetchable(queryName: "RepositoryCommitsFilteredRefreshQuery")
      @argumentDefinitions(
        commits_count: {type: "Int"}
        commits_cursor: {type: "String"}
        filters: {type: "CommitFilters"}
      ) {
        #TODO: all fields expects "commits" should be left up to reduce unneseeary refetch
        status
        source {
          url
        }
        providerSCM
        commits(
          first: $commits_count
          after: $commits_cursor
          filters: $filters
        ) @connection(key: "RepositoryCommitsFiltered_repository_commits") {
          edges {
            node {
              id
              ...CommitsList_Item_commit
            }
          }
        }
      }
    `,
    props.repository
  );

  // todo: replace the implementation to use useState and startTransition when adopt concurrent mood
  const isFirstRun = useRef(true);
  useEffect(() => {
    if (isFirstRun.current) {
      isFirstRun.current = false;
      return;
    }

    refetch({filters: props.filters}, {fetchPolicy: 'store-and-network'});
  }, [refetch, props.filters]);

  function handelRefetch() {
    refetch({filters: props.filters}, {fetchPolicy: 'store-and-network'});
  }

  useStatusTracker(
    repository.status,
    handelRefetch,
    'PREDICTING',
    'READY',
    'WATCHED'
  );

  return (
    <CommitsList
      repoAddr={props.repoAddr}
      repoURL={repository.source.url}
      platform={repository.providerSCM}
      commits={repository.commits?.edges}>
      <LoadMoreCardButton
        label="Load more commits"
        pageSize={props.pageSize}
        itemsCount={repository.commits?.edges?.length || 0}
        isLoading={isLoadingNext}
        hasMore={hasNext}
        loadMore={loadNext}
      />
    </CommitsList>
  );
};
