import React, {Suspense} from 'react';
import {Dialog, DialogContent, DialogTitle} from '@rmwc/dialog';
import {IconButton} from '@rmwc/icon-button';
import {PageSpinner} from './Layout';
import './Modal.scss';
import styled from 'styled-components';
import {XIcon} from '@primer/octicons-react';

interface ModalProps {
  title: string;
  handelClose: () => void;
}

export const Modal: React.FC<ModalProps> = (props) => {
  return (
    <Dialog className="modal" open renderToPortal onClosed={props.handelClose}>
      <DialogTitle>{props.title}</DialogTitle>
      <DialogContent>
        <ModalCloseButton
          icon={<XIcon />}
          label="Close"
          onClick={props.handelClose}
          ripple={false}
        />
        <Suspense fallback={<PageSpinner />}>{props.children}</Suspense>
      </DialogContent>
    </Dialog>
  );
};

const ModalCloseButton = styled<any>(IconButton)`
  position: absolute;
  top: 5px;
  right: 5px;
`;
