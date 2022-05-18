import React from 'react';
import './UI.scss';

export const SectionHeader: React.FC<any> = ({children}) => {
  return <h2 className="section__header">{children}</h2>;
};

export const SubsectionHeader: React.FC = ({children}) => {
  return <h6 className="subsection__header">{children}</h6>;
};

export const SectionBody: React.FC<any> = ({children}) => {
  return <div className="section__body">{children}</div>;
};

export const CardHeader: React.FC<any> = ({children}) => {
  return <h3 className="card__header">{children}</h3>;
};

export const CardBody: React.FC<any> = ({children}) => {
  return <div className="card__body">{children}</div>;
};

export const FormActions: React.FC<any> = ({children}) => {
  return <div className="form__actions">{children}</div>;
};

export const TextFieldContainer: React.FC<any> = ({children}) => {
  return <div className="text-field-container">{children}</div>;
};

export const RadioContainer: React.FC<any> = ({children}) => {
  return <div className="radio-container">{children}</div>;
};
