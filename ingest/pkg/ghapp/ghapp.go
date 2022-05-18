package ghapp

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v32/github"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/internal/monitorauth"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/ingest/pkg/providers"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog/log"
)

const (
	defaultCacheLimit    = 20
	pushCheckName        = "Repofuel - Patch"
	pullRequestCheckName = "Repofuel - Pull Request"
)

var (
	ErrMissingChuckInfo      = errors.New("missing the chuck info")
	ErrMissingInstallationID = errors.New("missing the installation ID")
)

var (
	defaultBaseURL, _ = url.Parse("https://api.github.com/")
	uploadBaseURL, _  = url.Parse("https://uploads.github.com/")
	avatarBaseURL, _  = url.Parse("https://avatars.githubusercontent.com/")
	webBaseURL, _     = url.Parse("https://github.com/")
)

type RepositoryManager interface {
	AddRepository(ctx context.Context, repo *entity.Repository) error
	DeleteRepository(ctx context.Context, repo *entity.Repository) error
	ProcessRepository(info *jobinfo.JobInfo) error
}

type GithubApp struct {
	appsTransport *ghinstallation.AppsTransport

	cacheTransports map[int64]*ghinstallation.Transport
	cacheLog        []int64
	cacheIndex      int

	baseURL    *url.URL
	uploadURL  *url.URL
	avatarURL  string
	addRepoUrl string
	webURL     string

	mu sync.Mutex

	provider      string
	webhookSecret []byte

	organizationDB entity.OrganizationDataSource
	repoDB         entity.RepositoryDataSource
	pullsDB        entity.PullRequestDataSource
	monitorDB      entity.MonitorDataSource

	//deprecated
	srvClient *repofuel.Client

	mgr RepositoryManager
}

type Repository struct {
	owner  string
	repo   string
	ts     monitorauth.TokenSource
	github *github.Client
	app    *GithubApp
}

func NewGithubApp(mgr RepositoryManager, srv *repofuel.Client, cfg *entity.GithubAppConfig, provider string, repoDB entity.RepositoryDataSource, organizationDB entity.OrganizationDataSource, pullsDB entity.PullRequestDataSource, monitorDB entity.MonitorDataSource) *GithubApp {
	var appPrefix string
	var baseURL, uploadURL, avatarURL, webURL *url.URL
	if cfg.Server.Host == "github.com" {
		appPrefix = "apps"
		baseURL = defaultBaseURL
		uploadURL = uploadBaseURL
		avatarURL = avatarBaseURL
		webURL = webBaseURL
	} else {
		appPrefix = "github-apps" // for enterprise servers
		baseURL, _ = cfg.Server.Parse("/api/v3/")
		uploadURL, _ = cfg.Server.Parse("/")
		avatarURL, _ = cfg.Server.Parse("/avatars/")
		webURL, _ = cfg.Server.Parse("/")
	}

	pk, ok := cfg.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		err := fmt.Errorf("expecting Private Key that represents an RSA key, got: %T", cfg.PrivateKey)
		log.Fatal().Err(err).Msg("parsing the provider private key")
	}

	tr := ghinstallation.NewAppsTransportFromPrivateKey(http.DefaultTransport, cfg.AppID, pk)
	tr.BaseURL = strings.TrimSuffix(baseURL.String(), "/")
	addURL, _ := cfg.Server.Parse(fmt.Sprintf("/%s/%s/installations/new", appPrefix, strings.ToLower(cfg.AppName)))
	return &GithubApp{
		appsTransport:   tr,
		cacheTransports: make(map[int64]*ghinstallation.Transport, defaultCacheLimit),
		cacheLog:        make([]int64, defaultCacheLimit),
		cacheIndex:      0,
		baseURL:         baseURL,
		uploadURL:       uploadURL,
		avatarURL:       avatarURL.String(),
		addRepoUrl:      addURL.String(),
		webURL:          webURL.String(),
		mu:              sync.Mutex{},
		provider:        provider,
		webhookSecret:   []byte(cfg.WebhookSecret),
		organizationDB:  organizationDB,
		repoDB:          repoDB,
		pullsDB:         pullsDB,
		monitorDB:       monitorDB,
		srvClient:       srv,
		mgr:             mgr,
	}
}

func (ghr *Repository) BasicAuth(ctx context.Context) (*credentials.BasicAuth, error) {
	return basicAuth(ctx, ghr.ts)
}

//deprecated
func (app *GithubApp) AvatarURL(username string) string {
	return app.avatarURL + username
}

func basicAuth(ctx context.Context, ts monitorauth.TokenSource) (*credentials.BasicAuth, error) {
	t, err := ts.Token(ctx)
	if err != nil {
		return nil, err
	}

	return &credentials.BasicAuth{
		Username: "x-access-token",
		Password: t,
	}, nil
}

func (app *GithubApp) SetupURL(org *entity.Organization) (string, error) {
	cfg, ok := org.ProvidersConfig[org.ProviderSCM].(*entity.InstallationConfig)
	if !ok {
		return "", ErrMissingInstallationID
	}

	if cfg.DatabaseID > 0 {
		// The following is the ideal URL. Other forms are for backword compatibility, we can remove them later.
		return fmt.Sprintf("%s/permissions?target_id=%d", app.addRepoUrl, cfg.DatabaseID), nil
	}

	if org.Owner.Type == common.AccountPersonal {
		return fmt.Sprintf("%ssettings/installations/%d", app.webURL, cfg.InstallationID), nil
	}

	return fmt.Sprintf("%sorganizations/%s/settings/installations/%d", app.webURL, org.Owner.Slug, cfg.InstallationID), nil
}

func (app *GithubApp) Integration(ctx context.Context, repo *entity.Repository) (providers.Integration, error) {
	if repo.MonitorMode {
		return app.monitoredRepository(repo.ID, repo.Owner.Slug, repo.Source.RepoName), nil
	}

	org, err := app.organizationDB.FindByID(ctx, repo.Organization)
	if err != nil {
		return nil, err
	}

	cfg, ok := org.ProvidersConfig[repo.ProviderSCM].(*entity.InstallationConfig)
	if !ok {
		return nil, ErrMissingInstallationID
	}

	return app.installationRepository(cfg.InstallationID, repo.Owner.Slug, repo.Source.RepoName), nil
}

func (app *GithubApp) installationClient(installation int64) *github.Client {
	ts := app.installationTransport(installation)
	return app.ghClient(ts)
}

func (app *GithubApp) installationRepository(installation int64, ownerSlug, repoName string) *Repository {
	ts := app.installationTransport(installation)

	return &Repository{
		owner:  ownerSlug,
		repo:   repoName,
		ts:     ts,
		github: app.ghClient(ts),
		app:    app,
	}
}

func (app *GithubApp) monitoredRepository(id identifier.RepositoryID, ownerSlug, repoName string) *Repository {
	ts := monitorauth.NewRepositoryTokenSource(app.provider, id, app.srvClient.Accounts, app.monitorDB)

	return &Repository{
		owner: ownerSlug,
		repo:  repoName,
		ts:    ts,
		github: app.ghClient(&monitorauth.Transport{
			Source: ts,
		}),
		app: app,
	}
}

func (app *GithubApp) MonitorRepository(ctx context.Context, userID permission.UserID, ownerSlug, repoName string) error {
	client := app.ghClient(&monitorauth.Transport{
		Source: monitorauth.NewUserTokenSource(app.srvClient.Accounts, app.provider, userID),
	})

	repo, _, err := client.Repositories.Get(ctx, ownerSlug, repoName)
	if err != nil {
		return err
	}

	if repo.GetPrivate() {
		//todo: update the database if it already monitoried
		return errors.New("monitoring a private repository is not supported")
	}

	org, err := app.creatOrUpdateOrganization(ctx, client, nil, repo.Owner)
	if err != nil {
		return err
	}

	entityRepo := &entity.Repository{
		Organization: org.ID,
		Source:       *toCommonRepository(repo),
		Owner:        org.Owner,
		ProviderSCM:  app.provider,
		ProviderITS:  app.provider,
		MonitorMode:  true,
	}

	err = app.repoDB.InsertOrUpdate(ctx, entityRepo)
	if err != nil {
		return err
	}

	err = app.monitorDB.InsertMonitor(ctx, &identifier.MonitorID{
		RepoID: entityRepo.ID,
		UserID: identifier.UserID(userID),
	})
	if err != nil {
		return err
	}

	return app.mgr.ProcessRepository(&jobinfo.JobInfo{
		Action: invoke.ActionMonitorRepository,
		RepoID: entityRepo.ID,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity: entityRepo,
		},
	})
}

func (app *GithubApp) ghClient(ts http.RoundTripper) *github.Client {
	c := github.NewClient(&http.Client{
		Transport: ts,
	})
	c.BaseURL = app.baseURL
	c.UploadURL = app.uploadURL

	return c
}

func (app *GithubApp) installationTransport(id int64) *ghinstallation.Transport {
	tr, ok := app.cacheTransports[id]
	if ok {
		return tr
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	tr, ok = app.cacheTransports[id]
	if ok {
		return tr
	}

	tr = ghinstallation.NewFromAppsTransport(app.appsTransport, id)

	// removes the oldest
	delete(app.cacheTransports, app.cacheLog[app.cacheIndex])

	app.cacheLog[app.cacheIndex] = id
	app.cacheIndex = (app.cacheIndex + 1) % len(app.cacheLog)

	app.cacheTransports[id] = tr
	return tr
}

func (ghr *Repository) FetchRepositoryInfo(ctx context.Context) (*common.Repository, error) {
	repo, _, err := ghr.github.Repositories.Get(ctx, ghr.owner, ghr.repo)
	if err != nil {
		return nil, err
	}

	return toCommonRepository(repo), nil
}

func (ghr *Repository) FetchCollaborators(ctx context.Context) (map[string]common.Permissions, error) {
	var collaborators map[string]common.Permissions

	err := listAllCollaborators(ctx, ghr.github, ghr.owner, ghr.repo, func(users []*github.User) error {
		//todo: we could sync it with the organization, or filter "outside" collaborators only
		collaborators = toCommonPermissions(users)
		return nil
	})

	//todo: register the missed collaborates in the organization
	return collaborators, err
}

func (ghr *Repository) FetchRepositoryCollaborators(ctx context.Context) (map[string]common.Permissions, error) {
	//fixme: we should use listAllCollaborators to get all pages
	coll, _, err := ghr.github.Repositories.ListCollaborators(ctx, ghr.owner, ghr.repo, nil)
	if err != nil {
		return nil, err
	}

	return toCommonPermissions(coll), nil
}

func (ghr *Repository) ListOpenPullRequests(ctx context.Context) common.PullRequestItr {
	return itrAllPullRequestPages(func(page int) (requests []*github.PullRequest, response *github.Response, e error) {
		return ghr.github.PullRequests.List(ctx, ghr.owner, ghr.repo, &github.PullRequestListOptions{
			State: "open",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
	})
}

var (
	checkStatusQueued     = "queued"
	checkStatusInProgress = "in_progress"
	checkStatusCompleted  = "completed"
)

type checkInfo struct {
	ID   int64
	Name string
}

func (ghr *Repository) CreateCheckRun(ctx context.Context, headSHA, name string) (jobinfo.Store, error) {
	detailsURL := ghr.CheckRunDetailsURL()

	check, _, err := ghr.github.Checks.CreateCheckRun(ctx, ghr.owner, ghr.repo, github.CreateCheckRunOptions{
		Name:       name,
		HeadSHA:    headSHA,
		DetailsURL: &detailsURL,
		Status:     &checkStatusQueued,
	})
	if err != nil {
		return nil, err
	}

	return jobinfo.Store{
		keyChuckID:   check.GetID(),
		keyChuckName: name,
	}, nil
}

func getCheckInfo(details jobinfo.Store) (*checkInfo, error) {
	id, ok := details[keyChuckID].(int64)
	if !ok {
		return nil, ErrMissingChuckInfo
	}

	name, ok := details[keyChuckName].(string)
	if !ok {
		return nil, ErrMissingChuckInfo
	}

	return &checkInfo{
		ID:   id,
		Name: name,
	}, nil
}

func (ghr *Repository) StartCheckRun(ctx context.Context, jobID identifier.JobID, details jobinfo.Store) error {
	check, err := getCheckInfo(details)
	if err != nil {
		return err
	}

	strID := jobID.Hex()
	_, _, err = ghr.github.Checks.UpdateCheckRun(ctx, ghr.owner, ghr.repo, check.ID, github.UpdateCheckRunOptions{
		Name:       check.Name,
		ExternalID: &strID,
		Status:     &checkStatusInProgress,
	})

	return err
}

func (ghr *Repository) CheckRunDetailsURL() string {
	//todo: use dynamic domain
	const repofuelDomain = "http://dev.repofuel.com"
	return fmt.Sprintf("%s/repos/%s/%s/%s", repofuelDomain, ghr.app.provider, ghr.owner, ghr.repo)
}

func (ghr *Repository) FinishCheckRun(ctx context.Context, details jobinfo.Store, s status.Stage, summarizer providers.CheckRunSummarizer) error {
	check, err := getCheckInfo(details)
	if err != nil {
		return err
	}

	var conclusion string
	switch s {
	case status.Ready, status.Watched:
		conclusion = "neutral"
	default:
		conclusion = "cancelled"
	}

	title := summarizer.Title()
	summary := summarizer.Summary()
	text := summarizer.DetailsText(ghr.app.provider, ghr.app.webURL, ghr.owner, ghr.repo)

	_, _, err = ghr.github.Checks.UpdateCheckRun(ctx, ghr.owner, ghr.repo, check.ID, github.UpdateCheckRunOptions{
		Name:        check.Name,
		Status:      &checkStatusCompleted,
		Conclusion:  &conclusion,
		CompletedAt: &github.Timestamp{Time: time.Now()},
		Output: &github.CheckRunOutput{
			Title:   &title,
			Summary: &summary,
			Text:    &text,
		},
	})

	return err
}

type fetchFunc func(page int) ([]*github.PullRequest, *github.Response, error)

type pullRequestItr struct {
	fun fetchFunc
}

func itrAllPullRequestPages(fun fetchFunc) *pullRequestItr {
	return &pullRequestItr{fun: fun}
}

func (itr *pullRequestItr) ForEach(fun func(*common.PullRequest) error) error {
	page := 1
	for page > 0 {
		pulls, resp, err := itr.fun(page)
		if err != nil {
			return err
		}
		page = resp.NextPage

		for _, p := range pulls {
			err := fun(toCommonPullRequest(p))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func toCommonRepository(repo *github.Repository) *common.Repository {
	return &common.Repository{
		ID:            repo.GetNodeID(),
		RepoName:      repo.GetName(),
		Description:   repo.GetDescription(),
		DefaultBranch: repo.GetDefaultBranch(),
		HTMLURL:       repo.GetHTMLURL(),
		CloneURL:      repo.GetCloneURL(),
		SSHURL:        repo.GetSSHURL(),
		CreatedAt:     repo.GetCreatedAt().Time,
		Private:       repo.GetPrivate(),
	}
}

func toCommonAccount(user *github.User) common.Account {
	var t common.AccountType
	switch user.GetType() {
	case "Organization":
		t = common.AccountOrganization
	case "User":
		t = common.AccountPersonal
	}

	return common.Account{
		ID:   user.GetNodeID(),
		Slug: user.GetLogin(),
		Type: t,
	}
}

//deprecated
func toCommonPermissions(users []*github.User) map[string]common.Permissions {
	permissions := make(map[string]common.Permissions, len(users))
	for _, user := range users {

		p := user.GetPermissions()
		permissions[user.GetNodeID()] = common.Permissions{
			Admin: p["admin"],
			Read:  p["pull"],
			Write: p["push"],
		}
	}
	return permissions
}
