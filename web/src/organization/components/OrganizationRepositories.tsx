import React from 'react';
import {GridCell, GridRow} from '@rmwc/grid';
import {Card} from '@rmwc/card';
import {RepositoriesList} from '../../repositories/components/RepositoriesList';
import {RouteComponentProps, useParams} from 'react-router-dom';
import graphql from 'babel-plugin-relay/macro';
import {useLazyLoadQuery} from 'react-relay/hooks';
import {OrganizationRepositoriesQuery} from './__generated__/OrganizationRepositoriesQuery.graphql';
import {OrganizationLayout} from './OrganizationLayout';
import {Page404} from '../../ui/Page404';

interface OrganizationRepositoriesProps extends RouteComponentProps {}

export const OrganizationRepositories: React.FC<OrganizationRepositoriesProps> = ({
  location,
}) => {
  const {platform, owner} = useParams<any>();

  const {organization} = useLazyLoadQuery<OrganizationRepositoriesQuery>(
    graphql`
      query OrganizationRepositoriesQuery($provider: String!, $owner: String!) {
        organization(provider: $provider, owner: $owner) {
          ...OrganizationLayout_organization
          repositories(first: 100) {
            ...RepositoriesList_repositories
          }
        }
      }
    `,
    {
      provider: platform,
      owner: owner,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!organization) return <Page404 location={location} />;

  return (
    <OrganizationLayout organization={organization}>
      <GridRow>
        <GridCell span={12}>
          <Card outlined>
            <RepositoriesList repositories={organization?.repositories} />
          </Card>
        </GridCell>
      </GridRow>
    </OrganizationLayout>
  );
};
