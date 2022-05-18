import React from 'react';
import {Story, Meta} from '@storybook/react';
import {
  AddPublicRepositoryDialog,
  AddPublicRepositoryDialogProps,
} from '../repositories/components/AddPublicRepository';
import {createMockEnvironment} from 'relay-test-utils';
import {RelayEnvironmentProvider} from 'react-relay/hooks';
import '../App.scss';
import '../index.scss';

import {
  faBitbucket,
  faGithub,
  faGithubSquare,
  faGitlab,
  faJira,
} from '@fortawesome/free-brands-svg-icons';
import {library} from '@fortawesome/fontawesome-svg-core';
library.add(faGithub, faGithubSquare, faBitbucket, faGitlab, faJira);

export default {
  title: 'Components/Add Public Repository',
  component: AddPublicRepositoryDialog,
} as Meta;

const RelayMockEnvironment = createMockEnvironment();

const Template: Story<AddPublicRepositoryDialogProps> = (args) => (
  <RelayEnvironmentProvider environment={RelayMockEnvironment}>
    <AddPublicRepositoryDialog {...args} />
  </RelayEnvironmentProvider>
);

export const Github = Template.bind({});
Github.args = {
  open: true,
};
