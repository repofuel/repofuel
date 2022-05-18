package credentials

import (
	"crypto/rsa"
	"github.com/dghubble/oauth1"
	"golang.org/x/oauth2"
)

type OAuth1 struct {
	ConsumerKey     string    `json:"consumer_key"              bson:"consumer_key"`
	ConsumerSecret  String    `json:"consumer_secret,omitempty" bson:"consumer_secret,omitempty"`
	CallbackURL     string    `json:"callback_url,omitempty"    bson:"callback_url,omitempty"`
	RequestTokenURL string    `json:"request_token_url"         bson:"request_token_url"`
	AuthorizeURL    string    `json:"authorize_url"             bson:"authorize_url"`
	AccessTokenURL  string    `json:"access_token_url"          bson:"access_token_url"`
	Realm           string    `json:"realm,omitempty"           bson:"realm,omitempty"`
	PrivateKey      Interface `json:"private_key,omitempty"     bson:"private_key,omitempty"`
}

func (auth *OAuth1) Config() *oauth1.Config {
	var s oauth1.Signer
	pk, ok := auth.PrivateKey.(*rsa.PrivateKey)
	if ok {
		s = &oauth1.RSASigner{
			PrivateKey: pk,
		}
	} else if auth.ConsumerSecret != "" {
		s = &oauth1.HMACSigner{
			ConsumerSecret: string(auth.ConsumerSecret),
		}
	}

	return &oauth1.Config{
		ConsumerKey:    auth.ConsumerKey,
		ConsumerSecret: string(auth.ConsumerSecret),
		CallbackURL:    auth.CallbackURL,
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: auth.RequestTokenURL,
			AuthorizeURL:    auth.AuthorizeURL,
			AccessTokenURL:  auth.AccessTokenURL,
		},
		Realm:  auth.Realm,
		Signer: s,
	}
}

type OAuth2 struct {
	ClientID     string           `json:"client_id"     bson:"client_id"`
	ClientSecret String           `json:"client_secret" bson:"client_secret"`
	AuthURL      string           `json:"auth_url"      bson:"auth_url"`
	TokenURL     string           `json:"token_url"     bson:"token_url"`
	AuthStyle    oauth2.AuthStyle `json:"auth_style"    bson:"auth_style"`
	RedirectURL  string           `json:"redirect_url"  bson:"redirect_url"`
	Scopes       []string         `json:"scopes"        bson:"scopes"`
}

func (auth *OAuth2) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     auth.ClientID,
		ClientSecret: string(auth.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   auth.AuthURL,
			TokenURL:  auth.TokenURL,
			AuthStyle: auth.AuthStyle,
		},
		RedirectURL: auth.RedirectURL,
		Scopes:      auth.Scopes,
	}
}
