package configs

import (
	"log"

	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"github.com/repofuel/repofuel/accounts/pkg/keys"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/mongocon"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/repofuel/repofuel/pkg/utilconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
	"golang.org/x/oauth2/github"
)

var zeroEndpoint oauth2.Endpoint
var zeroOauth1Endpoint oauth1.Endpoint

type Configs struct {
	Keys      keys.ServiceKeys
	Repofuel  repofuel.Options
	DB        mongocon.DatabaseOptions
	Providers struct {
		Github struct {
			Oauth2 oauth2Config `yaml:"oauth2,omitempty"`
			Server string       `yaml:"server"`
		}
		Jira struct {
			Oauth1     oauth1Config       `yaml:"oauth1,omitempty"`
			PrivateKey keys.RSAPrivateKey `yaml:"private_key"`
		}
		Bitbucket struct {
			Oauth1     oauth1Config       `yaml:"oauth1,omitempty"`
			PrivateKey keys.RSAPrivateKey `yaml:"private_key"`
		}
	}
}

type oauth2Config struct {
	ClientID     string          `yaml:"client_id"`
	ClientSecret string          `yaml:"client_secret"`
	Endpoint     oauth2.Endpoint `yaml:"endpoint"`
	RedirectURL  string          `yaml:"redirect_url"`
	Scopes       []string        `yaml:"scopes"`
}

type oauth1Endpoint struct {
	RequestTokenURL string `yaml:"request_token_url"`
	AuthorizeURL    string `yaml:"authorize_url"`
	AccessTokenURL  string `yaml:"access_token_url"`
}

type oauth1Config struct {
	ConsumerKey    string         `yaml:"consumer_key"`
	ConsumerSecret string         `yaml:"consumer_secret"`
	CallbackURL    string         `yaml:"callback_url"`
	Endpoint       oauth1Endpoint `yaml:"endpoint"`
	Realm          string         `yaml:"realm"`
	Signer         oauth1.Signer
}

func (o *oauth1Config) Config() *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    o.ConsumerKey,
		ConsumerSecret: o.ConsumerSecret,
		CallbackURL:    o.CallbackURL,
		Endpoint:       oauth1.Endpoint(o.Endpoint),
		Realm:          o.Realm,
		Signer:         o.Signer,
	}
}

//deprecated
func (cfg *Configs) Oauth2Config(p common.Name) *oauth2.Config {
	var c oauth2.Config
	switch p {
	case common.Github:
		c = oauth2.Config(cfg.Providers.Github.Oauth2)
	}

	c.Scopes = []string{"user:email", "repo"}
	if c.Endpoint == zeroEndpoint {
		switch p {
		case common.Github:
			c.Endpoint = github.Endpoint
		case common.Bitbucket:
			c.Endpoint = bitbucket.Endpoint
		}
	}
	return &c
}

//deprecated
func (cfg *Configs) Oauth1Config(p common.Name) *oauth1.Config {
	var c *oauth1.Config
	switch p {
	case common.Jira:
		c = cfg.Providers.Jira.Oauth1.Config()
		c.Signer = &oauth1.RSASigner{
			PrivateKey: cfg.Providers.Jira.PrivateKey.Key(),
		}
	case common.Bitbucket:
		c = cfg.Providers.Bitbucket.Oauth1.Config()
		c.Signer = &oauth1.RSASigner{
			PrivateKey: cfg.Providers.Bitbucket.PrivateKey.Key(),
		}
	}

	return c
}

func Parse() (*Configs, error) {
	if err := godotenv.Load(".env", "accounts/.env"); err != nil {
		log.Println("missing .env file, will continue without it")
	}

	var cfg Configs
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "COMMON_SECRETS"); err != nil {
		return nil, err
	}
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "SERVICE_SECRETS"); err != nil {
		return nil, err
	}

	return &cfg, nil
}
