package loginwith2

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/accounts/pkg/tokens"
	"github.com/repofuel/repofuel/pkg/common"
	"golang.org/x/oauth2"
)

type Authenticator interface {
	LoginToken(context.Context, *permission.AccessInfo) (*oauth2.Token, error)
}

type Handler struct {
	config         *oauth2.Config
	FetchOauthUser common.FetchAuthUserFunc
	usersDB        entity.UsersDataSource
	auth           Authenticator
}

func NewHandler(auth Authenticator, srv entity.UsersDataSource, fetchFunc common.FetchAuthUserFunc, cfg *oauth2.Config) *Handler {
	return &Handler{
		config:         cfg,
		FetchOauthUser: fetchFunc,
		usersDB:        srv,
		auth:           auth,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.HandleLogin)
	r.Get("/callback", h.HandleCallback)

	return r
}

func (handler *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := handler.config.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (handler *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// todo: should be  moved to middleware
	if r.FormValue("state") != oauthStateString {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := handler.UserByAuthorizationCode(ctx, r.FormValue("code"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := handler.auth.LoginToken(ctx, usr.UserInfo())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tokens.OauthToTokenJSON(token))
	if err != nil {
		log.Println(err)
	}
}

func (handler *Handler) UserByAuthorizationCode(ctx context.Context, authCode string) (*entity.User, error) {
	token, err := handler.config.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}

	client := handler.config.Client(ctx, token)
	oauthUser, err := handler.FetchOauthUser(ctx, client)
	if err != nil {
		return nil, err
	}
	oauthUser.Cred = token.AccessToken

	return handler.usersDB.FindAndModifyProvider(ctx, oauthUser)
}

var (
	//todo: the string should be randomized and stored for each request
	oauthStateString = "pseudo-random"
)
