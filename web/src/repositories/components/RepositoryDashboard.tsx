import React, {useEffect, useMemo, useRef} from 'react';
import graphql from 'babel-plugin-relay/macro';

import {RepositoryDashboardQuery} from './__generated__/RepositoryDashboardQuery.graphql';
import {
  useFragment,
  useLazyLoadQuery,
  useRefetchableFragment,
} from 'react-relay/hooks';
import {CommitsList} from './CommitsList';
import {GridCell, GridRow} from '@rmwc/grid';
import {Typography} from '@rmwc/typography';
import {Card, CardActionButton, CardActions} from '@rmwc/card';
import {RiskyCommitsDoughnut} from './RiskyCommitsDoughnut';

import './RepositoryDashboard.scss';

import {Chart, ChartDataSets, TimeUnit} from 'chart.js';
import {Tooltip} from '@rmwc/tooltip';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faQuestionCircle} from '@fortawesome/free-regular-svg-icons';
import {RepositoryAddress} from '../types';
import Masonry from 'react-masonry-css';
import {Link, RouteComponentProps, useParams} from 'react-router-dom';
import {RepositoryLayout} from './RepositoryLayout';
import {RepositoryDashboard_repository$key} from './__generated__/RepositoryDashboard_repository.graphql';
import {Page404} from '../../ui/Page404';
import {Flash} from '@primer/components';
import {useStatusTracker} from './Progress';
import {faArrowRight} from '@fortawesome/free-solid-svg-icons';
import {FormatNumber} from '../../util/number-formate';
import {fillPointsOnTimeLine, generateTimeLine} from '../../util/chart';
import {RepositoryDashboard_tagsCount$key} from './__generated__/RepositoryDashboard_tagsCount.graphql';
import {RepositoryDashboard_commitsOverTime$key} from './__generated__/RepositoryDashboard_commitsOverTime.graphql';
import {RepositoryDashboard_avgCommitFilesOverTime$key} from './__generated__/RepositoryDashboard_avgCommitFilesOverTime.graphql';
import {PullRequestsList} from './PullRequestsList';

const DASHBOARD_SCREEN_PAGE_SIZE = 5;

export const BreakpointColumnsObj = {
  default: 2,
  839: 1,
};

interface RepositoryDashboardProps {
  repository: RepositoryDashboard_repository$key;
  repoAddr: RepositoryAddress; //todo: should be removed
  pageSize: number;
}

export const RepositoryDashboard: React.FC<RepositoryDashboardProps> = ({
  repoAddr,
  ...props
}) => {
  const [repository, refetch] = useRefetchableFragment(
    graphql`
      fragment RepositoryDashboard_repository on Repository
      @refetchable(queryName: "RepositoryDashboardRefreshQuery")
      #fixme: reanme $commits_count sincee it is used here for pulls also
      @argumentDefinitions(commits_count: {type: "Int"}) {
        id
        name
        branches {
          name
        }
        providerSCM
        source {
          url
          defaultBranch #fixme: it is not used, it is just cached so when we move to the commits page will render without waiting for it
        }
        status
        commits(first: $commits_count) {
          edges {
            node {
              id
              ...CommitsList_Item_commit
            }
          }
        }
        pullRequests(first: $commits_count) {
          edges {
            node {
              id
              ...PullRequestsList_Item_pullRequest
            }
          }
        }
        ...RepositoryDashboard_commitsOverTime
        ...RepositoryDashboard_tagsCount
        ...RepositoryDashboard_avgCommitFilesOverTime
        branchesCount
        buggyCommitsCount
        commitPredictionsCount
        commitsCount
        fixCommitsCount
        contributorsCount
        Confidence
        PredictionStatus
      }
    `,
    props.repository
  );

  function handelRefetch() {
    refetch({}, {fetchPolicy: 'store-and-network'});
  }

  useStatusTracker(
    repository.status,
    handelRefetch,
    'ANALYZING',
    'PREDICTING',
    'READY',
    'WATCHED'
  );

  const timeLine = generateTimeLine('YEAR', 'MONTHLY');

  return (
    <>
      {repository.PredictionStatus !== 0 && (
        <Flash
          variant="warning"
          style={{textAlign: 'center', marginBottom: '1.5em'}}>
          <strong>
            {repository.PredictionStatus === 6
              ? "The accuracy of the AI model for this repository isn't the best."
              : 'We need more data to build a high-quality AI model.'}
          </strong>
          <br />
          Repofuel will follow the repo and build a high-quality model once it
          has enough data.
        </Flash>
      )}
      <GridRow className="masonry-grid-width">
        <WidgetCard desc="Risky Commits">
          <RiskyCommitsDoughnut
            className="risk-chart-dash"
            commitNum={repository.commitsCount}
            riskyCommitNum={repository.buggyCommitsCount}
          />
        </WidgetCard>
        <WidgetCard desc="Contributors">
          {FormatNumber(repository.contributorsCount)}
        </WidgetCard>
        <WidgetCard desc="Commit Predictions">
          {FormatNumber(repository.commitPredictionsCount)}
        </WidgetCard>
        <WidgetCard desc="Total Commits">
          {FormatNumber(repository.commitsCount)}
        </WidgetCard>
      </GridRow>
      <Masonry
        breakpointCols={BreakpointColumnsObj}
        className="masonry-grid masonry-grid-width"
        columnClassName="mdc-layout-grid__cell mdc-layout-grid__cell--span-12 masonry-grid_column">
        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              Latest Pull Requests
            </Typography>
          </div>
          <PullRequestsList
            repoAddr={repoAddr}
            pullRequests={repository.pullRequests.edges}>
            <CardActions fullBleed>
              <CardActionButton
                tag={Link}
                to={`/repos/${repoAddr.platform}/${repoAddr.owner}/${repoAddr.repo}/pulls`}
                label="Full Pull Requests List"
                trailingIcon={<FontAwesomeIcon icon={faArrowRight} />}
              />
            </CardActions>
          </PullRequestsList>
        </Card>

        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              Latest Commits
            </Typography>
          </div>
          <CommitsList
            repoAddr={repoAddr}
            repoURL={repository.source.url}
            platform={repository.providerSCM}
            commits={repository.commits?.edges}>
            <CardActions fullBleed>
              <CardActionButton
                tag={Link}
                to={`/repos/${repoAddr.platform}/${repoAddr.owner}/${repoAddr.repo}/commits`}
                label="Full Commits List"
                trailingIcon={<FontAwesomeIcon icon={faArrowRight} />}
              />
            </CardActions>
          </CommitsList>
        </Card>

        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              Commits
            </Typography>
            <Tooltip
              content="The chart shows the number of buggy commits over time in this repository."
              showArrow
              align={'bottom'}>
              <FontAwesomeIcon
                className="margin-left"
                icon={faQuestionCircle}
              />
            </Tooltip>
          </div>
          <BuggyCommitsChart timeLine={timeLine} repository={repository} />
        </Card>

        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              Activities
            </Typography>
          </div>
          <TagsChart repository={repository} />
        </Card>

        <Card outlined>
          <div style={{margin: '15px'}}>
            <Typography use="headline5" style={{marginBottom: '5px'}}>
              Change Size
            </Typography>
            <Tooltip
              content="The chart shows the size of changes in commits over time in this repository."
              showArrow
              align={'bottom'}>
              <FontAwesomeIcon
                className="margin-left"
                icon={faQuestionCircle}
              />
            </Tooltip>
          </div>
          <CommitFilesChart timeLine={timeLine} repository={repository} />
        </Card>
      </Masonry>
    </>
  );
};

interface RepositoryDashboardScreenProps extends RouteComponentProps {}

export const RepositoryDashboardScreen: React.FC<RepositoryDashboardScreenProps> = ({
  location,
  ...props
}) => {
  //fixme: we should not pass down the repoAddr, the data should be gotten from the query result
  const repoAddr: any = useParams();
  const {platform, owner, repo} = repoAddr;

  const {repository} = useLazyLoadQuery<RepositoryDashboardQuery>(
    graphql`
      query RepositoryDashboardQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $commits_count: Int
      ) {
        repository(provider: $provider, owner: $owner, name: $name) {
          status
          ...RepositoryLayout_repository
          ...RepositoryDashboard_repository
            @arguments(commits_count: $commits_count)
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      commits_count: DASHBOARD_SCREEN_PAGE_SIZE,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository) return <Page404 location={location} />;

  return (
    <RepositoryLayout repository={repository}>
      <RepositoryDashboard
        repoAddr={repoAddr}
        repository={repository}
        pageSize={DASHBOARD_SCREEN_PAGE_SIZE}
      />
    </RepositoryLayout>
  );
};

interface WidgetCardProps {
  desc: string;
  size?: 2 | 4 | 8;
}

export const WidgetCard: React.FC<WidgetCardProps> = ({
  children,
  desc,
  size = 2,
}) => {
  return (
    <GridCell span={size * 1.5} tablet={size * 2} phone={size * 6}>
      <Card style={{minHeight: '98px'}}>
        <div style={{margin: '15px'}}>
          <div>{desc}</div>
          <div style={{textAlign: 'right'}}>
            <Typography use="headline3">{children}</Typography>
          </div>
        </div>
      </Card>
    </GridCell>
  );
};

interface TagsChartProps {
  repository: RepositoryDashboard_tagsCount$key;
}

const TagsChart: React.FC<TagsChartProps> = (props) => {
  const {tagsCount} = useFragment(
    graphql`
      fragment RepositoryDashboard_tagsCount on Repository {
        tagsCount {
          nodes {
            x: tag
            y: count
          }
        }
      }
    `,
    props.repository
  );

  const datasets = useMemo(
    () => [
      {
        backgroundColor: 'rgba(81,66,129,0.8)',
        data: tagsCount?.nodes?.map((v: any) => v.y),
      },
    ],
    [tagsCount]
  );

  return (
    <HorizontalBarChart
      datasets={datasets}
      labels={tagsCount?.nodes?.map((v: any) => v.x) || []}
    />
  );
};

interface BuggyCommitsChartProps {
  repository: RepositoryDashboard_commitsOverTime$key;
  timeLine: string[];
}

const BuggyCommitsChart: React.FC<BuggyCommitsChartProps> = (props) => {
  const {buggyCommitsOverTime, commitsOverTime} = useFragment(
    graphql`
      fragment RepositoryDashboard_commitsOverTime on Repository {
        buggyCommitsOverTime {
          nodes {
            x: date
            y: count
          }
        }
        commitsOverTime {
          nodes {
            x: date
            y: count
          }
        }
      }
    `,
    props.repository
  );

  const datasets = useMemo(
    () => [
      {
        label: 'Buggy Commits',
        data: fillPointsOnTimeLine(props.timeLine, buggyCommitsOverTime?.nodes),
        backgroundColor: 'rgba(255,99,104,0.7)',
      },
      {
        label: 'All Commits',
        data: fillPointsOnTimeLine(props.timeLine, commitsOverTime?.nodes),
        backgroundColor: 'rgba(54,162,235,0.35)',
      },
    ],
    [buggyCommitsOverTime, commitsOverTime, props.timeLine]
  );

  return <LineChart unit="month" datasets={datasets} />;
};

interface CommitFilesChartProps {
  repository: RepositoryDashboard_avgCommitFilesOverTime$key;
  timeLine: string[];
}

const CommitFilesChart: React.FC<CommitFilesChartProps> = (props) => {
  const {avgCommitFilesOverTime} = useFragment(
    graphql`
      fragment RepositoryDashboard_avgCommitFilesOverTime on Repository {
        avgCommitFilesOverTime {
          nodes {
            x: date
            y: avg
          }
        }
      }
    `,
    props.repository
  );

  const datasets = useMemo(
    () => [
      {
        label: 'Average number of files',
        data: fillPointsOnTimeLine(
          props.timeLine,
          avgCommitFilesOverTime?.nodes
        ),
        backgroundColor: 'rgba(81,66,129,0.4)',
      },
    ],
    [avgCommitFilesOverTime, props.timeLine]
  );

  return <LineChart unit="month" datasets={datasets} />;
};

interface LineChartProps {
  datasets: ChartDataSets[];
  unit: TimeUnit;
}

export const LineChart: React.FC<LineChartProps> = ({datasets, unit}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartInstanceRef = useRef<Chart | null>(null);

  useEffect(() => {
    const chartInstance = chartInstanceRef.current;
    if (chartInstance) {
      if (chartInstance.options.scales?.xAxes?.[0].time) {
        chartInstance.options.scales.xAxes[0].time.unit = unit;
      }
      chartInstance.data.datasets = datasets;
      chartInstance.update();
      return;
    }

    if (!canvasRef.current) {
      throw Error('canvasRef should be assigned to an HTMLCanvasElement');
    }

    chartInstanceRef.current = new Chart(canvasRef.current, {
      type: 'line',
      data: {
        datasets,
      },
      options: {
        legend: {
          reverse: true,
          position: 'bottom',
          labels: {
            usePointStyle: true,
          },
        },
        scales: {
          yAxes: [
            {
              ticks: {
                beginAtZero: true,
              },
              afterFit: function (scaleInstance) {
                if (scaleInstance.width < 50) {
                  scaleInstance.width = 50;
                }
              },
            },
          ],
          xAxes: [
            {
              type: 'time',
              time: {
                unit,
              },
            },
          ],
        },
      },
    });
  }, [datasets, unit]);

  return <canvas ref={canvasRef} />;
};

interface HorizontalBarChartProps {
  datasets: ChartDataSets[];
  labels: string[];
}

export const HorizontalBarChart: React.FC<HorizontalBarChartProps> = ({
  datasets,
  labels,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartInstanceRef = useRef<Chart | null>(null);

  useEffect(() => {
    const chartInstance = chartInstanceRef.current;
    if (chartInstance) {
      chartInstance.data.labels = labels;
      chartInstance.data.datasets = datasets;
      chartInstance.update();
      return;
    }

    if (!canvasRef.current) {
      throw Error('canvasRef should be assigned to an HTMLCanvasElement');
    }

    chartInstanceRef.current = new Chart(canvasRef.current, {
      type: 'horizontalBar',
      data: {
        labels,
        datasets,
      },
      options: {
        legend: {
          display: false,
        },
      },
    });
  }, [labels, datasets]);

  return <canvas ref={canvasRef} />;
};
