import React from 'react';
import {GridCell, GridRow} from '@rmwc/grid';
import {Organization} from '../types';
import {Redirect} from 'react-router-dom';

interface OrganizationSettingsProps {
  org: Organization;
}

export const OrganizationSettings: React.FC<OrganizationSettingsProps> = ({
  org,
}) => {
  // not content now, sow temporary we redirect to the integrations
  return (
    <GridRow>
      <GridCell span={12}>
        <Redirect
          to={`/orgs/${org.provider_scm}/${org.owner.slug}/settings/integrations`}
        />
      </GridCell>
    </GridRow>
  );
};
