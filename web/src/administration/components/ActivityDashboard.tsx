import React, {MouseEvent, useMemo, useState} from 'react';
import {RouteComponentProps} from 'react-router-dom';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Page404} from '../../ui/Page404';
import {
  ActivityDashboardQuery,
  Period,
} from './__generated__/ActivityDashboardQuery.graphql';
import Layout from '../../ui/Layout';
import {Grid, GridCell, GridRow} from '@rmwc/grid';
import {
  BreakpointColumnsObj,
  LineChart,
  WidgetCard,
} from '../../repositories/components/RepositoryDashboard';
import {DrawerContent} from '@rmwc/drawer';
import {List, ListItemText} from '@rmwc/list';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faCommentAlt} from '@fortawesome/free-regular-svg-icons';
import Masonry from 'react-masonry-css';
import {Card} from '@rmwc/card';
import {Button} from '@rmwc/button';
import {Typography} from '@rmwc/typography';
import {FormatNumber} from '../../util/number-formate';
import {ButtonGroup} from '../../ui/ButtonGroup';
import styled from 'styled-components/macro';
import {
  fillPointsOnTimeLine,
  frequencyFromPeriod,
  generateTimeLine,
} from '../../util/chart';
import {ProjectIcon} from '@primer/octicons-react';
import {NavListItem} from '../../ui/List';

export const FloatLeftButtonGroup = styled(ButtonGroup)`
  float: right;
`;

interface ActivityDashboardScreenProps extends RouteComponentProps {}

export const ActivityDashboardScreen: React.FC<ActivityDashboardScreenProps> = ({
  location,
  ...props
}) => {
  const [period, setPeriod] = useState<Period>('MONTH');
  const [frequency, timeUnit] = frequencyFromPeriod(period);

  const {activity} = useLazyLoadQuery<ActivityDashboardQuery>(
    graphql`
      query ActivityDashboardQuery($period: Period, $frequency: Frequency!) {
        activity {
          commitsAnalyzedTotalCount(period: $period)
          commitsAnalyzedCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              y: count
            }
          }
          commitsPredictTotalCount(period: $period)
          commitsPredictCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              y: count
            }
          }
          pullRequestAnalyzedTotalCount(period: $period)
          pullRequestAnalyzedCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              y: count
            }
          }
          jobsTotalCount(period: $period)
          organizationsTotalCount(period: $period)
          organizationsCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              y: count
            }
          }
          repositoriesTotalCount(period: $period)
          repositoriesCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              y: count
            }
          }
          viewsTotalCount(period: $period)
          visitorsTotalCount(period: $period)
          visitCount(period: $period, frequency: $frequency) {
            nodes {
              x: date
              views
              visitors
            }
          }
        }
      }
    `,
    {period, frequency},
    {fetchPolicy: 'store-or-network'}
  );

  const chartData = useMemo(() => {
    const timeLine = generateTimeLine(period, frequency);
    if (!activity) return null;
    return {
      traffic: [
        {
          label: 'Page Views',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.visitCount?.nodes,
            'views'
          ),
          backgroundColor: 'rgba(81,66,129,0.4)',
        },
        {
          label: 'Visitors',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.visitCount?.nodes,
            'visitors'
          ),
          backgroundColor: 'rgba(54,162,235,0.35)',
        },
      ],
      commits: [
        {
          label: 'Analyzed Commits',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.commitsAnalyzedCount?.nodes
          ),
          backgroundColor: 'rgba(81,66,129,0.4)',
        },
        {
          label: 'Commits Predicted for',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.commitsPredictCount?.nodes
          ),
          backgroundColor: 'rgba(54,162,235,0.35)',
        },
      ],
      pullRequests: [
        {
          label: 'Pull Requests',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.pullRequestAnalyzedCount?.nodes
          ),
          backgroundColor: 'rgba(81,66,129,0.4)',
        },
      ],
      repositories: [
        {
          label: 'Repositories',
          data: fillPointsOnTimeLine(
            timeLine,
            activity.repositoriesCount?.nodes
          ),
          backgroundColor: 'rgba(81,66,129,0.4)',
        },
      ],
    };
  }, [period, frequency, activity]);

  if (!chartData || !activity) return <Page404 location={location} />;

  function handelSelectTimeRange(evt: MouseEvent<HTMLButtonElement>) {
    setPeriod(evt.currentTarget.dataset.range as Period);
  }

  return (
    <Layout menuItems={<AdminMenu />}>
      <Grid className="fixed-width-layout">
        <GridRow className="masonry-grid-width">
          <GridCell span={12}>
            {/*todo: improve the styling*/}
            <div
              style={{display: 'inline-block'}}
              className="mdc-typography--headline4 margin-bottom">
              Activity Dashboard
            </div>
            <FloatLeftButtonGroup>
              <Button
                outlined
                data-range="WEEK"
                unelevated={period === 'WEEK'}
                onClick={handelSelectTimeRange}>
                Week
              </Button>
              <Button
                outlined
                data-range="MONTH"
                unelevated={period === 'MONTH'}
                onClick={handelSelectTimeRange}>
                Month
              </Button>
              <Button
                outlined
                data-range="YEAR"
                unelevated={period === 'YEAR'}
                onClick={handelSelectTimeRange}>
                Year
              </Button>
              {/*<Button*/}
              {/*    outlined data-range="ALL_TIME"*/}
              {/*    unelevated={period === "ALL_TIME"}*/}
              {/*    onClick={handelSelectTimeRange}>*/}
              {/*    All Time*/}
              {/*</Button>*/}
            </FloatLeftButtonGroup>
          </GridCell>
          <WidgetCard desc="Visitors">
            {FormatNumber(activity.visitorsTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Page Views">
            {FormatNumber(activity.viewsTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Analyzed Commits">
            {FormatNumber(activity.commitsAnalyzedTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Commits Predicted for">
            {FormatNumber(activity.commitsPredictTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Pull Requests">
            {FormatNumber(activity.pullRequestAnalyzedTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Repositories">
            {FormatNumber(activity.repositoriesTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Organizations">
            {FormatNumber(activity.organizationsTotalCount)}
          </WidgetCard>
          <WidgetCard desc="Processes">
            {FormatNumber(activity.jobsTotalCount)}
          </WidgetCard>
        </GridRow>
        <Masonry
          breakpointCols={BreakpointColumnsObj}
          className="masonry-grid masonry-grid-width"
          columnClassName="mdc-layout-grid__cell mdc-layout-grid__cell--span-12 masonry-grid_column">
          <Card outlined>
            <div style={{margin: '15px'}}>
              <Typography use="headline5" style={{marginBottom: '5px'}}>
                Traffic
              </Typography>
            </div>
            <LineChart unit={timeUnit} datasets={chartData.traffic} />
          </Card>

          <Card outlined>
            <div style={{margin: '15px'}}>
              <Typography use="headline5" style={{marginBottom: '5px'}}>
                Commits
              </Typography>
            </div>
            <LineChart unit={timeUnit} datasets={chartData.commits} />
          </Card>

          <Card outlined>
            <div style={{margin: '15px'}}>
              <Typography use="headline5" style={{marginBottom: '5px'}}>
                Pull Requests
              </Typography>
            </div>
            <LineChart unit={timeUnit} datasets={chartData.pullRequests} />
          </Card>

          <Card outlined>
            <div style={{margin: '15px'}}>
              <Typography use="headline5" style={{marginBottom: '5px'}}>
                Repositories
              </Typography>
            </div>
            <LineChart unit={timeUnit} datasets={chartData.repositories} />
          </Card>
        </Masonry>
      </Grid>
    </Layout>
  );
};

interface AdminMenuProps {}

export const AdminMenu: React.FC<AdminMenuProps> = () => {
  return (
    <DrawerContent>
      <List>
        <NavListItem to={'/admin/activity'}>
          <ProjectIcon className="list-close-icon" />
          <ListItemText>Activity Dashboard</ListItemText>
        </NavListItem>
        <NavListItem to={'/admin/feedback'}>
          <FontAwesomeIcon className="list-close-icon" icon={faCommentAlt} />
          <ListItemText>Feedback</ListItemText>
        </NavListItem>
      </List>
    </DrawerContent>
  );
};
