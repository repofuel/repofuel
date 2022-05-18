import React from 'react';

interface CellProps {
  numeric?: boolean;
}

export const DataTable: React.FC = (props) => {
  return (
    <div className="mdc-data-table">
      <table className="mdc-data-table__table">{props.children}</table>
    </div>
  );
};

export const DataTableHeader: React.FC = ({children}) => {
  return (
    <thead>
      <tr className="mdc-data-table__header-row">{children}</tr>
    </thead>
  );
};

export const DataTableHeaderCell: React.FC<CellProps> = ({
  numeric,
  children,
}) => {
  return (
    <th
      className={
        'mdc-data-table__header-cell' +
        (numeric ? ' mdc-data-table__header-cell--numeric' : '')
      }
      role="columnheader"
      scope="col">
      {children}
    </th>
  );
};

export const DataTableBody: React.FC = ({children}) => {
  return <tbody className="mdc-data-table__content">{children}</tbody>;
};

export const DataTableRow: React.FC<any> = ({onClick, children}) => {
  return (
    <tr className="mdc-data-table__row" onClick={onClick}>
      {children}
    </tr>
  );
};

export const DataTableCell: React.FC<CellProps> = ({numeric, children}) => {
  return (
    <td
      className={
        'mdc-data-table__cell' +
        (numeric ? ' mdc-data-table__cell--numeric' : '')
      }>
      {children}
    </td>
  );
};

export const DataTableCellDivider: React.FC<CellProps> = ({children}) => {
  return <td className={'mdc-data-table__cell mat-cell'}>{children}</td>;
};
