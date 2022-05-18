// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package entity

import (
	"context"
	"errors"
	"path"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/repofuel"
)

const (
	CurrentDataVersion uint32 = 3
	RepositoryLocation        = "./repos"
)

var (
	ErrRepositoryNotExist     = errors.New("repository not exist")
	ErrMissedViewerAccessInfo = errors.New("access info should specify the user permissions")
	ErrNoProviderInfo         = errors.New("missed user id for the repository provider")
)

type Updatable interface {
	SetUpdatedNow()
}

type RepositoryDataSource interface {
	InsertOrUpdate(context.Context, *Repository) error
	FindByID(ctx context.Context, repoID identifier.RepositoryID) (*Repository, error)
	StatusByID(ctx context.Context, repoID identifier.RepositoryID) (status.Stage, error)
	FindWhereStatusNot(context.Context, ...status.Stage) (RepositoryIter, error)
	FindByOwnerID(ctx context.Context, platform string, owner string) (RepositoryIter, error)
	FindByCollaborator(ctx context.Context, providers map[string]string) (RepositoryIter, error)
	FindUserRepos(ctx context.Context, provider, owner string) (RepositoryIter, error)
	FindUserReposByCollaborator(ctx context.Context, provider, owner, collaboratorID string) (RepositoryIter, error)
	FindByName(ctx context.Context, platform string, owner string, repo string) (*Repository, error)
	FindByProviderIDs(ctx context.Context, provider string, ids []string) (RepositoryIter, error)
	FindByProviderID(ctx context.Context, platform string, id string) (*Repository, error)
	Delete(context.Context, identifier.RepositoryID) error
	SaveStatus(context.Context, identifier.RepositoryID, status.Stage) error
	SaveCommitsCount(ctx context.Context, id identifier.RepositoryID, count int) error
	SaveBuggyCount(ctx context.Context, id identifier.RepositoryID, count int) error
	SaveBranches(ctx context.Context, id identifier.RepositoryID, bs map[string]identifier.Hash) error
	UpgradeDataVersion(ctx context.Context, id identifier.RepositoryID, v uint32) error
	SaveQuality(ctx context.Context, id identifier.RepositoryID, quality repofuel.PredictionStatus) error
	SaveConfidence(context.Context, identifier.RepositoryID, float32) error
	Branches(ctx context.Context, id identifier.RepositoryID) (map[string]identifier.Hash, error)
	UpdateOwner(ctx context.Context, id identifier.OrganizationID, owner *common.Account) error
	UpdateCollaborators(ctx context.Context, id identifier.RepositoryID, coll map[string]common.Permissions) error
	//deeprecate
	UpdateSource(context.Context, identifier.RepositoryID, *common.Repository) error
	FindAndUpdateChecksConfig(context.Context, identifier.RepositoryID, *ChecksConfig) (*Repository, error)
	AddCollaborator(ctx context.Context, platform string, repo string, user string, p common.Permissions) error
	DeleteCollaborator(ctx context.Context, platform string, repo string, user string) error

	FindOrgReposConnection(ctx context.Context, orgID identifier.OrganizationID, direction *OrderDirection, opts *PaginationInput) (RepositoryConnection, error)
	FindUserReposConnection(ctx context.Context, affiliations []*UserAffiliationInput, direction *OrderDirection, opts *PaginationInput) (RepositoryConnection, error)
	FindAllReposConnection(ctx context.Context, direction *OrderDirection, opts *PaginationInput) (RepositoryConnection, error)

	RepositoryStatisticDataSource
}

type UserAffiliationInput struct {
	UserID       identifier.UserID
	Providers    []*common.User
	Affiliations []RepositoryAffiliation
}

type RepositoryStatisticDataSource interface {
	TotalCount(ctx context.Context, since time.Time) (int64, error)
	CountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)
}

//deprecated
type SharedAccount struct {
	Provider     string `json:"provider"`
	Owner        string `json:"owner"`
	Installation string `json:"installation"`
	AvatarURL    string `json:"avatar_url"`
}

type SharedAccountIter interface {
	ForEach(context.Context, func(*SharedAccount) error) error
}

type Repository struct {
	ID            identifier.RepositoryID       `json:"id"                        bson:"_id,omitempty"`
	Organization  identifier.OrganizationID     `json:"org_id"                    bson:"org_id,omitempty"`
	Source        common.Repository             `json:"source"                    bson:"source,omitempty"`
	Owner         common.Account                `json:"owner,omitempty"           bson:"owner,omitempty"`
	ProviderSCM   string                        `json:"provider_scm"              bson:"provider_scm,omitempty"`
	ProviderITS   string                        `json:"provider_its"              bson:"provider_its,omitempty"`
	MonitorMode   bool                          `json:"monitor_mode"              bson:"monitor_mode,omitempty"`
	Collaborators map[string]common.Permissions `json:"collaborators,omitempty"   bson:"collaborators,omitempty"`
	Branches      map[string]identifier.Hash    `json:"branches"                  bson:"branches,omitempty"`
	Status        status.Stage                  `json:"status"                    bson:"status,omitempty"`
	Confidence    float32                       `json:"confidence,omitempty"      bson:"confidence,omitempty"`
	Quality       repofuel.PredictionStatus     `json:"quality,omitempty"         bson:"quality,omitempty"`
	CommitsCount  int                           `json:"commits_count"             bson:"commits_count,omitempty"`
	BuggyCount    int                           `json:"buggy_count"               bson:"buggy_count,omitempty"`
	ChecksConfig  *ChecksConfig                 `json:"checks_config,omitempty"   bson:"checks_config,omitempty"`
	DataVersion   uint32                        `json:"version"                   bson:"version,omitempty"`
	CreatedAt     time.Time                     `json:"created_at"                bson:"created_at,omitempty"`
	UpdatedAt     time.Time                     `json:"updated_at"                bson:"updated_at,omitempty"`
}

type ChecksConfig struct {
	Enable bool `bson:"enable"`
}

//deprecated
func (r *Repository) IsNode() {}

//deprecated
func (r *Repository) IsProgressable() {}

func NewRepository() *Repository {
	return &Repository{DataVersion: CurrentDataVersion}
}

func (r *Repository) Path() string {
	return pathFromID(r.ProviderSCM, r.ID)
}

func (r *Repository) IsChecksEnabled() bool {
	return r.ChecksConfig != nil && r.ChecksConfig.Enable
}

func pathFromID(provider string, repoID identifier.RepositoryID) string {
	idStr := repoID.Hex()
	return path.Join(RepositoryLocation, provider, idStr[:4], idStr)
}

//deprecated
func NewRepositoryByProvider(repo *common.Repository) *Repository {
	return &Repository{
		DataVersion: CurrentDataVersion,
		Source:      *repo,
	}
}

func (r *Repository) SetUpdatedNow() {
	r.UpdatedAt = time.Now()
}

func (r *Repository) CollaboratorsCount() int {
	return len(r.Collaborators)
}

func (r *Repository) IsAuthorized(userId permission.UserID) bool {
	panic("implement me")
}
