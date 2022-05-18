package entity

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

type LoginMethod uint8

const (
	LoginDisabled LoginMethod = iota
	LoginOauth1
	LoginOauth2
)

type AuthMethod uint8

func (l AuthMethod) String() string {
	switch l {
	case OAuth:
		return "OAuth"
	case BasicAuth:
		return "Basic authentication"
	case 0:
		return "None"
	default:
		return fmt.Sprintf("Linking Method %d", l)
	}
}

const (
	BasicAuth AuthMethod = iota + 1
	OAuth
)

type ProviderConfig interface {
	Driver() string
}

type ProviderDataSource interface {
	FindByID(context.Context, string) (*Provider, error)
	FindByServer(ctx context.Context, server string) (*Provider, error)
	Insert(ctx context.Context, p *Provider) error
}

type Provider struct {
	ID         string `json:"id"         bson:"_id,omitempty"`
	Name       string
	Server     string
	Platform   common.System
	SourceCode bool
	Issues     bool
	//deprecated
	Webhook     bool
	LoginWith   LoginMethod
	AuthMethods []AuthMethod
	Config      ProviderConfig
	CreatedAt   time.Time `json:"created_at"            bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"            bson:"updated_at"`
}

type AtlassianAppLinkConfig struct {
	Server       string
	ConsumerName string
	OAuth1       credentials.OAuth1
	PublicKey    string
}

type BitbucketAppLinkConfig AtlassianAppLinkConfig

func (*BitbucketAppLinkConfig) Driver() string {
	return "bitbucket_applink"
}

type JiraAppLinkConfig AtlassianAppLinkConfig

func (*JiraAppLinkConfig) Driver() string {
	return "jira_applink"
}

type GithubAppConfig struct {
	Server        *url.URL
	AppID         int64
	AppName       string
	WebhookSecret credentials.String
	PrivateKey    credentials.Interface //todo: we can have rse private key type in the credentials
	OAuth2        *credentials.OAuth2
}

func (*GithubAppConfig) Driver() string {
	return "github_app"
}
