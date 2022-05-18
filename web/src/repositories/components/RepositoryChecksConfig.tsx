import React from 'react';
import {useFragment, useMutation} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {Card} from '@rmwc/card';
import {Switch} from '@rmwc/switch';
import {Typography} from '@rmwc/typography';
import {RepositoryChecksConfig_repository$key} from './__generated__/RepositoryChecksConfig_repository.graphql';
import styled from 'styled-components/macro';

interface ChecksConfigCardProps {
  repository: RepositoryChecksConfig_repository$key;
}

const FieldContainer = styled.div`
  padding: 1em 0 1em 2em;
`;

export const RepositoryChecksConfig: React.FC<ChecksConfigCardProps> = (
  props
) => {
  const repository = useFragment(
    graphql`
      fragment RepositoryChecksConfig_repository on Repository {
        id
        checksConfig {
          enable
        }
      }
    `,
    props.repository
  );

  const [commit, isInFlight] = useMutation(graphql`
    mutation RepositoryChecksConfigMutation(
      $id: ID!
      $checksConfig: ChecksConfigInput!
    ) {
      updateRepository(input: {id: $id, checksConfig: $checksConfig}) {
        repository {
          checksConfig {
            enable
          }
        }
      }
    }
  `);

  function handelEnableChecksSwitch(evt: React.ChangeEvent<HTMLInputElement>) {
    const data = {
      id: repository.id,
      checksConfig: {
        enable: evt.currentTarget.checked,
      },
    };

    commit({
      variables: data,
      optimisticResponse: {
        updateRepository: {
          repository: data,
        },
      },
    });
  }

  return (
    <Card outlined>
      <div style={{margin: '15px'}}>
        <Typography use="headline5" style={{marginBottom: '5px'}}>
          Checks Configuration
        </Typography>
        <FieldContainer>
          <Switch
            checked={!!repository.checksConfig?.enable}
            disabled={isInFlight}
            onChange={handelEnableChecksSwitch}
            label="Enable Checks"
          />
        </FieldContainer>
      </div>
    </Card>
  );
};
