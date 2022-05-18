import React from 'react';
import {NavLink, NavLinkProps} from 'react-router-dom';
import {ListItem, ListItemProps} from 'rmwc';

interface NavListItemProps extends ListItemProps, NavLinkProps {}

export const NavListItem: React.FC<NavListItemProps> = (props) => {
  return (
    <ListItem
      tag={NavLink}
      activeClassName="mdc-list-item--activated"
      {...props}
    />
  );
};
