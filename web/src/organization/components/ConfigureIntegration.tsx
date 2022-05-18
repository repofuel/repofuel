import React from 'react';
import {Organization} from '../types';
import {CardBody, CardHeader, SectionHeader} from '../../ui/UI';
import ReactJson from 'react-json-view';
import {Card} from '@rmwc/card';
import {GridCell, GridRow} from '@rmwc/grid';

interface ConfigureIntegrationProps {
  org: Organization;
}

export const ConfigureIntegration: React.FC<ConfigureIntegrationProps> = ({
  org,
}) => {
  return (
    <>
      <SectionHeader>Integrations</SectionHeader>
      <GridRow>
        <GridCell span={12}>
          <Card outlined>
            <GridRow>
              <GridCell span={12}>
                <CardHeader>Currant integrations</CardHeader>
                <CardBody>
                  <ReactJson
                    name="integrations"
                    displayDataTypes={false}
                    displayObjectSize={false}
                    style={{marginLeft: '1em'}}
                    src={org.config}
                  />
                </CardBody>
              </GridCell>
            </GridRow>
          </Card>
        </GridCell>
      </GridRow>
    </>
  );
};
