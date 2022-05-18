package manage

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/atlassian"
	"github.com/repofuel/repofuel/ingest/pkg/ghapp"
	"github.com/repofuel/repofuel/ingest/pkg/providers"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog/log"
)

type ServiceProvider interface {
	SetupURL(*entity.Organization) (string, error)
	Integration(ctx context.Context, repo *entity.Repository) (providers.Integration, error)
	Routes(string) http.Handler
}

type MonitorProvider interface {
	MonitorRepository(ctx context.Context, userID permission.UserID, ownerSlug, repoName string) error
}

type IntegrationManager struct {
	providerDB     entity.ProviderDataSource
	orgDB          entity.OrganizationDataSource
	verificationDB entity.VerificationDataSource
	integrations   map[string]ServiceProvider
	authCheck      *jwtauth.AuthCheck
	mgr            *Manager
	mu             sync.Mutex
	//deprecated
	// todo:  to be removed
	rfc *repofuel.Client
}

func NewIntegrationManager(mgr *Manager, providerDB entity.ProviderDataSource, orgDB entity.OrganizationDataSource, verificationDB entity.VerificationDataSource, authCheck *jwtauth.AuthCheck, rfc *repofuel.Client) *IntegrationManager {
	return &IntegrationManager{
		providerDB:     providerDB,
		orgDB:          orgDB,
		verificationDB: verificationDB,
		integrations:   make(map[string]ServiceProvider),
		authCheck:      authCheck,
		mgr:            mgr,
		rfc:            rfc,
	}
}

func (p *IntegrationManager) HandelProviders(w http.ResponseWriter, r *http.Request) {
	sp, err := p.ServiceProvider(r.Context(), chi.URLParam(r, "provider"))
	if err != nil {
		log.Err(err).Msg("issue in the provider route")
		http.Error(w, "issue in the provider route", http.StatusInternalServerError)
		return
	}
	//todo: we can cache the router
	prefix := r.URL.Path[:len(r.URL.Path)-len(chi.URLParam(r, "*"))-1]
	sp.Routes(prefix).ServeHTTP(w, r)
}

func (p *IntegrationManager) ServiceProvider(ctx context.Context, providerID string) (ServiceProvider, error) {
	sp, ok := p.integrations[providerID]
	if ok {
		return sp, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	sp, ok = p.integrations[providerID]
	if ok {
		return sp, nil
	}

	provider, err := p.providerDB.FindByID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	var i ServiceProvider
	switch cfg := provider.Config.(type) {
	case *entity.GithubAppConfig:
		i = ghapp.NewGithubApp(p.mgr, p.rfc, cfg, provider.ID, p.mgr.srv.Repo, p.mgr.srv.Organization, p.mgr.srv.PullRequest, p.mgr.srv.Monitor)

	case *entity.JiraAppLinkConfig:
		i = atlassian.NewJiraAppLink(provider, p.authCheck, p.orgDB, p.verificationDB, cfg)

	case nil:
		switch provider.Platform {
		case common.SystemJiraCloud, common.SystemJiraServer:
			//fixme: temporary JiraAppLink should be abstracted
			i = atlassian.NewJiraAppLink(provider, p.authCheck, p.orgDB, p.verificationDB, new(entity.JiraAppLinkConfig))
		}

	default:
		return nil, errors.New("unrecognized provider configurations")
	}

	p.integrations[providerID] = i
	return i, nil
}

func (p *IntegrationManager) RepositorySCM(ctx context.Context, repo *entity.Repository) (providers.SourceIntegration, error) {
	sp, err := p.ServiceProvider(ctx, repo.ProviderSCM)
	if err != nil {
		return nil, err
	}

	i, err := sp.Integration(ctx, repo)
	if err != nil {
		return nil, err
	}

	scm, ok := i.(providers.SourceIntegration)
	if !ok {
		return nil, errors.New("provider does not support scm integration")
	}

	return scm, nil
}

func (p *IntegrationManager) RepositoryITS(ctx context.Context, repo *entity.Repository) (providers.IssuesIntegration, error) {
	sp, err := p.ServiceProvider(ctx, repo.ProviderITS)
	if err != nil {
		return nil, err
	}

	i, err := sp.Integration(ctx, repo)
	if err != nil {
		return nil, err
	}

	its, ok := i.(providers.IssuesIntegration)
	if !ok {
		return nil, errors.New("provider does not support its integration")
	}

	return its, nil
}

func (p *IntegrationManager) RepositoryIntegrations(ctx context.Context, repo *entity.Repository) (providers.SourceIntegration, providers.IssuesIntegration, error) {
	scm, err := p.RepositorySCM(ctx, repo)
	if err != nil {
		return nil, nil, err
	}

	if repo.ProviderITS == "" {
		return scm, nil, nil
	}

	if repo.ProviderITS == repo.ProviderSCM {
		its, ok := scm.(providers.IssuesIntegration)
		if !ok {
			return nil, nil, errors.New("provider does not support ITS integration")
		}

		return scm, its, nil
	}

	its, err := p.RepositoryITS(ctx, repo)
	if err != nil {
		return nil, nil, err
	}

	return scm, its, nil
}
