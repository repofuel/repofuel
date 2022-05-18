import React, {useState} from 'react';

import {LinearProgress} from '@rmwc/linear-progress';
import {Card} from '@rmwc/card';
import {TextField} from '@rmwc/textfield';
import {Button} from '@rmwc/button';
import {Typography} from '@rmwc/typography';

import './LoginScreen.scss';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import styled from 'styled-components/macro';
import {Link} from 'react-router-dom';

interface LoginScreenProps {
  isFetching: boolean;
}

const LinedWithOr: React.FC = (props) => {
  return (
    <div {...props}>
      <div />
      <span>OR</span>
    </div>
  );
};

const StyledLinedWithOr = styled(LinedWithOr)`
  text-align: center;
  margin: 1rem 0;

  & > div {
    background: rgb(221, 221, 221);
    width: 100%;
    height: 1px;
    position: relative;
    top: 1rem;
  }
  & > span {
    line-height: 1.8;
    background-color: white;
    color: rgb(187, 187, 187);
    padding: 0 1rem;
    position: relative;
  }
`;

const TypographyWithMargins = styled((props) => <Typography {...props} />)`
  margin: 2rem 0;
`;

export const LoginScreen: React.FC<LoginScreenProps> = ({isFetching}) => {
  return (
    <div className={'login-page'}>
      <LinearProgress closed={!isFetching} />

      <Card className="login-card">
        <div>
          {/*<FontAwesomeIcon size="5x" icon={faUserCircle}/>*/}
          <Link className="repofuel-logo" to={'/'}>
            Repofuel
          </Link>
        </div>

        <TypographyWithMargins use="headline5">Sign in</TypographyWithMargins>

        <Button
          tag={'a'}
          disabled={isFetching}
          outlined
          href="/accounts/login/github"
          rel="nofollow"
          icon={<FontAwesomeIcon icon={['fab', 'github']} />}>
          Continue with Github
        </Button>

        <Button
          disabled={isFetching || true}
          outlined
          // href="/accounts/login/bitbucket"
          rel="nofollow"
          icon={<FontAwesomeIcon icon={['fab', 'bitbucket']} />}>
          Continue with Bitbucket
        </Button>

        <Button
          disabled={isFetching || true}
          outlined
          // href="/accounts/login/gitlab"
          rel="nofollow"
          icon={<FontAwesomeIcon icon={['fab', 'gitlab']} />}>
          Continue with Gitlab
        </Button>
      </Card>
    </div>
  );
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const LoginWithUsernameAndPassword: React.FC<{isFetching?: boolean}> = ({
  isFetching,
}) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  return (
    <>
      <StyledLinedWithOr />

      <TextField
        disabled={isFetching}
        label="Username"
        value={username}
        onChange={(e: any) => setUsername(e.currentTarget.value)}
      />
      <TextField
        disabled={isFetching}
        label="Password"
        type={'password'}
        value={password}
        onChange={(e: any) => setPassword(e.currentTarget.value)}
      />
      <Button raised disabled={isFetching || true}>
        Sign in
      </Button>
    </>
  );
};

export default LoginScreen;
