import React, {useState} from 'react';
import {
  CardBody,
  CardHeader,
  FormActions,
  RadioContainer,
  TextFieldContainer,
} from '../../ui/UI';
import {TextField} from '@rmwc/textfield';
import {Organization} from '../types';
import {useForm} from 'react-hook-form';
import {useDispatch} from 'react-redux';
import {
  CheckJiraUrl,
  fetchOrganization,
  GetOAuthUrl,
  SendConditionals,
} from '../actions';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faCheckCircle} from '@fortawesome/free-regular-svg-icons';
import {Button} from '@rmwc/button';
import {
  faCircleNotch,
  faExternalLinkAlt,
} from '@fortawesome/free-solid-svg-icons';
import {Radio} from '@rmwc/radio';
import {useHistory} from 'react-router-dom';
import {GridCell, GridRow} from '@rmwc/grid';
import {Card} from '@rmwc/card';

const AUTH_METHOD_BASIC = '1';
const AUTH_METHOD_OAUTH = '2';

const handelOAuthFlow = (data: {oauth_url: string}) => {
  const url = new URL(data.oauth_url);
  const win = window.open(
    data.oauth_url,
    'oauth_' + url.host,
    'toolbar=no,directories=no,status=no,menubar=no,scrollbars=no,resizable=no,modal=yes'
  );
  if (!win) {
    return Promise.reject({message: 'Cannot open the authorization windows'});
  }

  return new Promise((resolve, reject) => {
    const id = setInterval(function () {
      if (win.closed) {
        clearInterval(id);
        //todo: linking process has been finished, refresh and check
        resolve('finished');
      }
    }, 1000);
  });
};

const selectAuthMethod = (data: any) => {
  const keys = Object.keys(data.auth_methods);
  if (keys.length === 1) {
    data.auth_method = keys[0];
  }
  return data;
};

interface JiraIntegrationProps {
  org: Organization;
}

export const AddJiraIntegration: React.FC<JiraIntegrationProps> = ({org}) => {
  return (
    <GridCell span={12}>
      <Card outlined>
        <GridRow>
          <GridCell span={7}>
            <CardHeader>Add Jira Integration</CardHeader>
            <CardBody>
              <AddJiraIntegrationForm org={org} />
            </CardBody>
          </GridCell>
        </GridRow>
      </Card>
    </GridCell>
  );
};

const AddJiraIntegrationForm: React.FC<JiraIntegrationProps> = ({org}) => {
  const {handleSubmit, register, errors} = useForm();
  const dispatch = useDispatch<any>();
  const [form, setFromValues] = useState<any>({});
  const history = useHistory();

  //todo: 1) redirect to the integration itself, 2) handel if we have a slash in the end of the path
  const successful = () =>
    dispatch(fetchOrganization(org.provider_scm, org.owner.slug)).then(() =>
      history.push('../settings/integrations')
    );

  const fail = (err: any) => setFromValues({...form, isLoading: false}); //todo: should show error message

  const onSubmitURL = (values: any) => {
    setFromValues({...form, isLoading: true});

    if (!form.valid_server) {
      dispatch(CheckJiraUrl(values.base_url))
        .then(selectAuthMethod)
        .then(setFromValues)
        .catch(fail);
    } else if (form.auth_method === AUTH_METHOD_BASIC) {
      dispatch(SendConditionals(form.provider, org.id, values))
        .then(successful)
        .catch(fail);
    } else if (form.auth_method === AUTH_METHOD_OAUTH) {
      dispatch(GetOAuthUrl(form.provider, org.id))
        .then(handelOAuthFlow)
        .then(successful)
        .catch(fail);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmitURL)}>
      <p>
        Enter your Jira base URL below to connect Repofuel to your Jira account.
      </p>
      <TextFieldContainer>
        <TextField
          outlined
          label="Jira Base URL"
          name="base_url"
          placeholder="e.g., https://issues.redhat.com"
          inputRef={register({
            pattern: {
              value: /http(s)?:\/\/.[-a-zA-Z0-9@:%._+~#=]{2,256}\/([-a-zA-Z0-9@:%_+.~#?&/=]*)/,
              message: 'Invalid URL',
            },
            required: 'Base URL is required',
          })}
          invalid={errors.base_url || form.valid_server === false}
          disabled={form.valid_server || form.isLoading}
          helpText={{
            persistent: form.server_title,
            validationMsg: true,
            children:
              form.server_title ||
              errors.base_url?.message ||
              'Cannot reach the Jira endpoint',
          }}
          trailingIcon={
            form.valid_server && <FontAwesomeIcon icon={faCheckCircle} />
          }
        />
      </TextFieldContainer>

      {!form.valid_server && (
        <FormActions>
          <Button
            raised
            disabled={form.isLoading}
            trailingIcon={
              form.isLoading && (
                <FontAwesomeIcon spin={true} icon={faCircleNotch} />
              )
            }>
            Next
          </Button>
        </FormActions>
      )}

      {Object.keys(form.auth_methods || {}).length > 1 && (
        <>
          <p>
            Select the authentication method that you prefer.
            <br />
            <small>*Recommend OAuth</small>
          </p>
          {Object.keys(form.auth_methods).map((auth_method: string) => (
            <RadioContainer key={auth_method}>
              <Radio
                name="auth_method"
                value={auth_method}
                label={form.auth_methods[auth_method]}
                inputRef={register({required: true})}
                onClick={() => setFromValues({...form, auth_method})}
              />
            </RadioContainer>
          ))}
        </>
      )}

      {form.auth_method === AUTH_METHOD_BASIC && (
        <>
          <p>
            Enter the account credential that Repofuel should use to interact
            with Jira.
          </p>
          <CredentialsInput register={register} errors={errors} />
          <FormActions>
            <Button
              raised
              disabled={form.isLoading}
              trailingIcon={
                form.isLoading && (
                  <FontAwesomeIcon spin={true} icon={faCircleNotch} />
                )
              }>
              Save
            </Button>
          </FormActions>
        </>
      )}

      {form.auth_method === AUTH_METHOD_OAUTH && (
        <FormActions>
          <Button
            raised
            disabled={form.isLoading}
            trailingIcon={
              form.isLoading ? (
                <FontAwesomeIcon spin={true} icon={faCircleNotch} />
              ) : (
                <FontAwesomeIcon icon={faExternalLinkAlt} />
              )
            }>
            Link to Jira
          </Button>
        </FormActions>
      )}
    </form>
  );
};

const CredentialsInput: React.FC<any> = ({register, errors}) => {
  return (
    <>
      <TextFieldContainer>
        <TextField
          outlined
          label="Username"
          name="user"
          className="full-width"
          placeholder="e.g., emadshihab"
          inputRef={register({required: true})}
          invalid={errors.user}
        />
      </TextFieldContainer>
      <TextFieldContainer>
        <TextField
          outlined
          label="Password"
          name="pass"
          placeholder="e.g., AsxDFkfw23pCnf2sa"
          type="password"
          inputRef={register({required: true})}
          invalid={errors.pass}
        />
      </TextFieldContainer>
    </>
  );
};
