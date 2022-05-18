import React from 'react';
import {Link, RouteComponentProps} from 'react-router-dom';
import {useLazyLoadQuery} from 'react-relay/lib/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Page404} from '../../ui/Page404';

import Layout from '../../ui/Layout';
import {Grid, GridCell, GridRow} from '@rmwc/grid';
import {
  Button,
  DataTable,
  DataTableBody,
  DataTableCell,
  DataTableContent,
  DataTableHead,
  DataTableHeadCell,
  DataTableRow,
} from 'rmwc';
import {
  FeedbackQuery,
  FeedbackQueryResponse,
} from './__generated__/FeedbackQuery.graphql';
import {AdminMenu, FloatLeftButtonGroup} from './ActivityDashboard';
import {authenticateDownload} from '../../account/token';
import {Blankslate} from '../../repositories/components/CommitsList';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faCommentAlt} from '@fortawesome/free-regular-svg-icons';
import styled from 'styled-components';

interface FeedbackScreenProps extends RouteComponentProps {}

export const FeedbackScreen: React.FC<FeedbackScreenProps> = ({
  location,
  ...props
}) => {
  const {feedback} = useLazyLoadQuery<FeedbackQuery>(
    graphql`
      query FeedbackQuery {
        feedback(first: 100) {
          edges {
            node {
              sender {
                id
              }
              target {
                hash
                repository {
                  name
                  providerSCM
                  owner {
                    slug
                  }
                }
              }
              message
              createdAt
            }
          }
        }
      }
    `,
    {},
    {fetchPolicy: 'store-and-network'}
  );

  if (!feedback) return <Page404 location={location} />;

  return (
    <Layout menuItems={<AdminMenu />}>
      <Grid>
        <GridRow>
          <GridCell span={12}>
            <div
              style={{display: 'inline-block'}}
              className="mdc-typography--headline4 margin-bottom">
              Feedback
            </div>
            <FloatLeftButtonGroup>
              <Button
                outlined
                onClick={authenticateDownload('/ingest/download/feedback.csv')}>
                Download
              </Button>
            </FloatLeftButtonGroup>
          </GridCell>
          <GridCell span={12}>
            <FeedbackTable messages={feedback.edges} />
          </GridCell>
        </GridRow>
      </Grid>
    </Layout>
  );
};

type FeedbackResponse = NonNullable<FeedbackQueryResponse['feedback']>;

interface FeedbackTableProps {
  messages: FeedbackResponse['edges'];
}

const FullWidthDataTable = styled(DataTable)`
  width: 100%;

  td:last-child {
    width: 100%;
    white-space: normal;
    min-width: 400px;
  }
`;

const FeedbackTable: React.FC<FeedbackTableProps> = ({messages}) => {
  if (!messages?.length) {
    return (
      <Blankslate>
        {/* <GitCommitIcon size={'large'} /> */}
        <FontAwesomeIcon icon={faCommentAlt} size="5x" />
        <h3>There aren’t any feedback.</h3>
        <p>Feedback form developers will showup here.</p>
      </Blankslate>
    );
  }

  return (
    <FullWidthDataTable>
      <DataTableContent>
        <DataTableHead>
          <DataTableRow>
            <DataTableHeadCell>Repository</DataTableHeadCell>
            <DataTableHeadCell>Commit</DataTableHeadCell>
            {/* <DataTableHeadCell>Sender ID</DataTableHeadCell> */}
            <DataTableHeadCell>Date</DataTableHeadCell>
            <DataTableHeadCell>Message</DataTableHeadCell>
          </DataTableRow>
        </DataTableHead>
        <DataTableBody>
          {messages?.map(
            (edge) =>
              edge && edge.node && <FeedbackTableRow message={edge.node} />
          )}
        </DataTableBody>
      </DataTableContent>
    </FullWidthDataTable>
  );
};

type ArrayElement<ArrayType extends readonly unknown[]> = NonNullable<
  ArrayType extends readonly (infer ElementType)[] ? ElementType : never
>;

type FeedbackEdge = ArrayElement<NonNullable<FeedbackResponse['edges']>>;

interface FeedbackTableRowProps {
  message: NonNullable<FeedbackEdge['node']>;
}

const FeedbackTableRow: React.FC<FeedbackTableRowProps> = ({message}) => {
  if (!message.target) {
    return (
      <DataTableRow>
        <DataTableCell>Not Available</DataTableCell>
        <DataTableCell>-</DataTableCell>
        <DataTableCell>{message.createdAt}</DataTableCell>
        <DataTableCell>{message.message}</DataTableCell>
      </DataTableRow>
    );
  }

  const repo = message.target.repository;
  const repoHref = `/repos/${repo.providerSCM}/${repo.owner.slug}/${repo.name}`;

  return (
    <DataTableRow>
      <DataTableCell>
        <Link className="link" to={repoHref}>
          {message.target.repository.owner.slug}/
          {message.target.repository.name}
        </Link>
      </DataTableCell>
      <DataTableCell>
        <Link
          className="link"
          to={`${repoHref}/commits/${message.target.hash}`}>
          {message.target.hash.slice(0, 7)}
        </Link>
      </DataTableCell>
      <DataTableCell>{message.createdAt}</DataTableCell>
      <DataTableCell>{message.message}</DataTableCell>
    </DataTableRow>
  );
};
