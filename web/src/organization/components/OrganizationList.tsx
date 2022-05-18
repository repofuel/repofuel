import {List, ListItem, ListItemMeta, ListItemText} from '@rmwc/list';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import React from 'react';
import {Link} from 'react-router-dom';
import {Avatar} from '@rmwc/avatar';
import {PeopleIcon, PersonIcon} from '@primer/octicons-react';
import {IconName} from '@fortawesome/free-brands-svg-icons';
import {NavListItem} from '../../ui/List';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {OrganizationListQuery} from './__generated__/OrganizationListQuery.graphql';

interface OrganizationListProps {}

export const OrganizationList: React.FC<OrganizationListProps> = () => {
  const data = useLazyLoadQuery<OrganizationListQuery>(
    graphql`
      query OrganizationListQuery {
        organizations(first: 100) {
          edges {
            node {
              id
              providerSCM
              avatarURL
              owner {
                slug
              }
            }
          }
        }
      }
    `,
    {}
  );

  return (
    <List>
      <NavListItem to="/myrepos">
        <PersonIcon className="list-close-icon" />
        <ListItemText>My Repositories</ListItemText>
      </NavListItem>
      <NavListItem to="/repos">
        <PeopleIcon className="list-close-icon" />
        <ListItemText>All Repositories</ListItemText>
      </NavListItem>

      {data?.organizations.edges?.map((edge) => {
        if (!edge?.node) return null;

        const org = edge.node;
        return (
          <ListItem
            tag={Link}
            key={org.id}
            to={`/orgs/${org.providerSCM}/${org.owner.slug}`}>
            <Avatar
              square
              className="margin-right"
              src={org.avatarURL || undefined}
            />
            <ListItemText>{org.owner.slug}</ListItemText>
            <ListItemMeta>
              <FontAwesomeIcon icon={['fab', systemIcon(org.providerSCM)]} />
            </ListItemMeta>
          </ListItem>
        );
      })}
    </List>
  );
};

function systemIcon(system: string): IconName {
  switch (system) {
    case 'Github':
    case 'github':
      return 'github';
    default:
      return 'server'; //fixme
  }
}
