import React, {Suspense} from 'react';
import {Grid, GridCell} from '@rmwc/grid';
import Layout, {PageSpinner} from '../../ui/Layout';
import {DrawerContent} from '@rmwc/drawer';
import {List, ListItem, ListItemText} from '@rmwc/list';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import graphql from 'babel-plugin-relay/macro';
import {Button} from '@rmwc/button';
import {GearIcon, RepoIcon} from '@primer/octicons-react';
import {useFragment} from 'react-relay/hooks';
import {OrganizationLayout_organization$key} from './__generated__/OrganizationLayout_organization.graphql';
import {OrganizationLayout_Menu_organization$key} from './__generated__/OrganizationLayout_Menu_organization.graphql';
import {Avatar, DrawerHeader, DrawerTitle} from 'rmwc';
import {faJira} from '@fortawesome/free-brands-svg-icons';
import {NavListItem} from '../../ui/List';

interface OrganizationScreenProps {
  organization: OrganizationLayout_organization$key;
}

export const OrganizationLayout: React.FC<OrganizationScreenProps> = (
  props
) => {
  const organization = useFragment(
    graphql`
      fragment OrganizationLayout_organization on Organization {
        id
        owner {
          slug
        }
        providerSCM
        viewerCanAdminister
        providerSetupURL

        ...OrganizationLayout_Menu_organization
      }
    `,
    props.organization
  );

  return (
    <Layout menuItems={<OrganizationMenu organization={organization} />}>
      <Grid className="fixed-width-layout">
        <GridCell span={12}>
          <div className="repository-head">
            <span className="mdc-typography--headline5">
              <FontAwesomeIcon
                className="margin-right"
                icon={['fab', organization.providerSCM as any]}
              />
              {organization.owner.slug}
            </span>{' '}
            {organization.viewerCanAdminister &&
              organization.providerSetupURL && (
                <Button
                  tag="a"
                  label={'Configure on ' + organization.providerSCM}
                  icon={<GearIcon />}
                  unelevated
                  className="float-right"
                  href={organization.providerSetupURL}
                  target="_blank"
                  rel="nofollow"
                />
              )}
          </div>
        </GridCell>
        <GridCell span={12}>
          <Suspense fallback={<PageSpinner />}>{props.children}</Suspense>
        </GridCell>
      </Grid>
    </Layout>
  );
};

interface OrganizationMenuProps {
  organization: OrganizationLayout_Menu_organization$key;
}

const OrganizationMenu: React.FC<OrganizationMenuProps> = (props) => {
  const organization = useFragment(
    graphql`
      fragment OrganizationLayout_Menu_organization on Organization {
        owner {
          slug
        }
        providerSCM
        avatarURL
      }
    `,
    props.organization
  );

  const orgPath = `/orgs/${organization.providerSCM}/${organization.owner.slug}`;

  return (
    <>
      <DrawerHeader>
        <DrawerTitle>
          <Avatar
            name={organization.owner.slug}
            src={organization.avatarURL || undefined}
            className="margin-right"
          />
          {organization.owner.slug}
        </DrawerTitle>
      </DrawerHeader>
      <DrawerContent>
        <List>
          <NavListItem to={orgPath}>
            <RepoIcon className="list-close-icon" />
            <ListItemText>Repositories</ListItemText>
          </NavListItem>
          <ListItem
            disabled
            //tag={Link}
            // to={orgPath + '/integrations/jira'}
          >
            <FontAwesomeIcon className="list-close-icon" icon={faJira} />
            <ListItemText>Add Jira Integration</ListItemText>
          </ListItem>
          <ListItem
            disabled
            // tag={Link}
            // to={orgPath + '/settings'}
          >
            <GearIcon className="list-close-icon" />
            <ListItemText>Settings</ListItemText>
          </ListItem>
        </List>
      </DrawerContent>
    </>
  );
};
