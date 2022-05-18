import React, {useState} from 'react';
import {Avatar} from '@rmwc/avatar';
import {MenuItem, MenuSurface, MenuSurfaceAnchor} from '@rmwc/menu';
import {
  ListDivider,
  ListItem,
  ListItemSecondaryText,
  ListItemText,
} from '@rmwc/list';
import {useLazyLoadQuery} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {ProfileMenuQuery} from './__generated__/ProfileMenuQuery.graphql';
import tokenStore from '../token';

export const ProfileMenu: React.FC<any> = () => {
  const [isMenuOpen, setMenuState] = useState(false);
  const {viewer} = useLazyLoadQuery<ProfileMenuQuery>(
    graphql`
      query ProfileMenuQuery {
        viewer {
          firstName
          lastName
          avatarUrl
        }
      }
    `,
    {},
    {fetchPolicy: 'store-or-network'}
  );

  function closeMenu() {
    setMenuState(false);
  }

  function logout() {
    tokenStore.logout();
  }

  if (!viewer) return <MenuItem onClick={logout}>Sing out</MenuItem>;

  return (
    <MenuSurfaceAnchor>
      <Avatar
        src={viewer.avatarUrl || undefined}
        size="large"
        name={viewer.firstName + ' ' + viewer.lastName}
        onClick={() => setMenuState(true)}
        interactive
      />
      <MenuSurface
        open={isMenuOpen}
        onClose={closeMenu}
        anchorCorner="bottomLeft">
        <ListItem onClick={closeMenu}>
          <ListItemText>
            <ListItemSecondaryText>Signed in as</ListItemSecondaryText>
            {viewer.firstName + ' ' + viewer.lastName}
          </ListItemText>
        </ListItem>
        <ListDivider />
        <MenuItem onClick={logout}>Sign out</MenuItem>
      </MenuSurface>
    </MenuSurfaceAnchor>
  );
};
