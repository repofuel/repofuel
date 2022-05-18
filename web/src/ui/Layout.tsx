import React, {Suspense, useState} from 'react';
import {Link} from 'react-router-dom';
import styled from 'styled-components/macro';
import Skeleton from 'react-loading-skeleton';
import {
  TopAppBar,
  TopAppBarNavigationIcon,
  TopAppBarRow,
  TopAppBarSection,
  TopAppBarTitle,
} from '@rmwc/top-app-bar';
import {Drawer, DrawerAppContent} from '@rmwc/drawer';
import {Button} from '@rmwc/button';
import {Typography} from '@rmwc/typography';
import {CircularProgress} from '@rmwc/circular-progress';
import {ProfileMenu} from '../account/components/ProfileMenu';
import '@rmwc/circular-progress/styles';
import './Logo.scss';
import {ThreeBarsIcon} from '@primer/octicons-react';

const Content = styled.div`
  flex: 1 0 auto;
`;

const Footer = styled.footer`
  flex-shrink: 0;
  padding: 25px 24px 10px;
  text-align: center;
`;

const Layout: React.FC<any> = ({children, menuItems}) => {
  const [isSmallScreen, resize] = useState(function () {
    return window.innerWidth < 1280;
  });
  const [open, setOpen] = useState(!isSmallScreen);

  window.addEventListener('resize', function () {
    resize(window.innerWidth < 1280);
  });

  return (
    <div className="drawer-container">
      <TopAppBar>
        <TopAppBarRow>
          <TopAppBarSection alignStart>
            <TopAppBarNavigationIcon
              checked={open}
              icon={<ThreeBarsIcon />}
              onClick={() => setOpen(!open)}
            />
            <TopAppBarTitle>
              <Link className="repofuel-logo" to="/" />
            </TopAppBarTitle>
          </TopAppBarSection>
          <TopAppBarSection alignEnd>
            <Suspense
              fallback={<Skeleton circle={true} height={40} width={40} />}>
              <ProfileMenu />
            </Suspense>
          </TopAppBarSection>
        </TopAppBarRow>
      </TopAppBar>

      <Drawer
        onClose={() => setOpen(false)}
        className="mdc-top-app-bar--fixed-adjust"
        modal={isSmallScreen || !open}
        open={open}>
        {menuItems}
      </Drawer>

      <DrawerAppContent className="drawer-app-content mdc-top-app-bar--fixed-adjust">
        <Content>{children}</Content>
        <Footer>
          <Typography use="subtitle2">
            © 2020 Repofuel v{process.env.REACT_APP_VERSION}
          </Typography>
        </Footer>
      </DrawerAppContent>
    </div>
  );
};

export const PageSpinner: React.FC = () => {
  return (
    <CircularProgress
      style={{
        position: 'absolute',
        left: '50%',
        right: '50%',
        marginTop: '20%',
        width: '80px',
        height: '80px',
      }}
    />
  );
};

const CenterContent = styled.div`
  margin-top: 10%;
  text-align: center;
  padding: 24px;
`;

interface ErrorPageProps {
  retry?: () => void;
}

export const ErrorPage: React.FC<ErrorPageProps> = (props) => {
  return (
    <CenterContent>
      <h1>Something went wrong.</h1>
      <p>
        Sorry about that. Please try refreshing and contact us if the problem
        persists.
      </p>
      {props.children}
      <Button
        raised
        onClick={
          props.retry
            ? props.retry
            : window.location.reload.bind(window.location)
        }>
        Refresh
      </Button>
    </CenterContent>
  );
};

export default Layout;
