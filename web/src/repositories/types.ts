import {Stage} from './components/__generated__/JobsListItem_job.graphql';

export interface RepositoryAddress {
  platform: ProviderName;
  owner: string;
  repo: string;
}

export interface RepositoriesState {
  isFetching?: boolean;
  reposList: RepositoriesMap;
}

export interface RepositoriesMap {
  [addr: string]: Repository;
}

export interface CommitState {
  isFetching?: boolean;
  error?: string;
  commit?: Commit;
}

export interface PullRequestState {
  isFetching?: boolean;
  isCommitsFetching?: boolean;
  error?: string;
  pull?: PullRequest;
  commits?: Commit[];
}

export interface Model {
  id: string;
  repo_id: string;
  version: number;
  data: ModelDataPoints;
  report: ModelReport;
  medians: any;
  expired: boolean;
  created_at: string;
  last_use: string;
}

export interface ModelReport {
  format: string;
  params: ModelHyperParameters;
  feature_importance: {[feature: string]: number};
  accuracy: number;
  buggy: ModelClassificationMetrics;
  weighted_avg: ModelClassificationMetrics;
  macro_avg: ModelClassificationMetrics;
}

export interface ModelDataPoints {
  all: number;
  train: ModelTagsStat;
  test: ModelTagsStat;
  predict: number;
}

interface ModelTagsStat {
  buggy: number;
  non_buggy: number;
}

interface ModelHyperParameters {
  n_estimators: number;
  max_features: string;
  max_depth: number;
  bootstrap: boolean;
}

interface ModelClassificationMetrics {
  f1_score: number;
  precision: number;
  recall: number;
  support: number;
}

export interface Repository {
  id: string;
  status: string;
  listed?: boolean;
  isFetching: boolean;
  commits?: Commit[];
  isCommitsLoading?: boolean;
  jobs?: Job[];
  models?: Model[];
  isJobsLoading?: boolean;
  selected_branch?: string;
  error?: {status: number; message: string};

  pulls?: PullRequest[];
  isPullsLoading?: boolean;
  quality: number;
  branches: string[];
  source_id: string;
  name: string;
  owner: string;
  provider: 'github' | 'bitbucket';
  html_url: string;
  source: any;
  description: string;
  default_branch: string;
  created_at: string;
  source_created_at: string;
  collaborators_count: number;
  commits_count: number;
}

export interface Commit {
  id: string;
  message: string;
  author: CommitAuthor;
  created?: string;
  risk?: number;
  explain?: ExplainedRisk;
  metrics?: {[metric: string]: any};
  classification: string;
  tags: string[];
  files: {[name: string]: FileInfo};
  issues?: Issue[];
  fixes?: string[];
  insights?: Insight[];
  hash?: string; // todo: it should not be optional
}

export interface Insight {
  name: string;
  icon?: string;
  description: string;
}

export interface ExplainedRisk {
  diffusion: number;
  experience: number;
  history: number;
  size: number;
}

export interface FileInfo {
  Action: number;
}

export interface Issue {
  id: string;
  time?: string;
}

export interface CommitAuthor {
  name: string;
  avatar?: string;
  email?: string;
  date: string;
}

export interface Job {
  id: string;
  repo_id: string;
  log: JobStage[];
  error?: string;
}

export interface JobStage {
  status: Stage;
  startedAt: string;
}

export interface PullRequest {
  id: string;
  status: string;
  number: number;
  title: string;
  body: string;
  user: string;
  open: boolean;
  merged: boolean;
  created_at: string;
}

export type ProviderName = 'github' | 'bitbucket';
