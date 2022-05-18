package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/graph/generated"
	"github.com/repofuel/repofuel/ingest/graph/model"
	"github.com/repofuel/repofuel/ingest/internal/accesscontrol"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/pkg/common"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *activityResolver) RepositoriesTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.RepositoryDB.TotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) RepositoriesCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.RepositoryDB.CountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) OrganizationsTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.OrganizationDB.TotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) OrganizationsCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.OrganizationDB.CountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) CommitsAnalyzedTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.CommitDB.AnalyzedTotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) CommitsAnalyzedCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.CommitDB.AnalyzedCountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) CommitsPredictTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.CommitDB.PredictedTotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) CommitsPredictCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.CommitDB.PredictedCountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) JobsTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.JobDB.TotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) JobsCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.JobDB.CountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) PullRequestAnalyzedTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.PullRequestDB.AnalyzedTotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) PullRequestAnalyzedCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.CountOverTimeConnection, error) {
	nodes, err := r.PullRequestDB.AnalyzedCountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *activityResolver) ViewsTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.VisitDB.ViewsTotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) VisitorsTotalCount(ctx context.Context, obj *model.Activity, period *model.Period) (int, error) {
	res, err := r.VisitDB.VisitorsTotalCount(ctx, TimeFromPeriod(period))
	return int(res), err
}

func (r *activityResolver) VisitCount(ctx context.Context, obj *model.Activity, period *model.Period, frequency entity.Frequency) (*model.VisitOverTimeConnection, error) {
	nodes, err := r.VisitDB.CountOverTime(ctx, TimeFromPeriod(period), frequency)
	if err != nil {
		return nil, err
	}

	return &model.VisitOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *commitResolver) Hash(ctx context.Context, obj *entity.Commit) (string, error) {
	return obj.ID.CommitHash.Hex(), nil
}

func (r *commitResolver) Tags(ctx context.Context, obj *entity.Commit) ([]string, error) {
	return tagsToStrings(obj.Tags), nil
}

func (r *commitResolver) DeletedTags(ctx context.Context, obj *entity.Commit) ([]string, error) {
	return tagsToStrings(obj.DeletedTags), nil
}

func (r *commitResolver) Fixes(ctx context.Context, obj *entity.Commit, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.CommitConnection, error) {
	ids := make([]identifier.Hash, len(obj.Fixes))
	for i := range obj.Fixes {
		ids[i] = identifier.NewHash(obj.Fixes[i])
	}

	return r.CommitDB.SelectedCommitConnection(ctx, obj.ID.RepoID, ids, direction, &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	})
}

func (r *commitResolver) Repository(ctx context.Context, obj *entity.Commit) (*entity.Repository, error) {
	return r.RepositoryDB.FindByID(ctx, obj.ID.RepoID)
}

func (r *commitFileResolver) Type(ctx context.Context, obj *entity.File) (*string, error) {
	if obj.Type == 0 {
		return nil, nil
	}

	s := obj.Type.String()
	return &s, nil
}

func (r *commitFileResolver) Fixing(ctx context.Context, obj *entity.File) ([]*entity.Commit, error) {
	if len(obj.Fixing) == 0 {
		return nil, nil
	}

	c := graphql.GetFieldContext(ctx).Parent.Parent.Parent.Result.(*entity.Commit)

	itr, err := r.CommitDB.FindCommitsByHash(ctx, c.ID.RepoID, obj.Fixing...)
	if err != nil {
		return nil, err
	}
	return itr.Slice(ctx)
}

func (r *feedbackResolver) Sender(ctx context.Context, obj *entity.Feedback) (*model.User, error) {
	//todo: Fetch more information (from the accounts service) if more fields are required from the GraphQL API
	return &model.User{
		ID: obj.Sender,
	}, nil
}

func (r *feedbackResolver) Target(ctx context.Context, obj *entity.Feedback) (*entity.Commit, error) {
	return r.CommitDB.FindByID(ctx, &obj.CommitID)
}

func (r *mutationResolver) UpdateRepository(ctx context.Context, input model.UpdateRepositoryInput) (*model.UpdateRepositoryPayload, error) {
	repoID, err := identifier.RepositoryIDFromNodeID(input.ID)
	if err != nil {
		return nil, err
	}

	repo, err := r.RepositoryDB.FindAndUpdateChecksConfig(ctx, repoID, (*entity.ChecksConfig)(input.ChecksConfig))
	if err != nil {
		return nil, err
	}

	return &model.UpdateRepositoryPayload{
		Repository: repo,
		Errors:     nil,
	}, err
}

func (r *mutationResolver) SendCommitFeedback(ctx context.Context, input model.SendCommitFeedbackInput) (*entity.Feedback, error) {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return nil, errors.New("unauthorized")
	}

	feedback := &entity.Feedback{
		Sender:  identifier.UserID(viewer.UserID),
		Message: input.Message,
	}

	if err := feedback.CommitID.UnmarshalGQL(input.CommitID); err != nil {
		return nil, err
	}

	if err := r.FeedbackDB.Insert(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

func (r *mutationResolver) AddPublicRepository(ctx context.Context, input model.AddPublicRepositoryInput) (*model.AddPublicRepositoryPayload, error) {
	nameWithOwner := strings.Split(strings.TrimSuffix(input.NameWithOwner, ".git"), "/")
	if len(nameWithOwner) != 2 {
		return nil, errors.New("unexpected repository url")
	}

	p, err := r.Manager.Integrations.ServiceProvider(ctx, input.Provider)
	if err != nil {
		return nil, err
	}

	pm, ok := p.(manage.MonitorProvider)
	if !ok {
		return nil, errors.New("monitor public repositories not supported for this provider")
	}

	err = pm.MonitorRepository(ctx, permission.ViewerCtx(ctx).UserID, nameWithOwner[0], nameWithOwner[1])
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return nil, err
	}

	repo, err := r.RepositoryDB.FindByName(ctx, input.Provider, nameWithOwner[0], nameWithOwner[1])
	if err != nil {
		return nil, err
	}

	return &model.AddPublicRepositoryPayload{
		Repository: repo,
	}, nil
}

func (r *mutationResolver) StopRepositoryMonitoring(ctx context.Context, id string) (*model.StopRepositoryMonitoringPayload, error) {
	repoID, err := NodeIdToRepoId(id)
	if err != nil {
		return nil, err
	}

	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return nil, errors.New("currently only site admins can delete repositories")
	}

	err = r.MonitorDB.RemoveMonitor(ctx, &identifier.MonitorID{
		RepoID: repoID,
		UserID: identifier.UserID(viewer.UserID),
	})
	if err != nil {
		return nil, err
	}

	repo, err := r.RepositoryDB.FindByID(ctx, repoID)

	return &model.StopRepositoryMonitoringPayload{
		Repository: repo,
	}, err
}

func (r *mutationResolver) MonitorRepository(ctx context.Context, id string) (*model.MonitorRepositoryPayload, error) {
	repoID, err := NodeIdToRepoId(id)
	if err != nil {
		return nil, err
	}

	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return nil, errors.New("currently only site admins can delete repositories")
	}

	r.MonitorDB.InsertMonitor(ctx, &identifier.MonitorID{
		RepoID: repoID,
		UserID: identifier.UserID(viewer.UserID),
	})

	repo, err := r.RepositoryDB.FindByID(ctx, repoID)

	return &model.MonitorRepositoryPayload{
		Repository: repo,
	}, err
}

func (r *mutationResolver) DeleteRepository(ctx context.Context, id string) (*model.DeleteRepositoryPayload, error) {
	viewer := permission.ViewerCtx(ctx)
	if viewer.Role != permission.RoleSiteAdmin {
		return nil, errors.New("currently only site admins can delete repositories")
	}

	repoID, err := NodeIdToRepoId(id)
	if err != nil {
		return nil, err
	}

	repo, err := r.RepositoryDB.FindByID(ctx, repoID)
	if err != nil {
		return nil, err
	}

	err = r.Manager.DeleteRepository(ctx, repo)
	if err != nil {
		return nil, err
	}

	return &model.DeleteRepositoryPayload{
		Repository: repo,
	}, nil
}

func (r *mutationResolver) DeleteCommitTag(ctx context.Context, input model.DeleteCommitTagInput) (*model.DeleteCommitTagPayload, error) {
	//todo: check the authorization
	return nil, errors.New("'deleteCommitTag' mutation is disabled")

	commitID, err := NodeIdToCommitId(input.CommitID)
	if err != nil {
		return nil, err
	}

	err = r.CommitDB.DeleteCommitTag(ctx, commitID, input.Tag)
	if err != nil {
		return nil, err
	}

	commit, err := r.CommitDB.FindByID(ctx, commitID)

	return &model.DeleteCommitTagPayload{
		Commit: commit,
	}, err
}

func (r *organizationResolver) ProviderSetupURL(ctx context.Context, obj *entity.Organization) (*string, error) {
	p, err := r.Manager.Integrations.ServiceProvider(ctx, obj.ProviderSCM)
	if err != nil {
		return nil, err
	}

	url, err := p.SetupURL(obj)
	if err != nil {
		if strings.HasSuffix(err.Error(), "not implemented") {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *organizationResolver) Repositories(ctx context.Context, obj *entity.Organization, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.RepositoryConnection, error) {
	return r.RepositoryDB.FindOrgReposConnection(ctx, obj.ID, direction,
		&entity.PaginationInput{
			First:  first,
			After:  after,
			Last:   last,
			Before: before,
		},
	)
}

func (r *organizationResolver) ViewerCanAdminister(ctx context.Context, obj *entity.Organization) (bool, error) {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return false, errors.New("unauthorized")
	}

	userId := viewer.UserInfo.Providers[obj.ProviderSCM]
	return obj.Members[userId].Role == common.OrgAdmin, nil
}

func (r *pullRequestResolver) Progress(ctx context.Context, obj *entity.PullRequest) (*manage.Progress, error) {
	po := r.Observables.Get(obj.ID.NodeID())
	if po == nil {
		return nil, nil
	}
	return po.Progress(), nil
}

func (r *pullRequestResolver) Commits(ctx context.Context, obj *entity.PullRequest, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.CommitConnection, error) {
	return r.CommitDB.PullRequestCommitConnection(ctx, obj.RepoID, obj.ID, direction, &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	})
}

func (r *pullRequestSourceResolver) Head(ctx context.Context, obj *common.PullRequest) (*entity.Branch, error) {
	//todo: should be automatically binned
	return (*entity.Branch)(obj.Head), nil
}

func (r *pullRequestSourceResolver) Base(ctx context.Context, obj *common.PullRequest) (*entity.Branch, error) {
	//todo: should be automatically binned
	return (*entity.Branch)(obj.Base), nil
}

func (r *queryResolver) Viewer(ctx context.Context) (*model.User, error) {
	var user model.User
	fetchFromAccounts := false
	fields := graphql.CollectAllFields(ctx)
	for _, f := range fields {
		if fieldsFromAccount[f] {
			fetchFromAccounts = true
			break
		}
	}

	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return nil, errors.New("unauthorized")
	}

	if !fetchFromAccounts {
		user.ID = identifier.UserID(viewer.UserID)

		//fixme: this could cause inconstancy in the returned data (based on the query) if the JWT have diffrent data than the DB (it will be fix automaticy when the JWT expires)
		providers := make([]*common.User, 0, len(viewer.Providers))
		for provider, userID := range viewer.Providers {
			providers = append(providers, &common.User{
				//todo: if the GraphQL request more fields, it should be fetched from accounts service
				Provider: provider,
				ID:       userID,
			})
		}
		user.Providers = providers

		return &user, nil
	}

	url, err := r.RepofuelClient.Accounts.BaseURL.Parse(fmt.Sprintf("users/%s", viewer.UserID.Hex()))
	if err != nil {
		return nil, err
	}

	req, err := r.RepofuelClient.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	_, err = r.RepofuelClient.Do(req, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *queryResolver) Repository(ctx context.Context, provider string, owner string, name string) (*entity.Repository, error) {
	repo, err := r.RepositoryDB.FindByName(ctx, provider, owner, name)
	if err != nil {
		return nil, err
	}

	//fixme: this is a hack
	if !accesscontrol.UserPermissions(context.WithValue(ctx, accesscontrol.RepositoryCtxKey, repo)).Read {
		return nil, errors.New("unauthorized")
	}

	return repo, nil
}

func (r *queryResolver) Repositories(ctx context.Context, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.RepositoryConnection, error) {
	page := &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	}

	return r.RepositoryDB.FindAllReposConnection(ctx, direction, page)
}

func (r *queryResolver) Organizations(ctx context.Context, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.OrganizationConnection, error) {
	return r.OrganizationDB.FindAllOrgsConnection(ctx, direction, &entity.PaginationInput{
		Before: before,
		After:  after,
		First:  first,
		Last:   last,
	})
}

func (r *queryResolver) Organization(ctx context.Context, provider string, owner string) (*entity.Organization, error) {
	return r.OrganizationDB.FindBySlug(ctx, provider, owner)
}

func (r *queryResolver) Node(ctx context.Context, id string) (model.Node, error) {
	return r.Manager.Node(ctx, id)
}

func (r *queryResolver) Activity(ctx context.Context) (*model.Activity, error) {
	//fixme: this is a hack
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.Role != permission.RoleSiteAdmin {
		return nil, errors.New("unauthorized")
	}

	return &model.Activity{}, nil
}

func (r *queryResolver) Feedback(ctx context.Context, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.FeedbackConnection, error) {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.Role != permission.RoleSiteAdmin {
		return nil, errors.New("unauthorized")
	}

	return r.FeedbackDB.FeedbackConnection(direction, &entity.PaginationInput{
		Before: before,
		After:  after,
		First:  first,
		Last:   last,
	}), nil
}

func (r *repositoryResolver) DatabaseID(ctx context.Context, obj *entity.Repository) (string, error) {
	return obj.ID.Hex(), nil
}

func (r *repositoryResolver) Name(ctx context.Context, obj *entity.Repository) (string, error) {
	return obj.Source.RepoName, nil
}

func (r *repositoryResolver) Commit(ctx context.Context, obj *entity.Repository, hash string) (*entity.Commit, error) {
	return r.CommitDB.FindByID(ctx, identifier.NewCommitID(obj.ID, identifier.NewHash(hash)))
}

func (r *repositoryResolver) PullRequest(ctx context.Context, obj *entity.Repository, number int) (*entity.PullRequest, error) {
	return r.PullRequestDB.FindByNumber(ctx, obj.ID, number)
}

func (r *repositoryResolver) Progress(ctx context.Context, obj *entity.Repository) (*manage.Progress, error) {
	po := r.Observables.Get(obj.ID.NodeID())
	if po == nil {
		return nil, nil
	}
	return po.Progress(), nil
}

func (r *repositoryResolver) ViewerIsMonitor(ctx context.Context, obj *entity.Repository) (bool, error) {
	//todo: shoud we have a utility function to return the user info or error
	info := permission.ViewerCtx(ctx)
	if info == nil || info.UserInfo == nil {
		return false, errors.New("messing user information")
	}

	return r.MonitorDB.IsMonitor(ctx, &identifier.MonitorID{
		RepoID: obj.ID,
		UserID: identifier.UserID(info.UserID),
	})
}

func (r *repositoryResolver) MonitorCount(ctx context.Context, obj *entity.Repository) (int, error) {
	return r.MonitorDB.MonitorCount(ctx, obj.ID)
}

func (r *repositoryResolver) Commits(ctx context.Context, obj *entity.Repository, first *int, after *string, last *int, before *string, direction *entity.OrderDirection, filters *entity.CommitFilters) (entity.CommitConnection, error) {
	if filters == nil {
		filters = &entity.CommitFilters{}
	}
	if filters.Branch == nil {
		//fixme: could be problematic when we implement the auto projection for graphql
		filters.Branch = &obj.Source.DefaultBranch
	}

	return r.CommitDB.RepositoryCommitConnection(ctx, obj.ID, direction, filters, &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	})
}

func (r *repositoryResolver) PullRequests(ctx context.Context, obj *entity.Repository, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.PullRequestConnection, error) {
	return r.PullRequestDB.RepositoryPullRequestConnection(obj.ID, direction, &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	}), nil
}

func (r *repositoryResolver) Jobs(ctx context.Context, obj *entity.Repository, first *int, after *string, last *int, before *string, direction *entity.OrderDirection) (entity.JobConnection, error) {
	//fixme: this is a hack
	if !accesscontrol.UserPermissions(context.WithValue(ctx, accesscontrol.RepositoryCtxKey, obj)).Admin {
		return nil, errors.New("unauthorized")
	}

	return r.JobDB.RepositoryJobConnection(obj.ID, direction, &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	}), nil
}

func (r *repositoryResolver) Branches(ctx context.Context, obj *entity.Repository) ([]*entity.Branch, error) {
	return hashToBranches(obj.Branches), nil
}

func (r *repositoryResolver) DeveloperEmails(ctx context.Context, obj *entity.Repository) ([]string, error) {
	return r.CommitDB.DeveloperEmails(ctx, obj.ID)
}

func (r *repositoryResolver) DeveloperNames(ctx context.Context, obj *entity.Repository) ([]string, error) {
	return r.CommitDB.DeveloperNames(ctx, obj.ID)
}

func (r *repositoryResolver) PredictionStatus(ctx context.Context, obj *entity.Repository) (*int, error) {
	v := int(obj.Quality)
	return &v, nil
}

func (r *repositoryResolver) CommitPredictionsCount(ctx context.Context, obj *entity.Repository) (*int, error) {
	res, err := r.CommitDB.RepositoryPredictionsTotalCount(ctx, obj.ID)
	i := int(res)
	return &i, err
}

func (r *repositoryResolver) BuggyCommitsCount(ctx context.Context, obj *entity.Repository) (*int, error) {
	return &obj.BuggyCount, nil
}

func (r *repositoryResolver) FixCommitsCount(ctx context.Context, obj *entity.Repository) (*int, error) {
	//todo: not implemented
	return nil, nil
}

func (r *repositoryResolver) BranchesCount(ctx context.Context, obj *entity.Repository) (*int, error) {
	bn := len(obj.Branches)
	return &bn, nil
}

func (r *repositoryResolver) ContributorsCount(ctx context.Context, obj *entity.Repository) (*int, error) {
	count, err := r.CommitDB.ContributorsCount(ctx, obj.ID)
	return &count, err
}

func (r *repositoryResolver) BuggyCommitsOverTime(ctx context.Context, obj *entity.Repository) (*model.CountOverTimeConnection, error) {
	nodes, err := r.CommitDB.BuggyCommitsOverTime(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *repositoryResolver) CommitsOverTime(ctx context.Context, obj *entity.Repository) (*model.CountOverTimeConnection, error) {
	nodes, err := r.CommitDB.CommitsOverTime(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return &model.CountOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *repositoryResolver) TagsCount(ctx context.Context, obj *entity.Repository) (*model.TagsCountConnection, error) {
	nodes, err := r.CommitDB.CommitsTagCount(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return &model.TagsCountConnection{
		Nodes: nodes,
	}, nil
}

func (r *repositoryResolver) AvgEntropyOverTime(ctx context.Context, obj *entity.Repository) (*model.AvgOverTimeConnection, error) {
	nodes, err := r.CommitDB.AvgEntropyOverTime(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return &model.AvgOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *repositoryResolver) AvgCommitFilesOverTime(ctx context.Context, obj *entity.Repository) (*model.AvgOverTimeConnection, error) {
	nodes, err := r.CommitDB.AvgCommitFilesOverTime(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return &model.AvgOverTimeConnection{
		Nodes: nodes,
	}, nil
}

func (r *repositoryResolver) ViewerCanAdminister(ctx context.Context, obj *entity.Repository) (bool, error) {
	//todo: should not rely on the context to get the repository
	return accesscontrol.UserPermissions(context.WithValue(ctx, accesscontrol.RepositoryCtxKey, obj)).Admin, nil
}

func (r *repositorySourceResolver) URL(ctx context.Context, obj *common.Repository) (string, error) {
	return obj.HTMLURL, nil
}

func (r *subscriptionResolver) ChangeProgress(ctx context.Context, ids []string) (<-chan *manage.ProgressObservable, error) {
	obs := make(chan *manage.ProgressObservable)

	for i, id := range ids {
		po, err := r.Observables.GetOrCreateStateful(ctx, id)
		if err != nil {
			r.RemoveObserversOnCancel(ctx, obs, ids[:i])
			return nil, err
		}

		po.AddObserver(obs)
	}

	r.RemoveObserversOnCancel(ctx, obs, ids)
	return obs, nil
}

func (r *userResolver) Repositories(ctx context.Context, obj *model.User, first *int, after *string, last *int, before *string, direction *entity.OrderDirection, ownerAffiliations []entity.RepositoryAffiliation) (entity.RepositoryConnection, error) {
	page := &entity.PaginationInput{
		First:  first,
		After:  after,
		Last:   last,
		Before: before,
	}

	affiliations := []*entity.UserAffiliationInput{{
		UserID:       obj.ID,
		Providers:    obj.Providers,
		Affiliations: ownerAffiliations,
	}}

	return r.RepositoryDB.FindUserReposConnection(ctx, affiliations, direction, page)
}

// Activity returns generated.ActivityResolver implementation.
func (r *Resolver) Activity() generated.ActivityResolver { return &activityResolver{r} }

// Commit returns generated.CommitResolver implementation.
func (r *Resolver) Commit() generated.CommitResolver { return &commitResolver{r} }

// CommitFile returns generated.CommitFileResolver implementation.
func (r *Resolver) CommitFile() generated.CommitFileResolver { return &commitFileResolver{r} }

// Feedback returns generated.FeedbackResolver implementation.
func (r *Resolver) Feedback() generated.FeedbackResolver { return &feedbackResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Organization returns generated.OrganizationResolver implementation.
func (r *Resolver) Organization() generated.OrganizationResolver { return &organizationResolver{r} }

// PullRequest returns generated.PullRequestResolver implementation.
func (r *Resolver) PullRequest() generated.PullRequestResolver { return &pullRequestResolver{r} }

// PullRequestSource returns generated.PullRequestSourceResolver implementation.
func (r *Resolver) PullRequestSource() generated.PullRequestSourceResolver {
	return &pullRequestSourceResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Repository returns generated.RepositoryResolver implementation.
func (r *Resolver) Repository() generated.RepositoryResolver { return &repositoryResolver{r} }

// RepositorySource returns generated.RepositorySourceResolver implementation.
func (r *Resolver) RepositorySource() generated.RepositorySourceResolver {
	return &repositorySourceResolver{r}
}

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type activityResolver struct{ *Resolver }
type commitResolver struct{ *Resolver }
type commitFileResolver struct{ *Resolver }
type feedbackResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type organizationResolver struct{ *Resolver }
type pullRequestResolver struct{ *Resolver }
type pullRequestSourceResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type repositoryResolver struct{ *Resolver }
type repositorySourceResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
