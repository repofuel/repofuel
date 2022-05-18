// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package apientry

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/internal/loginwith1"
	"github.com/repofuel/repofuel/accounts/internal/loginwith2"
	"github.com/repofuel/repofuel/accounts/internal/rest"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/pkg/common"
)

type Handler struct {
	auth        *jwtauth.Authenticator
	mgr         *AuthProvidersManage
	restHandler *rest.Handler
}

func NewHandler(auth *jwtauth.Authenticator, mgr *AuthProvidersManage, restHandler *rest.Handler) *Handler {
	return &Handler{
		auth:        auth,
		mgr:         mgr,
		restHandler: restHandler,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.HandleFunc("/accounts/login/{provider}", h.mgr.HandleLogin)
	r.HandleFunc("/accounts/login/{provider}/callback", h.mgr.HandleCallback)

	//deprecated endpoints, moved to the ingest service
	r.Mount("/accounts/apps/github/webhook", http.RedirectHandler("/ingest/apps/github/webhook", http.StatusPermanentRedirect))
	r.Mount("/accounts/apps/github/add_repository", http.RedirectHandler("/ingest/apps/github/add_repository", http.StatusPermanentRedirect))
	r.Mount("/accounts/auth", h.auth.Routes())
	r.Mount("/accounts", h.restHandler.Routes())

	r.Mount("/debug", middleware.Profiler())

	return r
}

type OauthHandler interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request)
}

type AuthProvidersManage struct {
	authProvidersDB entity.AuthProviderDataSource
	integrations    map[string]OauthHandler
	auth            *jwtauth.Authenticator
	usersDB         entity.UsersDataSource
	mu              sync.Mutex
}

func NewAuthProvidersManage(auth *jwtauth.Authenticator, authProvidersDB entity.AuthProviderDataSource, usersDB entity.UsersDataSource) *AuthProvidersManage {
	return &AuthProvidersManage{
		authProvidersDB: authProvidersDB,
		integrations:    make(map[string]OauthHandler),
		auth:            auth,
		usersDB:         usersDB,
	}
}

func (mgr *AuthProvidersManage) HandleLogin(w http.ResponseWriter, r *http.Request) {
	h, err := mgr.providerHandler(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.HandleLogin(w, r)
}

func (mgr *AuthProvidersManage) HandleCallback(w http.ResponseWriter, r *http.Request) {
	h, err := mgr.providerHandler(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.HandleCallback(w, r)
}

func (mgr *AuthProvidersManage) providerHandler(r *http.Request) (OauthHandler, error) {
	provider := chi.URLParam(r, "provider")
	h, ok := mgr.integrations[provider]
	if ok {
		return h, nil
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	h, ok = mgr.integrations[provider]
	if ok {
		return h, nil
	}

	p, err := mgr.authProvidersDB.FindByID(r.Context(), provider)
	if err != nil {
		return nil, err
	}

	var fetchFunc common.FetchAuthUserFunc
	switch p.System {
	case common.SystemGithub, common.SystemGithubEnterprise:
		s, err := url.Parse(p.Server)
		if err != nil {
			return nil, err
		}
		fetchFunc = loginwith2.FetchGithubUser(p.ID, s)
	case common.SystemBitbucketServer:
		fetchFunc = loginwith2.FetchBitbucketServerUserFunc(p.ID, p.Server)
	case common.SystemJiraServer, common.SystemJiraCloud:
		fetchFunc = loginwith2.FetchJiraUserFunc(p.ID, p.Server)

	default:
		return nil, errors.New("cannot find auth integration")
	}

	if p.OAuth2 != nil {
		h = loginwith2.NewHandler(mgr.auth, mgr.usersDB, fetchFunc, p.OAuth2.Config())
	} else if p.OAuth1 != nil {
		h = loginwith1.NewHandler(mgr.auth, mgr.usersDB, fetchFunc, p.OAuth1.Config())
	} else {
		return nil, errors.New("cannot find OAuth configuration")
	}

	mgr.integrations[provider] = h
	return h, nil
}
