// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package jwtauth

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/accounts/pkg/tokens"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid RefreshToken")
	ErrInvalidIssuer       = errors.New("invalid token issuer: only the ACCOUNTS service can authenticate users")
)

const (
	// fixme: we extend the access token just for now, should be in minuets
	AccessTokenLifetime = 10 * time.Minute
)

type Authenticator struct {
	ServiceName string
	Audience    string
	privateKey  crypto.PrivateKey
	tokensDB    entity.TokenDataSource
	usersDB     entity.UsersDataSource
}

func NewAuthenticator(name string, k crypto.PrivateKey, usersDB entity.UsersDataSource, tokensDB entity.TokenDataSource) *Authenticator {
	return &Authenticator{
		ServiceName: name,
		privateKey:  k,
		tokensDB:    tokensDB,
		usersDB:     usersDB,
	}
}

func (auth *Authenticator) Routes() http.Handler {
	r := chi.NewRouter()

	//r.Get("/authorize", auth.HandelAuthorize)
	r.Post("/access_token", auth.HandelAccessToken)
	r.Post("/register_service", auth.HandelRegisterService)

	return r
}

func (auth *Authenticator) HandelAuthorize(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (auth *Authenticator) HandelAccessToken(w http.ResponseWriter, r *http.Request) {
	//todo: support the access scope
	//todo: track the client_id
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "cannot parse the form", http.StatusBadRequest)
		return
	}

	switch r.Form.Get("grant_type") {
	case "refresh_token":
		auth.HandelRefreshingAccessToken(w, r)
	case "":
		http.Error(w, "missing grant_type", http.StatusBadRequest)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (auth *Authenticator) HandelRegisterService(w http.ResponseWriter, r *http.Request) {
	// todo: need  to be  implemented
	if r.FormValue("service_token") != "i'm a dummy secret; need to be implemented" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//todo: insert it in the DB

	t, err := auth.LoginToken(r.Context(), &permission.AccessInfo{
		Role: permission.RoleService,
		ServiceInfo: &permission.ServiceInfo{
			ServiceID: "", //todo
		},
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(tokens.OauthToTokenJSON(t))
	if err != nil {
		log.Println(err)
	}
}

func (auth *Authenticator) HandelRefreshingAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := auth.RefreshAccessToken(ctx, r.Form.Get("refresh_token"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writeTokenResponse(w, token)
}

func writeTokenResponse(w http.ResponseWriter, token *oauth2.Token) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	err := json.NewEncoder(w).Encode(tokens.OauthToTokenJSON(token))
	if err != nil {
		log.Println(err)
	}
}

func (auth *Authenticator) LoginToken(ctx context.Context, userInfo *permission.AccessInfo) (*oauth2.Token, error) {
	exp := time.Now().Add(AccessTokenLifetime)

	accessToken, err := auth.CreateAccessToken(exp, userInfo)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.CreateRefreshToken(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		Expiry:       exp,
	}, nil
}

func (auth *Authenticator) CreateAccessToken(exp time.Time, accessInfo *permission.AccessInfo) (string, error) {
	c := &AccessClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  auth.Audience,
			ExpiresAt: exp.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    auth.ServiceName,
		},
		AccessInfo: accessInfo,
	}

	return jwt.NewWithClaims(jwt.SigningMethodES384, c).SignedString(auth.privateKey)
}

func (auth *Authenticator) CreateRefreshToken(ctx context.Context, usr *permission.AccessInfo) (string, error) {
	t, err := auth.tokensDB.GenerateToken(ctx, usr.UserID)
	if err != nil {
		return "", err
	}
	return t.String(), nil
}

func (auth *Authenticator) RefreshAccessToken(ctx context.Context, refToken string) (*oauth2.Token, error) {
	token, err := auth.tokensDB.Find(ctx, refToken)
	if err != nil {
		return nil, err
	}

	if !token.IsValid() {
		return nil, ErrInvalidRefreshToken
	}
	//todo: renew refresh_token is it about to expired

	usr, err := auth.usersDB.Find(ctx, token.UserId)
	if err != nil {
		return nil, err
	}

	exp := time.Now().Add(AccessTokenLifetime)

	accessToken, err := auth.CreateAccessToken(exp, usr.UserInfo())
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refToken,
		Expiry:       exp,
	}, nil
}

// TokenSource generates a token source to by used locally by the service when
// it's communicate with other Repofuel's services through the REST api.
func (auth *Authenticator) TokenSource(ctx context.Context, accessInfo *permission.AccessInfo) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, &localTokenSource{
		auth:       auth,
		accessInfo: accessInfo,
	})
}

type localTokenSource struct {
	auth       *Authenticator
	accessInfo *permission.AccessInfo
}

func (ts *localTokenSource) Token() (*oauth2.Token, error) {
	expiry := time.Now().Add(AccessTokenLifetime)
	accessToken, err := ts.auth.CreateAccessToken(expiry, ts.accessInfo)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken: accessToken,
		Expiry:      expiry,
	}, nil
}

type AccessClaims struct {
	jwt.StandardClaims
	*permission.AccessInfo
}

func (c *AccessClaims) Valid() error {
	if c.Issuer != "accounts" && c.UserInfo != nil {
		return ErrInvalidIssuer
	}

	return c.StandardClaims.Valid()
}
