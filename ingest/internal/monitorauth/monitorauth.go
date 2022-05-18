package monitorauth

import (
	"context"
	"net/http"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/repofuel"
)

type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

type Transport struct {
	Source TokenSource
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.Source.Token(req.Context())
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	return http.DefaultTransport.RoundTrip(req)
}

type RepositoryTokenSource struct {
	token string

	Provider string
	RepoID   identifier.RepositoryID

	accountsSrv *repofuel.AccountsService
	monitorDB   entity.MonitorDataSource
}

func NewRepositoryTokenSource(provider string, repoID identifier.RepositoryID, accountsSrv *repofuel.AccountsService, monitorDB entity.MonitorDataSource) *RepositoryTokenSource {
	return &RepositoryTokenSource{
		Provider:    provider,
		RepoID:      repoID,
		accountsSrv: accountsSrv,
		monitorDB:   monitorDB,
	}
}

func (ts *RepositoryTokenSource) Token(ctx context.Context) (string, error) {
	if ts.token != "" {
		return ts.token, nil
	}

	userID, err := ts.monitorDB.LastRepositoryMonitorUserID(ctx, ts.RepoID)
	if err != nil {
		return "", err
	}

	token, _, err := ts.accountsSrv.ProvderToken(ctx, userID.Hex(), ts.Provider)
	if err != nil {
		return "", err
	}

	ts.token = token.AccessToken

	return ts.token, nil
}

type UserTokenSource struct {
	token string

	provider string
	userID   permission.UserID

	accountsSrv *repofuel.AccountsService
}

func NewUserTokenSource(accountsSrv *repofuel.AccountsService, provider string, userID permission.UserID) *UserTokenSource {
	return &UserTokenSource{
		provider:    provider,
		userID:      userID,
		accountsSrv: accountsSrv,
	}
}

func (ts *UserTokenSource) Token(ctx context.Context) (string, error) {
	if ts.token != "" {
		return ts.token, nil
	}

	token, _, err := ts.accountsSrv.ProvderToken(ctx, ts.userID.Hex(), ts.provider)
	if err != nil {
		return "", err
	}

	ts.token = token.AccessToken

	return ts.token, nil
}
