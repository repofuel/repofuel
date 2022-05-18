// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package ghapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/go-github/v32/github"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog/log"
)

//TODO: process the coming events in a queue

func (app *GithubApp) Routes(prefix string) http.Handler {
	r := chi.NewRouter()

	r.Get(prefix+"/add_repository", app.AddRepositoryLink)
	r.Post(prefix+"/webhook", app.WebhookHandler)

	return r
}

func (app *GithubApp) AddRepositoryLink(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, app.addRepoUrl, http.StatusTemporaryRedirect)
}

func (app *GithubApp) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	eventType := github.WebHookType(r)

	switch eventType {
	case "integration_installation", "integration_installation_repositories":
		// ignore deprecated events
		// The "integration_installation" and "integration_installation_repositories" events will
		// be removed after October 1st, 2020.
		_, _ = w.Write([]byte("ignored event"))
		return
	}

	payload, err := github.ValidatePayload(r, app.webhookSecret)
	if err != nil {
		log.Err(err).Msg("validate payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = app.processWebhookEvent(ctx, eventType, payload)
	if err != nil {
		log.Err(err).Msg("process webhook")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *GithubApp) processWebhookEvent(ctx context.Context, eventType string, payload []byte) error {
	switch eventType {
	case "installation":
		return app.installationEvent(ctx, payload)

	case "installation_repositories":
		return app.installationRepositoriesEvent(ctx, payload)

	case "push":
		return app.pushEvent(ctx, payload)

	case "pull_request":
		return app.pullRequestEvent(ctx, payload)

	case "member":
		return app.memberEvent(ctx, payload)

	case "check_suite":
		return app.checkSuiteEvent(ctx, payload)
	}

	return errors.New("unsupported webhoook event")
}

// PullRequestEvent is triggered when a pull request is assigned, unassigned, labeled,
// unlabeled, opened, edited, closed, reopened, synchronize, ready_for_review,
// locked, unlocked or when a pull request review is requested or removed.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#pullrequestevent
func (app *GithubApp) pullRequestEvent(ctx context.Context, payload []byte) error {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	switch event.GetAction() {
	case "opened", "synchronize", "closed", "edited", "reopened":

	default: //e.g.,: "labeled", "unlabeled", "assigned","unassigned","ready_for_review", "locked", "unlocked"
		log.Ctx(ctx).Debug().
			Str("action", event.GetAction()).
			Msg("unhandled action for pull request event")

		return nil
	}

	repo, err := app.repoEntity(ctx, event.Installation.GetID(), event.Sender, event.Repo)
	if err != nil {
		return err
	}

	pull, err := app.pullsDB.FindAndUpdateSource(ctx, repo.ID, toCommonPullRequest(event.PullRequest))
	if err != nil {
		return err
	}

	if event.PullRequest.Head.GetSHA() == pull.AnalyzedHead.Hex() {
		// if the head analyzed before, no need to analyze it again
		return nil
	}

	var action = invoke.ActionPullRequestUpdate
	var details = make(jobinfo.Store, 1)

	if repo.IsChecksEnabled() {
		switch event.GetAction() {
		case "opened":
		case "synchronize":
			if IsSameOrigin(event.PullRequest) {
				return nil
			}
		default:
			return nil
		}

		action = invoke.ActionPullRequestCheck
		repoClient := app.installationRepository(
			event.Installation.GetID(),
			event.Repo.Owner.GetLogin(),
			event.Repo.GetName(),
		)

		details, err = repoClient.CreateCheckRun(ctx, event.PullRequest.Head.GetSHA(), pullRequestCheckName)
		if err != nil {
			// todo: maybe we should not fail if the check run is not created
			return err
		}
	}

	details[jobinfo.PullRequestID] = pull.ID

	return app.mgr.ProcessRepository(&jobinfo.JobInfo{
		Action:  action,
		RepoID:  repo.ID,
		Details: details,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity:        repo,
			jobinfo.PullRequestEntity: pull,
		},
	})
}

func IsSameOrigin(pr *github.PullRequest) bool {
	return pr.GetHead().GetRepo().GetCloneURL() == pr.GetBase().GetRepo().GetCloneURL()
}

func toCommonPullRequest(p *github.PullRequest) *common.PullRequest {
	return &common.PullRequest{
		ID:        p.GetNodeID(),
		Number:    p.GetNumber(),
		Title:     p.GetTitle(),
		Body:      p.GetBody(),
		ClosedAt:  p.GetClosedAt(),
		MergedAt:  p.GetMergedAt(),
		Head:      toCommonBranch(p.GetHead()),
		Base:      toCommonBranch(p.GetBase()),
		CreatedAt: p.GetCreatedAt(),
		UpdatedAt: p.GetUpdatedAt(),
	}
}

func toCommonBranch(b *github.PullRequestBranch) *common.Branch {
	return &common.Branch{
		Name:     b.GetRef(),
		SHA:      b.GetSHA(),
		CloneURL: b.GetRepo().GetCloneURL(),
	}
}

// MemberEvent triggered when a user accepts an invitation or is removed as a collaborator
// to a repository, or has their permissions changed.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#memberevent
func (app *GithubApp) memberEvent(ctx context.Context, payload []byte) error {
	var event github.MemberEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	switch event.GetAction() {
	case "added":
		_, err := app.srvClient.Ingest.AddCollaborator(ctx,
			app.provider,
			event.Repo.GetNodeID(),
			&repofuel.AddCollaboratorEvent{
				User: common.User{
					ID: event.Member.GetNodeID(),
				},
				Permissions: common.Permissions{
					Admin: false,
					Read:  true,
					Write: true,
				},
			})
		return err

	case "deleted":
		_, err := app.srvClient.Ingest.DeleteCollaborator(ctx,
			app.provider,
			event.Repo.GetNodeID(),
			event.Member.GetNodeID())
		return err

	default:
		log.Ctx(ctx).Debug().
			Str("action", event.GetAction()).
			Msg("unhandled action for member event")
	}
	return nil
}

// PushEvent triggered on a push to a repository, including branch pushes and repository tag pushes.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#pushevent
func (app *GithubApp) pushEvent(ctx context.Context, payload []byte) error {
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	repo, err := app.repoEntity(ctx, event.Installation.GetID(), event.Sender, event.Repo)
	if err != nil {
		return err
	}

	if repo.IsChecksEnabled() {
		// ignore: checks is enabled for this repository
		return nil
	}

	return app.mgr.ProcessRepository(&jobinfo.JobInfo{
		Action: invoke.ActionRepositoryPush,
		RepoID: repo.ID,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity: repo,
		},
	})
}

type AddableRepository interface {
	GetNodeID() string
	GetID() int64
	GetDefaultBranch() string
	GetName() string
	GetCloneURL() string
	GetHTMLURL() string
	GetOwner() *github.User
}

func (app *GithubApp) repoEntity(ctx context.Context, installationID int64, sender *github.User, r AddableRepository) (*entity.Repository, error) {
	repo, err := app.repoDB.FindByProviderID(ctx, app.provider, r.GetNodeID())
	if err == nil {
		return repo, nil
	}

	if err != entity.ErrRepositoryNotExist {
		return nil, err
	}

	// todo: log this

	switch r := r.(type) {
	case *github.Repository:
		err = app.addInstallationRepositories(ctx, installationID, sender, r.GetOwner(), r)
	case *github.PushEventRepository:
		err = app.addInstallationRepositories(ctx, installationID, sender, r.GetOwner(), &github.Repository{
			ID:              r.ID,
			NodeID:          r.NodeID,
			Owner:           r.Owner,
			Name:            r.Name,
			FullName:        r.FullName,
			Description:     r.Description,
			Homepage:        r.Homepage,
			DefaultBranch:   r.DefaultBranch,
			MasterBranch:    r.MasterBranch,
			CreatedAt:       r.CreatedAt,
			PushedAt:        r.PushedAt,
			UpdatedAt:       r.UpdatedAt,
			HTMLURL:         r.HTMLURL,
			CloneURL:        r.CloneURL,
			GitURL:          r.GitURL,
			SSHURL:          r.SSHURL,
			SVNURL:          r.SVNURL,
			Language:        r.Language,
			Fork:            r.Fork,
			ForksCount:      r.ForksCount,
			OpenIssuesCount: r.OpenIssuesCount,
			StargazersCount: r.StargazersCount,
			WatchersCount:   r.WatchersCount,
			Size:            r.Size,
			Archived:        r.Archived,
			Disabled:        r.Disabled,
			Private:         r.Private,
			HasIssues:       r.HasIssues,
			HasWiki:         r.HasWiki,
			HasPages:        r.HasPages,
			HasDownloads:    r.HasDownloads,
			URL:             r.URL,
			ArchiveURL:      r.ArchiveURL,
			PullsURL:        r.PullsURL,
			StatusesURL:     r.StatusesURL,
			//Organization:    r.Organization,
		})
	default:
		return nil, errors.New("unexpected github repository type")
	}

	if err != nil {
		return nil, err
	}

	return app.repoDB.FindByProviderID(ctx, app.provider, r.GetNodeID())
}

// InstallationEvent triggered when someone installs (created) , uninstalls (deleted),
// or accepts new permissions (new_permissions_accepted) for a GitHub HttpHandler. When a GitHub
// HttpHandler owner requests new permissions, the person who installed the GitHub HttpHandler must accept
// the new permissions request.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#installationevent
func (app *GithubApp) installationEvent(ctx context.Context, payload []byte) error {
	var event github.InstallationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	switch event.GetAction() {
	case "created":
		return app.addInstallationRepositories(ctx, event.Installation.GetID(), event.Installation.Account, event.Sender, event.Repositories...)

	case "deleted":
		return app.removeOrganization(ctx, event.Installation)

	case "new_permissions_accepted":
		// not implemented
		return nil

	default:
		return errors.New("unexpected installation event action")

	}
}

// InstallationRepositoriesEvent triggered when a check suite activity has occurred.
//
// GitHub API docs: https://developer.github.com/webhooks/event-payloads/#check_suite
func (app *GithubApp) checkSuiteEvent(ctx context.Context, payload []byte) error {
	var event github.CheckSuiteEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return err
	}

	switch event.GetAction() {
	case "requested":
		return app.checkSuiteRequestedEvent(ctx, &event)

	case "rerequested", "completed":
		// not implemented
		return nil

	default:
		return errors.New("unexpected check suite event action")

	}
}

const keyChuckID = "github_chuck_id"
const keyChuckName = "github_chucks_name"

func (app *GithubApp) checkSuiteRequestedEvent(ctx context.Context, event *github.CheckSuiteEvent) error {
	repo, err := app.repoEntity(ctx, event.Installation.GetID(), event.Sender, event.Repo)
	if err != nil {
		return err
	}

	if !repo.IsChecksEnabled() {
		// ignore: checks is not enabled for this repository
		return nil
	}

	repoClient := app.installationRepository(
		event.Installation.GetID(),
		event.Repo.Owner.GetLogin(),
		event.Repo.GetName(),
	)

	name := pushCheckName
	action := invoke.ActionPushCheck
	if len(event.CheckSuite.PullRequests) > 0 {
		name = pullRequestCheckName
		action = invoke.ActionPullRequestCheck
	}

	details, err := repoClient.CreateCheckRun(ctx, event.CheckSuite.GetHeadSHA(), name)
	if err != nil {
		// todo: maybe we should not fail if the check run is not created
		return err
	}

	if len(event.CheckSuite.PullRequests) > 0 {
		details[jobinfo.PullRequestNumbers] = PullRequestNumbers(event.CheckSuite.PullRequests)
	} else {
		details[jobinfo.KeyPushBeforeSHA] = event.CheckSuite.GetBeforeSHA()
		details[jobinfo.KeyPushAfterSHA] = event.CheckSuite.GetAfterSHA()
	}

	return app.mgr.ProcessRepository(&jobinfo.JobInfo{
		RepoID:  repo.ID,
		Action:  action,
		Details: details,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity: repo,
		},
	})
}

func PullRequestNumbers(pulls []*github.PullRequest) []int {
	ids := make([]int, len(pulls))
	for i, pull := range pulls {
		ids[i] = pull.GetNumber()
	}
	return ids
}

// InstallationRepositoriesEvent triggered when a repository is added or removed
// from an installation.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/types/#installationrepositoriesevent
func (app *GithubApp) installationRepositoriesEvent(ctx context.Context, payload []byte) error {
	var event github.InstallationRepositoriesEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return err
	}

	err = app.addInstallationRepositories(ctx, event.Installation.GetID(), event.Installation.Account, event.Sender, event.RepositoriesAdded...)
	if err != nil {
		return err
	}

	return app.removeRepositories(ctx, event.RepositoriesRemoved)
}

func (app *GithubApp) removeOrganization(ctx context.Context, installation *github.Installation) error {
	itr, err := app.repoDB.FindByOwnerID(ctx, app.provider, installation.Account.GetNodeID())
	if err != nil {
		return err
	}

	err = itr.ForEach(ctx, func(repo *entity.Repository) error {
		return app.mgr.DeleteRepository(ctx, repo)
	})
	if err != nil {
		return err
	}

	return app.organizationDB.DeleteByOwnerID(ctx, app.provider, installation.Account.GetNodeID())
}

func (app *GithubApp) removeRepositories(ctx context.Context, repos []*github.Repository) error {
	if len(repos) == 0 {
		return nil
	}

	ids := make([]string, len(repos))
	for i, repo := range repos {
		ids[i] = repo.GetNodeID()
	}

	itr, err := app.repoDB.FindByProviderIDs(ctx, app.provider, ids)
	if err != nil {
		return err
	}

	return itr.ForEach(ctx, func(repo *entity.Repository) error {
		return app.mgr.DeleteRepository(ctx, repo)
	})
}

var pullPermissions = map[string]bool{
	"push": true,
}

func (app *GithubApp) addInstallationRepositories(ctx context.Context, installationID int64, account *github.User, addedBy *github.User, repos ...*github.Repository) error {
	if len(repos) == 0 {
		return nil
	}

	ghClient := app.installationClient(installationID)

	org, err := app.creatOrUpdateOrganization(ctx, ghClient, &installationID, account)
	if err != nil {
		return err
	}

	if addedBy != nil {
		addedBy = &github.User{NodeID: addedBy.NodeID, Permissions: &pullPermissions}
	}

	// Add the primary permissions, should update it when process the repository
	collaborators := toCommonPermissions([]*github.User{
		{NodeID: account.NodeID, Permissions: &pullPermissions},
		addedBy,
	})

	for _, repo := range repos {
		if !isCompleteRepoInfo(repo) {
			repo, _, err = ghClient.Repositories.GetByID(ctx, repo.GetID())
			if err != nil {
				return err
			}
		}

		err = app.mgr.AddRepository(ctx, &entity.Repository{
			Organization:  org.ID,
			Source:        *toCommonRepository(repo),
			Owner:         org.Owner,
			ProviderSCM:   org.ProviderSCM,
			ProviderITS:   org.ProviderITS,
			Collaborators: collaborators,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func isCompleteRepoInfo(repo *github.Repository) bool {
	return repo.DefaultBranch != nil && repo.Name != nil && repo.CloneURL != nil && repo.HTMLURL != nil
}

func (app *GithubApp) creatOrUpdateOrganization(ctx context.Context, ghClient *github.Client, installationID *int64, account *github.User) (*entity.Organization, error) {
	orgDraft := &entity.Organization{
		Owner:       toCommonAccount(account),
		ProviderSCM: app.provider,
		ProviderITS: app.provider,
		AvatarURL:   account.GetAvatarURL(),
	}

	if installationID != nil {
		orgDraft.ProvidersConfig = map[string]entity.IntegrationConfig{
			app.provider: &entity.InstallationConfig{
				DatabaseID:     account.GetID(),
				InstallationID: *installationID,
			},
		}
	}

	if orgDraft.Owner.Type == common.AccountPersonal {
		orgDraft.Members = map[string]common.Membership{
			account.GetNodeID(): {Role: common.OrgAdmin},
		}
	}

	org, err := app.organizationDB.FindOrCreate(ctx, orgDraft)
	if err != nil {
		return nil, err
	}

	if org.Owner != orgDraft.Owner {
		org.Owner = orgDraft.Owner
		err = app.organizationDB.UpdateOwner(ctx, org.ID, &org.Owner)
		if err != nil {
			return nil, err
		}
		err = app.repoDB.UpdateOwner(ctx, org.ID, &org.Owner)
		if err != nil {
			return nil, err
		}
	}

	if len(org.Members) > 0 || org.Owner.Type != common.AccountOrganization || installationID == nil {
		// the organization is exist before
		return org, nil
	}

	// todo: we could store the members in a collection with their details
	members := make(map[string]common.Membership)
	err = listAllMembers(ctx, ghClient, org.Owner.Slug, "admin", func(users []*github.User) error {
		for i := range users {
			members[users[i].GetNodeID()] = common.Membership{Role: common.OrgAdmin}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = listAllMembers(ctx, ghClient, org.Owner.Slug, "member", func(users []*github.User) error {
		for i := range users {
			members[users[i].GetNodeID()] = common.Membership{Role: common.OrgMember}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = app.organizationDB.UpdateMembers(ctx, org.ID, members)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func listAllCollaborators(ctx context.Context, ghClient *github.Client, owner, repo string, fn func([]*github.User) error) error {
	users, resp, err := ghClient.Repositories.ListCollaborators(ctx, owner, repo, &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return err
	}

	err = fn(users)
	if err != nil {
		return err
	}

	for resp.NextPage < resp.LastPage {
		users, resp, err = ghClient.Repositories.ListCollaborators(ctx, owner, repo, &github.ListCollaboratorsOptions{
			ListOptions: github.ListOptions{
				Page:    resp.NextPage,
				PerPage: 100,
			},
		})
		if err != nil {
			return err
		}

		err = fn(users)
		if err != nil {
			return err
		}
	}
	return err
}

func listAllMembers(ctx context.Context, ghClient *github.Client, org, role string, fn func([]*github.User) error) error {
	users, resp, err := ghClient.Organizations.ListMembers(ctx, org, &github.ListMembersOptions{
		PublicOnly: false,
		Role:       role,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return err
	}

	err = fn(users)
	if err != nil {
		return err
	}

	for resp.NextPage < resp.LastPage {
		users, resp, err = ghClient.Organizations.ListMembers(ctx, org, &github.ListMembersOptions{
			PublicOnly: false,
			Role:       role,
			ListOptions: github.ListOptions{
				Page:    resp.NextPage,
				PerPage: 100,
			},
		})
		if err != nil {
			return err
		}
		err = fn(users)
		if err != nil {
			return err
		}
	}
	return err
}

func listAllInstallations(ctx context.Context, ghClient *github.Client, fn func([]*github.Installation) error) error {
	installations, resp, err := ghClient.Apps.ListInstallations(ctx, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return err
	}

	err = fn(installations)
	if err != nil {
		return err
	}

	for resp.NextPage < resp.LastPage {
		installations, resp, err = ghClient.Apps.ListInstallations(ctx, &github.ListOptions{
			Page:    resp.NextPage,
			PerPage: 100,
		})
		if err != nil {
			return err
		}
		err = fn(installations)
		if err != nil {
			return err
		}
	}
	return err
}

func listAllRepositoriesForInstallation(ctx context.Context, ghClient *github.Client, fn func([]*github.Repository) error) error {
	repos, resp, err := ghClient.Apps.ListRepos(ctx, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return err
	}

	err = fn(repos)
	if err != nil {
		return err
	}

	for resp.NextPage < resp.LastPage {
		repos, resp, err = ghClient.Apps.ListRepos(ctx, &github.ListOptions{
			Page:    resp.NextPage,
			PerPage: 100,
		})
		if err != nil {
			return err
		}
		err = fn(repos)
		if err != nil {
			return err
		}
	}
	return err
}

func (app *GithubApp) AddAllRepository(ctx context.Context) error {
	appClient := app.ghClient(app.appsTransport)

	return listAllInstallations(ctx, appClient, func(installations []*github.Installation) error {
		for _, installation := range installations {
			c := app.installationClient(installation.GetID())
			err := listAllRepositoriesForInstallation(ctx, c, func(repositories []*github.Repository) error {
				return app.addInstallationRepositories(ctx, installation.GetID(), installation.Account, nil, repositories...)
			})
			if err != nil {
				return err
			}
			//temporary
			time.Sleep(5 * time.Second)
		}
		return nil
	})
}
