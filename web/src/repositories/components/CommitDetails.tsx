import React, {useEffect, useRef, useState} from 'react';
import {CommitsListItem, CommitsListSlimItem} from './CommitsList';
import {RepositoryAddress} from '../types';
import {GridCell, GridRow} from '@rmwc/grid';
import {Chip, ChipSet} from '@rmwc/chip';
import {
  DiffAddedIcon,
  DiffModifiedIcon,
  DiffRemovedIcon,
  DiffRenamedIcon,
  FileCodeIcon,
  IssueOpenedIcon,
  LightBulbIcon,
  ToolsIcon,
  XIcon,
} from '@primer/octicons-react';
import {formatDistanceToNow} from 'date-fns';
import {Chart} from 'chart.js';
import {Card} from '@rmwc/card';
import {Typography} from '@rmwc/typography';
import {Tooltip} from '@rmwc/tooltip';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {
  faPaperPlane,
  faQuestionCircle,
} from '@fortawesome/free-regular-svg-icons';
import {SubsectionHeader} from '../../ui/UI';
import {
  faHistory,
  faPuzzlePiece,
  faRuler,
  faSpinner,
  faUserClock,
  IconDefinition,
} from '@fortawesome/free-solid-svg-icons';
import {useFragment, useLazyLoadQuery, useMutation} from 'react-relay/hooks';
import graphql from 'babel-plugin-relay/macro';
import {
  CommitDetails_commit,
  CommitDetails_commit$key,
  DeltaType,
} from './__generated__/CommitDetails_commit.graphql';
import {Page404, Page404Custom} from '../../ui/Page404';
import {CommitDetailsModalQuery} from './__generated__/CommitDetailsModalQuery.graphql';
import {RouteComponentProps, useHistory, useParams} from 'react-router-dom';
import {RepositoryLayout} from './RepositoryLayout';
import {CommitDetailsQuery} from './__generated__/CommitDetailsQuery.graphql';
import {Modal} from '../../ui/Modal';
import './CommitDetails.scss';
import {Label} from '@primer/components';
import {Button} from '@rmwc/button';
import {
  ListItemPrimaryText,
  ListItemSecondaryText,
  ListItemText,
} from '@rmwc/list';
import styled from 'styled-components';
import {dialogs} from '../../ui/dialogs';
import {CommitDetails_SendFeedbackMutation} from './__generated__/CommitDetails_SendFeedbackMutation.graphql';
import {
  CommitDetails_DeleteTagMutation,
  Tag,
} from './__generated__/CommitDetails_DeleteTagMutation.graphql';

interface CommitDetailsProps {
  commit: CommitDetails_commit$key;
  repoURL: string;
  platform: string;
  repoAddr: RepositoryAddress;
}

export const CommitDetails: React.FC<CommitDetailsProps> = (props) => {
  const history = useHistory();
  const commit = useFragment(
    graphql`
      fragment CommitDetails_commit on Commit
      @refetchable(queryName: "CommitDetailsRefreshQuery") {
        ...CommitsList_Item_commit
        id
        hash
        metrics {
          age
          entropy
          exp
          ha
          hd
          la
          ld
          lt
          nd
          ndev
          nf
          ns
          nuc
          rexp
          sexp
        }
        author {
          name
          date
        }
        files {
          insights {
            name
            icon
            color
            description
          }
          path
          fix
          type
          language
          fixing {
            ...CommitsList_SlimItem_commit
          }
          action
          metrics {
            la
            ld
            ha
            hd
            lt
            ndev
            age
            nuc
          }
        }
        analysis {
          indicators {
            experience
            history
            size
            diffusion
          }
          insights {
            description
            icon
          }
          bugPotential
        }
        tags
        issues {
          id
          fetched
          createdAt
        }
        fixes(first: 100) {
          nodes {
            hash
            ...CommitsList_SlimItem_commit
          }
        }
      }
    `,
    props.commit
  );

  return (
    <GridRow>
      <GridCell span={12}>
        <Card outlined>
          <div className="mdc-list--non-interactive divided-list mdc-list mdc-list--two-line mdc-list--avatar-list">
            <CommitsListItem
              repoURL={props.repoURL}
              commit={commit}
              platform={props.platform}
              onSelect={() =>
                history.push(commitAddress(props.repoAddr, commit.hash))
              }
            />
          </div>
        </Card>
      </GridCell>
      <GridCell span={8}>
        {/* <CommitTags
          repoAddr={props.repoAddr}
          commitHash={commit.hash}
          tags={commit.tags}
        /> */}

        <div className="commit-summaries">
          {!!commit.issues?.length && <IssuesDetails issues={commit.issues} />}

          {!!commit.fixes?.nodes?.length && (
            <FixesDetails
              repoURL={props.repoURL}
              platform={props.platform}
              repoAddr={props.repoAddr}
              fixes={commit.fixes}
            />
          )}

          {commit.analysis?.insights?.map((insight, i) => (
            <div className="details-summary" key={i}>
              <LightBulbIcon className="details-icon" />
              {insight.description}.
            </div>
          ))}

          {!!commit.files.length && <FileDetails files={commit.files} />}
        </div>
      </GridCell>

      <GridCell span={4} style={{position: 'relative'}}>
        {/* <FeedbackBox>
          <span>Is it bug free?</span>
          <IconButton icon={<ThumbsupIcon />} label="Not buggy" />
          <IconButton icon={<ThumbsdownIcon />} label="Buggy" />
        </FeedbackBox> */}

        {commit.analysis && (
          <ExplainabilityFigure dimensions={commit.analysis.indicators} />
        )}
      </GridCell>

      <CenterGridCell span={12}>
        <FeedbackButton commitID={commit.id} />
      </CenterGridCell>

      {commit.metrics &&
        false &&
        Object.entries(RiskIndicators).map(
          ([indecenter, indecenter_metrics]) => (
            <GridCell key={indecenter} span={6} tablet={12}>
              <SubsectionHeader>
                <FontAwesomeIcon icon={DimensionIcons[indecenter]} />{' '}
                {indecenter}
              </SubsectionHeader>
              <GridRow>
                {indecenter_metrics.map((metric) => (
                  <GridCell key={metric} phone={2}>
                    <MetricCard
                      metric={metric}
                      value={
                        // @ts-ignore fixme: should not ignore linters
                        commit.metrics[metric]
                      }
                    />
                  </GridCell>
                ))}
              </GridRow>
            </GridCell>
          )
        )}
    </GridRow>
  );
};

interface FeedbackButtonProps {
  commitID: string;
}
const FeedbackButton: React.FC<FeedbackButtonProps> = (props) => {
  const [
    commit,
    isInFlight,
  ] = useMutation<CommitDetails_SendFeedbackMutation>(graphql`
    mutation CommitDetails_SendFeedbackMutation(
      $input: SendCommitFeedbackInput!
    ) {
      sendCommitFeedback(input: $input) {
        id
      }
    }
  `);

  function handelFeedbackSubmission(message: string | null) {
    if (!message) {
      return;
    }

    commit({
      variables: {
        input: {
          commitID: props.commitID,
          message: message,
        },
      },
    });
  }

  function handelFeedbackButtonClick() {
    dialogs
      .prompt({
        title: 'Send Feedback',
        body: 'Have feedback? We’d love to hear it.',
        acceptLabel: 'Submit',
        cancelLabel: 'Cancel',
        inputProps: {
          textarea: true,
          outlined: true,
          fullwidth: true,
          label: 'Your message',
          // rows: 8,
          maxLength: 10000,
          characterCount: true,
          helpText: {
            validationMsg: true,
            children: 'The field is required',
          },
        },
      })
      .then(handelFeedbackSubmission);
  }
  return (
    <Button
      outlined
      disabled={isInFlight}
      icon={
        isInFlight ? (
          <FontAwesomeIcon icon={faSpinner} spin />
        ) : (
          <FontAwesomeIcon icon={faPaperPlane} />
        )
      }
      onClick={handelFeedbackButtonClick}>
      Send Feedback
    </Button>
  );
};

// const FeedbackBox = styled.div`
//   display: flex;
//   flex-direction: row;
//   justify-content: center;
//   margin: 16px 0;

//   & > span {
//     display: flex;
//     flex-direction: column;
//     justify-content: center;
//   }
// `;

const CenterGridCell = styled<any>(GridCell)`
  text-align: center;
`;

interface CommitTagsProps {
  tags: readonly Tag[];
  commitID: string;
}

// fixme: to be remove if not need later
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const CommitTags: React.FC<CommitTagsProps> = ({commitID, tags = []}) => {
  return (
    <ChipSet choice>
      {tags.map((tag) => (
        <TagChip key={tag} commitID={commitID} tag={tag} />
      ))}
    </ChipSet>
  );
};

interface TagChipProps {
  tag: Tag;
  commitID: string;
}

const TagChip: React.FC<TagChipProps> = ({commitID, tag}) => {
  const [
    commit,
    isInFlight,
  ] = useMutation<CommitDetails_DeleteTagMutation>(graphql`
    mutation CommitDetails_DeleteTagMutation($input: DeleteCommitTagInput!) {
      deleteCommitTag(input: $input) {
        commit {
          tags
        }
      }
    }
  `);

  function handelDeleteTag() {
    if (isInFlight) return;

    commit({
      variables: {
        input: {
          commitID: commitID,
          tag: tag,
        },
      },
    });
  }

  return (
    <Chip
      trailingIconRemovesChip={false}
      selected={true}
      trailingIcon={
        isInFlight ? <FontAwesomeIcon icon={faSpinner} spin /> : <XIcon />
      }
      onTrailingIconInteraction={handelDeleteTag}>
      {tag}
    </Chip>
  );
};

interface FixesDetailsProps {
  repoAddr: RepositoryAddress;
  fixes: CommitDetails_commit['fixes'];
  platform: string;
  repoURL: string;
}

const FixesDetails: React.FC<FixesDetailsProps> = (props) => {
  const [isOpen, setOpen] = useState(false);

  const fixes = props.fixes?.nodes;
  if (!fixes) return null;

  return (
    <>
      <div className="details-summary">
        <ToolsIcon className="details-icon" />
        This commit was fixed by
        <span
          onClick={() => setOpen(!isOpen)}
          className="details-summary-button margin-left">
          {fixes.length === 1 ? 'one commit' : fixes.length + ' commits'}
        </span>
        .
      </div>
      {isOpen && (
        <div className="divided-list">
          {fixes.map(
            (commit) =>
              commit && (
                <CommitsListSlimItem
                  key={commit.hash}
                  repoAddr={props.repoAddr}
                  repoURL={props.repoURL}
                  platform={props.platform}
                  commit={commit}
                />
              )
          )}
        </div>
      )}
    </>
  );
};

//deprecated
export function commitAddress(repoAddr: RepositoryAddress, hash: string) {
  return `/repos/${repoAddr.platform}/${repoAddr.owner}/${repoAddr.repo}/commits/${hash}`;
}

interface FileDetailsProps {
  files: CommitDetails_commit['files'];
}

const FileDetails: React.FC<FileDetailsProps> = ({files}) => {
  const [isOpen, setOpen] = useState(true);

  return (
    <>
      <div className="details-summary">
        <FileCodeIcon className="details-icon" />
        This commit modifies
        <span
          onClick={() => setOpen(!isOpen)}
          className="details-summary-button margin-left">
          {files.length === 1 ? 'one file' : files.length + ' files'}
        </span>
        .
      </div>
      {isOpen && (
        <div className="no-pointer divided-list mdc-list mdc-list--two-line mdc-list--avatar-list">
          {files.map((file) => (
            <div key={file.path} className="mdc-list-item">
              <ListItemText>
                <ListItemPrimaryText>
                  <FileDetailsItem action={file.action} /> {file.path}
                </ListItemPrimaryText>
                <ListItemSecondaryText>
                  {file.type || 'Non code'}
                  {file.language && ' - ' + file.language}
                  {file.insights?.map((insight) => (
                    <Tooltip
                      key={insight.name}
                      content={insight.description}
                      showArrow>
                      <Label
                        style={{
                          cursor: 'pointer',
                        }}
                        ml={2}
                        variant="small"
                        bg={insight.color || undefined}>
                        {insight.name}
                      </Label>
                    </Tooltip>
                  ))}
                </ListItemSecondaryText>
              </ListItemText>
            </div>
          ))}
        </div>
      )}
    </>
  );
};

interface FileDetailsItemProps {
  action: DeltaType;
}

const FileDetailsItem: React.FC<FileDetailsItemProps> = ({action}) => {
  switch (action) {
    case 'DELETED':
      return <DiffRemovedIcon className="red" />;
    case 'ADDED':
      return <DiffAddedIcon className="green" />;
    case 'MODIFIED':
    case 'UNMODIFIED':
      return <DiffModifiedIcon className="yellow" />;
    case 'RENAMED':
      return <DiffRenamedIcon className="yellow" />;
    default:
      return <span>{action} - </span>;
  }
};

interface IssuesDetailsProps {
  issues: NonNullable<CommitDetails_commit['issues']>;
}

const IssuesDetails: React.FC<IssuesDetailsProps> = ({issues}) => {
  const [isOpen, setOpen] = useState(false);

  return (
    <>
      <div className="details-summary">
        <IssueOpenedIcon className="details-icon" />
        This commit is linked to
        <span
          onClick={() => setOpen(!isOpen)}
          className="details-summary-button margin-left">
          {issues.length === 1 ? 'one issue' : issues.length + ' issues'}
        </span>
        .
      </div>
      {isOpen && (
        <div className="divided-list">
          {issues.map((issue) => (
            <div key={issue.id} className="list-item commit-overview-list-item">
              Issue {issue.id}{' '}
              {issue.createdAt &&
                'opened ' +
                  formatDistanceToNow(new Date(issue.createdAt), {
                    addSuffix: true,
                  })}
            </div>
          ))}
        </div>
      )}
    </>
  );
};

const DimensionIcons: {[name: string]: IconDefinition} = {
  Diffusion: faPuzzlePiece,
  Experience: faUserClock,
  Size: faRuler,
  History: faHistory,
};

interface ExplainabilityFigureProps {
  dimensions: NonNullable<CommitDetails_commit['analysis']>['indicators'];
}

const ExplainabilityFigure: React.FC<ExplainabilityFigureProps> = ({
  dimensions,
}) => {
  let doughnut: any = useRef(null);
  useEffect(() => {
    if (doughnut == null) {
      return;
    }

    const data = Object.values(dimensions).map(
      (n) => Math.round((n || 0) * 1000) / 10
    );
    const total = data.reduce((a, b) => a + b, 0);
    const ctxChart = doughnut.current.getContext('2d');
    new Chart(ctxChart, {
      type: 'polarArea',
      data: {
        labels: Object.keys(dimensions),
        datasets: [
          {
            borderColor: 'rgba(255,255,255,0.56)',
            hoverBorderColor: 'rgba(129,112,174,0.26)',
            backgroundColor: data.map((n) => (n / total) * 100).map(riskColor),
            data: data,
          },
        ],
      },
      options: {
        responsive: true,
        aspectRatio: 1.5,
        tooltips: {
          displayColors: false,
          bodyFontSize: 16,
        },
        legend: {
          display: false,
        },
        scale: {
          ticks: {
            display: false,
            stepSize: Math.max(...data) / 3,
          },
        },
      },
    });
  }, [dimensions]);

  return (
    <>
      {Object.values(DimensionIcons).map((icon, i) => (
        <div key={i} className="radar-box noselect">
          <div className={'radar-box-content radar-box-' + i}>
            <FontAwesomeIcon icon={icon} size={'lg'} />
          </div>
        </div>
      ))}
      <canvas style={{position: 'absolute'}} ref={doughnut} />
    </>
  );
};

function riskColor(risk: number): string {
  risk = Math.round(risk / 10);
  if (risk < 0) risk = 0;
  // these colors match the same scale SCSS i.e., .risk-points
  switch (risk) {
    case 0:
      return 'rgba(124,179,66,0.50)';
    case 1:
      return 'rgba(174,213,129,0.50)';
    case 2:
      return 'rgba(220,231,117,0.50)';
    case 3:
      return 'rgba(205,220,57,0.50)';
    case 4:
      return 'rgba(255,167,38,0.50)';
    case 5:
      return 'rgba(251,140,0,0.50)';
    case 6:
      return 'rgba(239,108,0,0.50)';
    case 7:
      return 'rgba(230,81,0,0.50)';
    case 8:
      return 'rgba(229,57,53,0.50)';
    case 9:
      return 'rgba(183,28,28,0.50)';
    default:
      return 'rgba(213,0,0,0.50)';
  }
}

const MetricDict = {
  ns: 'Number of modified subsystems',
  nd: 'Number of modified directories',
  nf: 'Number of modified files',
  entropy: 'Entropy (distribution) of the changes',
  la: 'Lines added',
  ld: 'Lines deleted',
  lt: 'Total lines',
  ndev: 'Number of developers contributing',
  age: 'Age from last change',
  nuc: 'Number of unique changes',
  exp: 'Developer experience',
  rexp: 'Recent developer experience',
  sexp: 'Subsystem developer experience',
};

const RiskIndicators = {
  Experience: ['exp', 'rexp', 'sexp'],
  History: ['ndev', 'nuc', 'age'],
  Diffusion: ['ns', 'nd', 'nf', 'entropy', 'ha', 'hd'],
  Size: ['la', 'ld', 'lt'],
};

interface MetricCardProps {
  // metric: "ns"|"nd"|"nf"
  metric: string;
  value: number;
}

const MetricCard: React.FC<MetricCardProps> = ({metric, value}) => {
  // @ts-ignore
  const description = MetricDict[metric];
  return (
    <Card outlined className="metric-card">
      <Typography use="button">
        {metric}
        {description && (
          <Tooltip content={description} showArrow>
            <FontAwesomeIcon className="margin-left" icon={faQuestionCircle} />
          </Tooltip>
        )}
      </Typography>
      <div className="metric">{Math.round(value * 100) / 100}</div>
    </Card>
  );
};

interface CommitDetailsModalProps extends CommitDetailsContainerProps {
  handelClose: () => void;
}

export const CommitDetailsModal: React.FC<CommitDetailsModalProps> = ({
  handelClose,
  ...props
}) => {
  return (
    <Modal title="Commit Details" handelClose={handelClose}>
      <CommitDetailsContainer {...props} />
    </Modal>
  );
};

interface CommitDetailsContainerProps {
  commitID: string;
  repoURL: string;
  platform: string;
  repoAddr: RepositoryAddress;
}

const CommitDetailsContainer: React.FC<CommitDetailsContainerProps> = (
  props
) => {
  const {node: commit} = useLazyLoadQuery<CommitDetailsModalQuery>(
    graphql`
      query CommitDetailsModalQuery($id: ID!) {
        node(id: $id) {
          ... on Commit {
            ...CommitDetails_commit
          }
        }
      }
    `,
    {id: props.commitID},
    {fetchPolicy: 'store-and-network'}
  );

  if (!commit) return <Page404Custom>Cannot find the commit</Page404Custom>;

  return (
    <CommitDetails
      repoURL={props.repoURL}
      platform={props.platform}
      repoAddr={props.repoAddr}
      commit={commit}
    />
  );
};

interface CommitDetailsScreenProps extends RouteComponentProps {}

export const CommitDetailsScreen: React.FC<CommitDetailsScreenProps> = (
  props
) => {
  const repoAddr: any = useParams();
  const {platform, owner, repo, hash} = repoAddr;

  const {repository} = useLazyLoadQuery<CommitDetailsQuery>(
    graphql`
      query CommitDetailsQuery(
        $provider: String!
        $owner: String!
        $name: String!
        $hash: String!
      ) {
        repository(provider: $provider, owner: $owner, name: $name) {
          ...RepositoryLayout_repository
          providerSCM
          source {
            url
          }
          commit(hash: $hash) {
            ...CommitDetails_commit
          }
        }
      }
    `,
    {
      provider: platform,
      owner,
      name: repo,
      hash: hash,
    },
    {fetchPolicy: 'store-and-network'}
  );

  if (!repository || !repository.commit)
    return <Page404 location={props.location} />;

  return (
    <RepositoryLayout repository={repository}>
      <CommitDetails
        repoAddr={repoAddr}
        platform={repository.providerSCM}
        repoURL={repository.source.url}
        commit={repository.commit}
      />
    </RepositoryLayout>
  );
};
