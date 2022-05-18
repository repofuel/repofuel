import React, {useState} from 'react';
import {useMutation} from 'react-relay/lib/hooks';
import graphql from 'babel-plugin-relay/macro';
import {
  Button,
  Dialog,
  DialogActions,
  DialogButton,
  DialogContent,
  DialogTitle,
  TextField,
} from 'rmwc';
import {useForm} from 'react-hook-form';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {useHistory} from 'react-router-dom';
import {PlusIcon} from '@primer/octicons-react';
import {faSpinner} from '@fortawesome/free-solid-svg-icons';
import {faGithub} from '@fortawesome/free-brands-svg-icons';

export interface AddPublicRepositoryProps {}

export const AddPublicRepository: React.FC<AddPublicRepositoryProps> = () => {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button
        outlined
        onClick={() => {
          setOpen(true);
        }}>
        Monitor repos
      </Button>
      <AddPublicRepositoryDialog open={open} onClose={() => setOpen(false)} />
    </>
  );
};

export interface AddPublicRepositoryDialogProps {
  open: boolean;
  onClose: () => void;
}

export const AddPublicRepositoryDialog: React.FC<AddPublicRepositoryDialogProps> = ({
  open,
  onClose,
}) => {
  const [commit, isInFlight] = useMutation(graphql`
    mutation AddPublicRepositoryMutation($input: AddPublicRepositoryInput!) {
      addPublicRepository(input: $input) {
        errors
        repository {
          name
          owner {
            slug
          }
          providerSCM
        }
      }
    }
  `);

  const {handleSubmit, register, errors} = useForm();
  const history = useHistory();
  // const [serverErrors, setServerError] = useState(undefined);

  return (
    <Dialog open={open} onClose={onClose} renderToPortal>
      <DialogTitle>Monitor public Github repositories</DialogTitle>
      <DialogContent>
        <TextField
          style={{
            width: '100%',
          }}
          outlined
          name="nameWithOwner"
          placeholder="owner/repo"
          inputRef={register({
            pattern: {
              value: /^[\w.\-~]+\/[\w.\-~]+$/,
              message: 'Please double check the owner and name.',
            },
          })}
          invalid={errors.nameWithOwner}
          disabled={isInFlight}
          helpText={{
            validationMsg: true,
            children: errors.nameWithOwner?.message,
          }}
          icon={<FontAwesomeIcon icon={faGithub} />}
        />
      </DialogContent>
      <DialogActions>
        <DialogButton
          icon={
            isInFlight ? (
              <FontAwesomeIcon icon={faSpinner} spin />
            ) : (
              <PlusIcon />
            )
          }
          unelevated
          disabled={isInFlight}
          onClick={handleSubmit((data) => {
            commit({
              variables: {
                input: {
                  provider: 'github',
                  nameWithOwner: data.nameWithOwner,
                },
              },
              onCompleted: (data: any) => {
                const repo = data?.addPublicRepository?.repository;
                if (repo) {
                  history.push(
                    `/repos/${repo.providerSCM}/${repo.owner.slug}/${repo.name}`
                  );
                }
              },
            });
          })}>
          Add repository
        </DialogButton>
      </DialogActions>
    </Dialog>
  );
};
