package atlassian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/ingest/internal/accesscontrol"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/providers"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

var (
	ErrWrongOrgID = errors.New("wrong origination ID")
)

var (
	// jiraIssuesRegexp inspired from `((?<!([A-Z0-9]{1,10})-?)[A-Z0-9]+-\d+)` that founded on Atlassian
	// @link https://confluence.atlassian.com/stashkb/integrating-with-custom-jira-issue-key-313460921.html
	jiraIssuesRegexp = regexp.MustCompile("([A-Z0-9]{1,10}-)?([A-Z]{1}[A-Z0-9]+-(\\d+))")
)

const htmlResponse = `
<script>
window.close();
</script>
`

type JiraAppLink struct {
	orgDB          entity.OrganizationDataSource
	VerificationDB entity.VerificationDataSource
	config         *oauth1.Config
	baseUrl        string
	authCheck      *jwtauth.AuthCheck
	provider       *entity.Provider
}

func NewJiraAppLink(provider *entity.Provider, authCheck *jwtauth.AuthCheck, orgDB entity.OrganizationDataSource, VerificationDB entity.VerificationDataSource, cfg *entity.JiraAppLinkConfig) *JiraAppLink {
	return &JiraAppLink{
		orgDB:          orgDB,
		VerificationDB: VerificationDB,
		config:         cfg.OAuth1.Config(),
		baseUrl:        cfg.Server,
		authCheck:      authCheck,
		provider:       provider,
	}
}

func (app *JiraAppLink) Routes(prefix string) http.Handler {
	r := chi.NewRouter()

	//todo: we could remove org_id from the path
	r.Get(prefix+"/organizations/{org_id}/link/callback", app.HandleOrganisationLinkingCallback)

	r.Route(prefix+"/organizations/{org_id}", func(r chi.Router) {
		r.Use(
			app.authCheck.Middleware,
			accesscontrol.CtxOrganizationByID(app.orgDB),
			accesscontrol.OnlyOrganizationAdmin,
		)

		r.Get("/link", app.HandleOrganisationLinking)
		r.Post("/basic", app.HandleOrganisationBasicAuth)

	})

	return r
}

func (app *JiraAppLink) HandleOrganisationLinking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org := ctx.Value(accesscontrol.OrganizationCtxKey).(*entity.Organization)

	cfg := *app.config

	cbURL, _ := url.Parse(cfg.CallbackURL)
	cbURL = cbURL.ResolveReference(r.URL)
	cbURL.Path += "/callback"
	cfg.CallbackURL = cbURL.String()

	token, secret, err := cfg.RequestToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	authURL, err := cfg.AuthorizationURL(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = app.VerificationDB.Insert(ctx, &entity.Verification{
		ID:        token,
		ExpiredAt: time.Now().Add(9 * time.Minute),
		Payload: &entity.LinkingVerificationOauth1{
			OrgID:         org.ID,
			RequestSecret: secret,
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{
		"oauth_url": authURL.String(),
	})
	if err != nil {
		log.Println(err)
	}
}

func (app *JiraAppLink) HandleOrganisationLinkingCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if verifier != "denied" {
		err = app.handleOrganisationLinkingCallback(ctx, chi.URLParam(r, "org_id"), requestToken, verifier)
		if err != nil {
			//todo: should check the errors and return http.StatusBadRequest other status if suitable
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//todo: return an better HTML response
	_, err = fmt.Fprintf(w, htmlResponse)
	if err != nil {
		log.Println(err)
	}
}

func (app *JiraAppLink) handleOrganisationLinkingCallback(ctx context.Context, strOrgID, requestToken, verifier string) error {
	v, err := app.VerificationDB.FindByID(ctx, requestToken)
	if err != nil {
		return err
	}

	payload, ok := v.Payload.(*entity.LinkingVerificationOauth1)
	if !ok {
		return errors.New("invalid request verification")
	}

	if payload.OrgID.Hex() != strOrgID {
		log.Println(payload.OrgID.String(), strOrgID)
		return ErrWrongOrgID
	}

	accessToken, accessSecret, err := app.config.AccessToken(requestToken, payload.RequestSecret, verifier)
	if err != nil {
		return err
	}

	return app.orgDB.SetProviderConfig(ctx, payload.OrgID, app.provider.ID, &entity.JiraOAuth1Config{
		Token: &credentials.Token{
			Token:       accessToken,
			TokenSecret: accessSecret,
		},
	})
}

func (app *JiraAppLink) HandleOrganisationBasicAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org := ctx.Value(accesscontrol.OrganizationCtxKey).(*entity.Organization)

	var payload struct {
		BaseURL  string `json:"base_url"`
		Username string `json:"user"`
		Password string `json:"pass"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = app.orgDB.SetProviderConfig(ctx, org.ID, app.provider.ID, &entity.JiraBasicAuthConfig{
		Server: payload.BaseURL,
		Cred: &credentials.BasicAuth{
			Username: payload.Username,
			Password: payload.Password,
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

}

func (app *JiraAppLink) SetupURL(*entity.Organization) (string, error) {
	return "", errors.New("setup url for JIRA is not implemented")
}

func (app *JiraAppLink) Integration(ctx context.Context, repo *entity.Repository) (providers.Integration, error) {
	org, err := app.orgDB.FindByID(ctx, repo.Organization)
	if err != nil {
		return nil, err
	}

	cfg, ok := org.ProvidersConfig[repo.ProviderITS].(*entity.JiraOAuth1Config)
	if !ok {
		return nil, errors.New("organization missing credentials for the its")
	}
	return app.JiraSession(ctx, cfg)
}

func (app *JiraAppLink) JiraSession(ctx context.Context, cfg *entity.JiraOAuth1Config) (*JiraIntegration, error) {
	c := app.config.Client(ctx, (*oauth1.Token)(cfg.Token))

	jiraClient, err := jira.NewClient(c, app.baseUrl)
	if err != nil {
		return nil, err
	}

	return &JiraIntegration{jira: jiraClient}, nil
}

func BasicAuthJiraSession(cfg *entity.JiraBasicAuthConfig) (*JiraIntegration, error) {
	c := &http.Client{
		Transport: &jira.BasicAuthTransport{
			Username: cfg.Cred.Username,
			Password: cfg.Cred.Password,
		},
	}

	jiraClient, err := jira.NewClient(c, cfg.Server)
	if err != nil {
		return nil, err
	}

	return &JiraIntegration{jira: jiraClient}, nil
}

//todo: need to be renamed
type JiraIntegrationManager struct {
	providerDB entity.ProviderDataSource
}

func NewJiraIntegrationManager(providerDB entity.ProviderDataSource) *JiraIntegrationManager {
	return &JiraIntegrationManager{providerDB: providerDB}
}

func (mgr *JiraIntegrationManager) HandelCheckServerInfo(w http.ResponseWriter, r *http.Request) {
	var checkResp = struct {
		BaseURL     string                       `json:"base_url"`
		ValidServer bool                         `json:"valid_server"`
		ServerTitle string                       `json:"server_title,omitempty"`
		Provider    string                       `json:"provider,omitempty"`
		AuthMethods map[entity.AuthMethod]string `json:"auth_methods"`
	}{
		AuthMethods: make(map[entity.AuthMethod]string),
	}

	err := json.NewDecoder(r.Body).Decode(&checkResp)
	if err != nil {
		writeJson(w, &checkResp)
		log.Println(err)
		return
	}

	if !strings.HasSuffix(checkResp.BaseURL, "/") {
		checkResp.BaseURL += "/"
	}

	c, err := UnauthenticatedIntegration(checkResp.BaseURL)
	if err != nil {
		writeJson(w, &checkResp)
		log.Println(err)
		return
	}

	info, _, err := c.GetServerInfo(r.Context())
	if err != nil {
		writeJson(w, &checkResp)
		log.Println(err)
		return
	}
	checkResp.ValidServer = true
	checkResp.ServerTitle = info.ServerTitle
	checkResp.AuthMethods[entity.BasicAuth] = entity.BasicAuth.String()

	p, err := mgr.providerDB.FindByServer(r.Context(), checkResp.BaseURL)
	if err != nil {
		switch info.DeploymentType {
		case "Cloud":
			checkResp.Provider = "jira_cloud"
		case "Server":
			checkResp.Provider = "jira_server"

		}

		writeJson(w, &checkResp)
		log.Println(err)
		return
	}

	checkResp.Provider = p.ID
	for _, m := range p.AuthMethods {
		checkResp.AuthMethods[m] = m.String()
	}

	writeJson(w, &checkResp)
}

func writeJson(w http.ResponseWriter, v interface{}) {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func UnauthenticatedIntegration(baseUrl string) (*JiraIntegration, error) {
	jiraClient, err := jira.NewClient(http.DefaultClient, baseUrl)
	if err != nil {
		return nil, err
	}

	return &JiraIntegration{jira: jiraClient}, nil
}

type JiraIntegration struct {
	jira *jira.Client
}

type ServerInfo struct {
	BaseURL             string    `json:"baseUrl"`
	Version             string    `json:"version"`
	VersionNumbers      []int     `json:"versionNumbers"`
	DeploymentType      string    `json:"deploymentType"`
	BuildNumber         int       `json:"buildNumber"`
	BuildDate           jira.Time `json:"buildDate"`
	DatabaseBuildNumber int       `json:"databaseBuildNumber"`
	ScmInfo             string    `json:"scmInfo"`
	BuildPartnerName    string    `json:"buildPartnerName"`
	ServerTime          jira.Time `json:"serverTime"`
	ServerTitle         string    `json:"serverTitle"`
	DefaultLocale       Locale    `json:"defaultLocale"`
}

type Locale struct {
	Locale string `json:"locale"`
}

func (c *JiraIntegration) GetServerInfo(ctx context.Context) (*ServerInfo, *jira.Response, error) {
	req, err := c.jira.NewRequest("GET", "rest/api/2/serverInfo", nil)
	if err != nil {
		return nil, nil, err
	}

	var info ServerInfo
	resp, err := c.jira.Do(req.WithContext(ctx), &info)
	if err != nil {
		return nil, resp, jira.NewJiraError(resp, err)
	}

	return &info, resp, nil
}

func (c *JiraIntegration) IssuesFromText(ctx context.Context, s string) ([]common.Issue, bool, error) {
	var includeBugIssue bool
	ids := IssueIDsFromText(s)
	results := make([]common.Issue, len(ids))

	for i, id := range ids {
		issue, _, err := c.jira.Issue.Get(id, &jira.GetQueryOptions{
			UpdateHistory: false,
		})
		if err != nil {
			if err == context.Canceled {
				return nil, false, err
			}

			results[i] = common.Issue{
				Id:      id,
				Fetched: false,
			}
			continue
		}

		var bug bool
		for _, label := range issue.Fields.Labels {
			if label == "Bug" || label == "Defect" {
				bug = true
				includeBugIssue = true
				break
			}
		}

		results[i] = common.Issue{
			Id:        id,
			Fetched:   true,
			Bug:       bug,
			CreatedAt: time.Time(issue.Fields.Created),
		}
	}

	return results, includeBugIssue, nil
}

func IssueIDsFromText(s string) []string {
	var issues []string

	parts := jiraIssuesRegexp.FindAllStringSubmatch(s, -1)
	for _, v := range parts {
		// If the issue number > 0 (to avoid matches for PSR-0)
		if v[1] == "" && v[3] > "0" {
			issues = append(issues, v[2])
		}
	}

	return issues
}
