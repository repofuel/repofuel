export interface Organization {
  id: string;
  owner: Account;
  provider_scm: string;
  provider_its: string;
  avatar_url: string;
  system: string;
  config: any;
}

export interface OrganizationsState {
  isFetching?: boolean;
  current?: Organization;
  orgsList: {
    [addr: string]: Organization;
  };
}

export interface Account {
  id: string;
  slug: string;
  // type: string
}
