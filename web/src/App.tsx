import React, {Suspense} from 'react';
import {BrowserRouter, Redirect, Route, Switch} from 'react-router-dom';

import './App.scss';

import {Page404} from './ui/Page404';

import {library} from '@fortawesome/fontawesome-svg-core';
import {
  faBitbucket,
  faGithub,
  faGithubSquare,
  faGitlab,
  faJira,
} from '@fortawesome/free-brands-svg-icons';
import AllRepositoriesScreen from './repositories/components/AllRepositoriesScreen';
import {RepositoryDashboardScreen} from './repositories/components/RepositoryDashboard';
import {RepositoryCommitsScreen} from './repositories/components/RepositoryCommits';
import Layout, {ErrorPage, PageSpinner} from './ui/Layout';
import {RepositoryPullRequestsScreen} from './repositories/components/RepositoryPullRequests';
import {RepositorySettingsScreen} from './repositories/components/RepositorySettings';
import {CommitDetailsScreen} from './repositories/components/CommitDetails';
import {PullRequestDetailsScreen} from './repositories/components/PullRequestDetails';
import {ActivityDashboardScreen} from './administration/components/ActivityDashboard';
import {Helmet} from 'react-helmet';
import {RelayEnvironmentProvider} from 'react-relay/hooks';
import RelayEnvironment from './graphql/RelayEnvironment';
import {Provider} from 'react-redux';
import {store} from './store';
import ProviderLogin from './account/components/LoginProvider';
import LoginScreen from './account/components/LoginScreen';
import PrivateRoute from './account/components/PrivateRoute';
import {Portal} from '@rmwc/base';
import {SnackbarQueue} from '@rmwc/snackbar';
import {messages} from './ui/snackbar';
import {ErrorBoundaryWithRetry} from './ui/ErrorBoundary';
import {dialogs} from './ui/dialogs';
import {DialogQueue} from '@rmwc/dialog';
import {FeedbackScreen} from './administration/components/Feedback';
import {OrganizationRepositories} from './organization/components/OrganizationRepositories';
import {OrganizationSettings} from './organization/components/OrganizationSettings';
import {ConfigureIntegration} from './organization/components/ConfigureIntegration';
import {AddJiraIntegration} from './organization/components/AddJiraIntegration';

library.add(faGithub, faGithubSquare, faBitbucket, faGitlab, faJira);

const Home = () => {
  return <Redirect to={{pathname: '/myrepos'}} />;
};

const PrivateApp: React.FC = () => {
  return (
    <Suspense
      fallback={
        <Layout>
          <PageSpinner />
        </Layout>
      }>
      <Helmet titleTemplate="%s - Repofuel" defaultTitle="Repofuel" />
      <Switch>
        <Route exact path="/" component={Home} />
        <Route exact path="/repos" component={AllRepositoriesScreen} />
        <Route exact path="/myrepos" component={AllRepositoriesScreen} />
        <Route
          exact
          path="/repos/:platform/:owner/:repo"
          component={RepositoryDashboardScreen}
        />
        <Route
          exact
          path="/repos/:platform/:owner/:repo/commits"
          component={RepositoryCommitsScreen}
        />
        <Route
          exact
          path="/repos/:platform/:owner/:repo/commits/:hash"
          component={CommitDetailsScreen}
        />
        <Route
          exact
          path="/repos/:platform/:owner/:repo/pulls"
          component={RepositoryPullRequestsScreen}
        />
        <Route
          exact
          path="/repos/:platform/:owner/:repo/pulls/:number"
          component={PullRequestDetailsScreen}
        />
        <Route
          exact
          path="/repos/:platform/:owner/:repo/settings"
          component={RepositorySettingsScreen}
        />
        <Route
          exact
          path="/orgs/:platform/:owner"
          component={OrganizationRepositories}
        />
        <Route
          exact
          path="/orgs/:platform/:owner/settings"
          component={OrganizationSettings}
        />
        <Route
          exact
          path="/orgs/:platform/:owner/settings/integrations"
          component={ConfigureIntegration}
        />
        <Route
          exact
          path="/orgs/:platform/:owner/integrations/jira"
          component={AddJiraIntegration}
        />

        <Route path="/admin/activity" component={ActivityDashboardScreen} />
        <Route path="/admin/feedback" component={FeedbackScreen} />

        <Route component={Page404} />
      </Switch>
    </Suspense>
  );
};

const App: React.FC<any> = (props) => {
  return (
    <ErrorBoundaryWithRetry fallback={<ErrorPage />}>
      <RelayEnvironmentProvider environment={RelayEnvironment}>
        <Provider store={store}>
          <BrowserRouter>
            <Switch>
              <Route path="/login/:provider" component={ProviderLogin} />
              <Route exact path="/login" component={LoginScreen} />
              <PrivateRoute path="/" component={PrivateApp} />
            </Switch>
          </BrowserRouter>
          <Portal />
          <SnackbarQueue messages={messages} />
          <DialogQueue dialogs={dialogs.dialogs} preventOutsideDismiss />
        </Provider>
      </RelayEnvironmentProvider>
    </ErrorBoundaryWithRetry>
  );
};

export default App;
