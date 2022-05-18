package loginwith1

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/accounts/pkg/tokens"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
	"golang.org/x/oauth2"
)

type Authenticator interface {
	LoginToken(context.Context, *permission.AccessInfo) (*oauth2.Token, error)
}

type KeyValueStorage interface {
	Set(key string, value string) error
	Get(key string) (string, error)
}

type Handler struct {
	config         *oauth1.Config
	FetchOauthUser common.FetchAuthUserFunc
	usersDB        entity.UsersDataSource
	auth           Authenticator
	db             KeyValueStorage
}

//todo: this is a simple storage: need expiration time, also this does not scale to mutable nodes
type memoryDB map[string]string

func (m memoryDB) Set(key string, value string) error {
	m[key] = value
	return nil
}

func (m memoryDB) Get(key string) (string, error) {
	if value, ok := m[key]; ok {
		delete(m, key)
		return value, nil
	}
	return "", errors.New("key is not exist")
}

func NewHandler(auth Authenticator, usersDB entity.UsersDataSource, fetchFunc common.FetchAuthUserFunc, cfg *oauth1.Config) *Handler {
	return &Handler{
		config:         cfg,
		FetchOauthUser: fetchFunc,
		usersDB:        usersDB,
		auth:           auth,
		db:             make(memoryDB),
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.HandleLogin)
	r.Get("/callback", h.HandleCallback)

	return r
}

func (handler *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	token, secret, err := handler.config.RequestToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	authURL, err := handler.config.AuthorizationURL(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = handler.db.Set(token, secret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	http.Redirect(w, r, authURL.String(), http.StatusFound)
}

//todo: move the internal logic to different function that return error
func (handler *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	requestSecret, err := handler.db.Get(requestToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	accessToken, accessSecret, err := handler.config.AccessToken(requestToken, requestSecret, verifier)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	usr, err := handler.UserByAccessToken(ctx, oauth1.NewToken(accessToken, accessSecret))
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

func (handler *Handler) UserByAccessToken(ctx context.Context, token *oauth1.Token) (*entity.User, error) {
	client := handler.config.Client(ctx, token)
	oauthUser, err := handler.FetchOauthUser(ctx, client)
	if err != nil {
		return nil, err
	}
	oauthUser.Cred = (*credentials.Token)(token)

	return handler.usersDB.FindAndModifyProvider(ctx, oauthUser)
}
