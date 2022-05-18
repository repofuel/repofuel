import React from 'react';
import './ButtonGroup.scss';

interface ButtonGroupProps {
  className?: string;
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  className,
  ...props
}) => {
  return (
    <div className={'button-group ' + className} {...props}>
      {children}
    </div>
  );
};
