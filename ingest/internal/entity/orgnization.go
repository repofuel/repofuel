package entity

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

type OrganizationDataSource interface {
	FindByID(context.Context, identifier.OrganizationID) (*Organization, error)
	FindBySlug(ctx context.Context, provider, slug string) (*Organization, error)
	Delete(ctx context.Context, id identifier.OrganizationID) error
	DeleteByOwnerID(ctx context.Context, provider, ownerID string) error
	DeleteProviderConfig(ctx context.Context, orgID identifier.OrganizationID, providerID string) error
	FindOrCreate(ctx context.Context, org *Organization) (*Organization, error)
	UpdateMembers(ctx context.Context, id identifier.OrganizationID, members map[string]common.Membership) error
	UpdateOwner(ctx context.Context, id identifier.OrganizationID, owner *common.Account) error
	SetProviderConfig(ctx context.Context, id identifier.OrganizationID, provider string, cfg IntegrationConfig) error
	ListUserOrganizations(ctx context.Context, providers map[string]string) (OrganizationIter, error)
	All(ctx context.Context) (OrganizationIter, error)

	FindAllOrgsConnection(ctx context.Context, direction *OrderDirection, opts *PaginationInput) (OrganizationConnection, error)

	OrganizationStatisticDataSource
}

type OrganizationStatisticDataSource interface {
	TotalCount(ctx context.Context, since time.Time) (int64, error)
	CountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)
}

type Organization struct {
	ID              identifier.OrganizationID    `json:"id"                     bson:"_id,omitempty"`
	Owner           common.Account               `json:"owner"                  bson:"owner,omitempty"`
	ProviderSCM     string                       `json:"provider_scm"           bson:"provider_scm,omitempty"`
	ProviderITS     string                       `json:"provider_its"           bson:"provider_its,omitempty"`
	AvatarURL       string                       `json:"avatar_url"             bson:"avatar,omitempty"`
	Members         map[string]common.Membership `json:"members,omitempty"      bson:"members,omitempty"`
	ProvidersConfig map[string]IntegrationConfig `json:"config,omitempty"       bson:"config,omitempty"`
	CreatedAt       time.Time                    `json:"created_at,omitempty"   bson:"created_at,omitempty"`
	UpdatedAt       time.Time                    `json:"updated_at,omitempty"   bson:"updated_at,omitempty"`
}

func (_ *Organization) IsNode()            {}
func (_ *Organization) IsRepositoryOwner() {}

type IntegrationConfig interface {
	ConfigType() ConfigType
}

type ConfigType string

const (
	AtlassianOAuth1 ConfigType = "atlassian_oauth1"
	BitBucketOAuth1            = "bb_oauth1"
	JiraBasicAuth              = "jira_basic"
	JiraOAuth1                 = "jira_oauth1"
	Installation               = "gh_installation"
)

type InstallationConfig struct {
	DatabaseID     int64 `json:"db_id"            bson:"db_id"`
	InstallationID int64 `json:"installation_id"  bson:"installation_id"`
}

func (i *InstallationConfig) ConfigType() ConfigType {
	return Installation
}

type AtlassianOAuth1Config struct {
	Token *credentials.Token `json:"-"          bson:"token"`
}

func (a *AtlassianOAuth1Config) ConfigType() ConfigType {
	return AtlassianOAuth1
}

type JiraOAuth1Config AtlassianOAuth1Config

func (a *JiraOAuth1Config) ConfigType() ConfigType {
	return JiraOAuth1
}

type JiraBasicAuthConfig struct {
	Server string                 `json:"server,omitempty"   bson:"server,omitempty"`
	Cred   *credentials.BasicAuth `json:"-"                  bson:"cred"`
}

func (a *JiraBasicAuthConfig) ConfigType() ConfigType {
	return JiraBasicAuth
}

type BitbucketOAuth1Config AtlassianOAuth1Config

func (a *BitbucketOAuth1Config) ConfigType() ConfigType {
	return BitBucketOAuth1
}
