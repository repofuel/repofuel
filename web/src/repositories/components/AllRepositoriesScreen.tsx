import React from 'react';
import {Grid, GridCell} from '@rmwc/grid';
import {Button} from '@rmwc/button';
import {RepositoriesList} from './RepositoriesList';
import Layout from '../../ui/Layout';
import {DrawerContent, DrawerHeader, DrawerTitle} from '@rmwc/drawer';
import {OrganizationList} from '../../organization/components/OrganizationList';
import {Avatar} from '@rmwc/avatar';
import {Card} from '@rmwc/card';
import {List, ListDivider, ListItem, ListItemText} from '@rmwc/list';
import {Link, RouteComponentProps} from 'react-router-dom';
import {PlusIcon, ProjectIcon} from '@primer/octicons-react';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {AllRepositoriesScreenQuery} from './__generated__/AllRepositoriesScreenQuery.graphql';
import {AddPublicRepository} from './AddPublicRepository';
import {ButtonGroup} from '../../ui/ButtonGroup';

interface AllRepositoriesScreenProps extends RouteComponentProps {}

export const AllRepositoriesScreen: React.FC<AllRepositoriesScreenProps> = ({
  match,
}) => {
  let {viewer, repositories} = useLazyLoadQuery<AllRepositoriesScreenQuery>(
    graphql`
      query AllRepositoriesScreenQuery($viewerRepos: Boolean!) {
        viewer {
          firstName
          lastName
          avatarUrl
          role
          repositories(first: 100) @include(if: $viewerRepos) {
            ...RepositoriesList_repositories
          }
        }
        repositories(first: 100) @skip(if: $viewerRepos) {
          ...RepositoriesList_repositories
        }
      }
    `,
    {
      viewerRepos: match.path === '/myrepos',
    },
    {fetchPolicy: 'store-and-network'}
  );

  return (
    <Layout menuItems={<SharedAccountsMenu viewer={viewer} />}>
      <Grid className="fixed-width-layout">
        <GridCell span={12} style={{textAlign: 'right'}}>
          <ButtonGroup className="float-right">
            <Button
              unelevated
              tag="a"
              label="Add From Github"
              icon={<PlusIcon />}
              href="/accounts/apps/github/add_repository"
              target="_blank"
              rel="nofollow"
            />
            <AddPublicRepository />
          </ButtonGroup>
        </GridCell>
        <GridCell span={12}>
          <Card outlined>
            <RepositoriesList
              repositories={repositories || viewer.repositories}
            />
          </Card>
        </GridCell>
      </Grid>
    </Layout>
  );
};

const SharedAccountsMenu: React.FC<any> = ({viewer}) => {
  const fullName = viewer?.firstName + ' ' + viewer?.lastName;
  return (
    <>
      <DrawerHeader>
        <DrawerTitle>
          <Avatar
            name={fullName}
            src={viewer?.avatarUrl || undefined}
            className="margin-right"
          />
          {fullName}
        </DrawerTitle>
      </DrawerHeader>
      <DrawerContent>
        <OrganizationList />
      </DrawerContent>
      {viewer?.role === 'SITE_ADMIN' && (
        <List>
          <ListDivider />
          <ListItem tag={Link} to={'/admin/activity'}>
            <ProjectIcon className="list-close-icon" />
            <ListItemText>Activity Dashboard</ListItemText>
          </ListItem>
        </List>
      )}
    </>
  );
};

export default AllRepositoriesScreen;
