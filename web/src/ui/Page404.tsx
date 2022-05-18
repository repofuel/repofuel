import React from 'react';

import {Link} from 'react-router-dom';
import styled from 'styled-components/macro';

interface Page404Props {
  location: {pathname: string};
}

const StyledPage404 = styled.div`
  margin-top: 50px;
  text-align: center;
`;

export const Page404: React.FC<Page404Props> = ({location}) => {
  return (
    <StyledPage404>
      <h1>404</h1>
      <p>
        No match found for <code>{location.pathname}</code>
      </p>
      <Link to="/">Repofuel - Home</Link>
    </StyledPage404>
  );
};

export const Page404Custom: React.FC = (props) => {
  return (
    <div className="page-404">
      <h1>404</h1>
      <p>{props.children}</p>
      <Link to="/">Repofuel - Home</Link>
    </div>
  );
};
