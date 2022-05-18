// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package rest

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/accesscontrol"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/atlassian"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/ingest/pkg/providers"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/metrics"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog"
)

var (
	ErrMissedUserInfo = errors.New("access info should specify the user")
)

//deprecated
type Integration interface {
	// Provider returns the name of the provider, e.g., github or bitbucket
	Provider() common.Name
	SourceIntegration(owner, repo string, installation common.Installation) providers.SourceIntegration
	//deprecated
	AvatarURL(username string) string
}

type Handler struct {
	logger          *zerolog.Logger
	reposDB         entity.RepositoryDataSource
	commitsDB       entity.CommitDataSource
	jobsDB          entity.JobDataSource
	pullsDB         entity.PullRequestDataSource
	organizationsDB entity.OrganizationDataSource
	feedbackDB      entity.FeedbackDataSource
	mgr             *manage.Manager
	auth            *jwtauth.AuthCheck
	jiraMgr         *atlassian.JiraIntegrationManager
}

func NewHandler(logger *zerolog.Logger, auth *jwtauth.AuthCheck, mgr *manage.Manager, r entity.RepositoryDataSource, c entity.CommitDataSource, j entity.JobDataSource, p entity.PullRequestDataSource, o entity.OrganizationDataSource, feed entity.FeedbackDataSource, jiraMgr *atlassian.JiraIntegrationManager) *Handler {
	return &Handler{
		logger:          logger,
		reposDB:         r,
		commitsDB:       c,
		jobsDB:          j,
		pullsDB:         p,
		organizationsDB: o,
		feedbackDB:      feed,
		mgr:             mgr,
		auth:            auth,
		jiraMgr:         jiraMgr,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(h.auth.Middleware)

	r.Get("/health", health)
	r.Get("/user/orgs", h.MyOrganizations)

	r.With(permission.OnlyAdmin).Get("/download/feedback.csv", h.FeedbackCSV)

	r.Post("/integrations/jira/check_url", h.jiraMgr.HandelCheckServerInfo)
	r.Route("/organizations/{org_id}/integrations", func(r chi.Router) {
		r.Use(
			accesscontrol.CtxOrganizationByID(h.organizationsDB),
			accesscontrol.OnlyOrganizationAdmin,
		)

		r.Get("/", h.ListOrganizationIntegrations)
		r.Delete("/{provider_id}", h.DeleteIntegration)
	})

	r.Route("/repositories/{repo_id}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(permission.OnlyServiceAccounts)

			r.Get("/metrics.csv", h.MetricsCSV)
			r.Get("/file_aggregated_metrics.csv", h.FileAggregatedMetricsCSV)
			r.Get("/developer_aggregated_metrics.csv", h.DeveloperAggregatedMetricsCSV)
		})

		r.Group(func(r chi.Router) {
			r.Use(h.ctxRepositoryByID)
			r.Use(accesscontrol.OnlyRepositoryAdmin)

			r.Get("/update", h.UpdateRepositorySource)
			r.Get("/process/trigger", h.Fetch)
			r.Get("/process/stop", h.Stop)
		})
	})

	r.Route("/platforms/{platform}", func(r chi.Router) {
		//todo: should be removed after refactor the github integration to remove the dependency of the REST API
		r.Route("/repositories/{source_id}", func(r chi.Router) {

			r.Use(permission.OnlyServiceAccounts)

			r.Post("/collaborators", h.AddCollaborator)
			r.Delete("/collaborators/{source_user_id}", h.DeleteCollaborator)
		})
	})

	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(`{"ok": true}`)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) ListOrganizationIntegrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org := ctx.Value(accesscontrol.OrganizationCtxKey).(*entity.Organization)

	providers := make(map[string]entity.ConfigType, len(org.ProvidersConfig))
	for id, providerConfig := range org.ProvidersConfig {
		providers[id] = providerConfig.ConfigType()
	}

	err := json.NewEncoder(w).Encode(providers)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org := ctx.Value(accesscontrol.OrganizationCtxKey).(*entity.Organization)

	var err = h.organizationsDB.DeleteProviderConfig(ctx, org.ID, chi.URLParam(r, "provider_id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
}

func (h *Handler) MyOrganizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	viewer := permission.ViewerCtx(ctx)

	itr, err := myOrganizations(ctx, h.organizationsDB, viewer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	encoder := json.NewEncoder(w)
	err = itr.ForEach(ctx, func(org *entity.Organization) error {
		return encoder.Encode(org)
	})
	if err != nil {
		log.Println(err)
	}
}

func myOrganizations(ctx context.Context, organizationsDB entity.OrganizationDataSource, accessInfo *permission.AccessInfo) (entity.OrganizationIter, error) {
	if accessInfo.Role == permission.RoleSiteAdmin {
		return organizationsDB.All(ctx)
	}

	if accessInfo.UserInfo == nil {
		return nil, ErrMissedUserInfo
	}

	return organizationsDB.ListUserOrganizations(ctx, accessInfo.UserInfo.Providers)
}

func (h *Handler) UpdateRepositorySource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repo := ctx.Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)

	scm, err := h.mgr.Integrations.RepositorySCM(ctx, repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	//todo: codider update the collaborates
	updated, err := scm.FetchRepositoryInfo(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	//todo: should consider update the collaborates

	err = h.reposDB.UpdateSource(ctx, repo.ID, updated)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func (h *Handler) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	var event repofuel.AddCollaboratorEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = h.reposDB.AddCollaborator(r.Context(),
		chi.URLParam(r, "platform"),
		chi.URLParam(r, "source_id"),
		event.User.ID, event.Permissions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func (h *Handler) DeleteCollaborator(w http.ResponseWriter, r *http.Request) {
	err := h.reposDB.DeleteCollaborator(r.Context(),
		chi.URLParam(r, "platform"),
		chi.URLParam(r, "source_id"),
		chi.URLParam(r, "source_user_id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)

	h.mgr.StopRepository(repo.ID)

	ctx := r.Context()
	for {
		// wait until the progress return an error, which means it is done
		p := h.mgr.Progress(repo.ID)
		if p == nil {
			return
		}

		select {
		case <-ctx.Done():
			// request canceled, stop checking
			log.Println("context for process cancellation:", ctx.Err())
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (h *Handler) Fetch(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)

	err := h.mgr.ProcessRepository(&jobinfo.JobInfo{
		Action: invoke.ActionRepositoryAdminTrigger,
		RepoID: repo.ID,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity: repo,
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) ctxRepositoryByName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		repo, err := h.reposDB.FindByName(ctx,
			chi.URLParam(r, "platform"),
			chi.URLParam(r, "owner"),
			chi.URLParam(r, "repo"))
		if err != nil {
			if err == entity.ErrRepositoryNotExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx = context.WithValue(ctx, accesscontrol.RepositoryCtxKey, repo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) ctxRepositoryByID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := identifier.RepositoryIDFromHex(chi.URLParam(r, "repo_id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		repo, err := h.reposDB.FindByID(ctx, id)
		if err != nil {
			if err == entity.ErrRepositoryNotExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx = context.WithValue(ctx, accesscontrol.RepositoryCtxKey, repo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) DeveloperAggregatedMetricsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// todo: simplified the id retrieval, maybe make func for each route and keep here the common logic
	repoID, err := identifier.RepositoryIDFromHex(chi.URLParam(r, "repo_id"))
	if err != nil {
		repo, ok := ctx.Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)
		if !ok {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		repoID = repo.ID
	}

	iter, err := h.commitsDB.DevelopersAggregatedMetrics(ctx, repoID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Disposition", "attachment; filename=developer_aggregated_metrics.csv")

	err = writeChangeMeasuresCSV(ctx, w, iter)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) FileAggregatedMetricsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// todo: simplified the id retrieval, maybe make func for each route and keep here the common logic
	repoID, err := identifier.RepositoryIDFromHex(chi.URLParam(r, "repo_id"))
	if err != nil {
		repo, ok := ctx.Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)
		if !ok {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		repoID = repo.ID
	}

	iter, err := h.commitsDB.FileAggregatedMetrics(ctx, repoID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Disposition", "attachment; filename=file_aggregated_metrics.csv")

	err = writeFileMeasuresCSV(ctx, w, iter)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) FeedbackCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	iter, err := h.feedbackDB.All(ctx)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Disposition", "attachment; filename=repofuel_feedback.csv")

	err = writeFeedbackCSV(ctx, w, iter)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) MetricsCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// todo: simplified the id retrieval, maybe make func for each route and keep here the common logic
	repoID, err := identifier.RepositoryIDFromHex(chi.URLParam(r, "repo_id"))
	if err != nil {
		repo, ok := ctx.Value(accesscontrol.RepositoryCtxKey).(*entity.Repository)
		if !ok {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		repoID = repo.ID
	}

	startJob := r.FormValue("start_job")
	lastJob := r.FormValue("last_job")
	iter, err := h.findCommits(ctx, repoID, startJob, lastJob)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	disp := fmt.Sprintf("attachment; filename=%s_commit_metrics.csv", repoID.Hex())

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Disposition", disp)

	err = writeCommitsCSV(ctx, w, iter, &writeCommitsCSVOptions{
		FullDumb: r.FormValue("full_dump") == "true",
	})
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) findCommits(ctx context.Context, repoID identifier.RepositoryID, startJob, lastJob string) (entity.CommitIter, error) {
	if startJob == "" && lastJob == "" {
		return h.commitsDB.FindRepoCommits(ctx, repoID)
	}

	if startJob == lastJob {
		jobID, err := identifier.JobIDFromHex(lastJob)
		if err != nil {
			return nil, err
		}
		return h.commitsDB.FindJobCommits(ctx, repoID, jobID)
	}

	if startJob == "" && lastJob != "" {
		jobID, err := identifier.JobIDFromHex(lastJob)
		if err != nil {
			return nil, err
		}
		return h.commitsDB.FindCommitsUntil(ctx, repoID, jobID)
	}

	if startJob != "" && lastJob != "" {
		start, err := identifier.JobIDFromHex(startJob)
		if err != nil {
			return nil, err
		}
		end, err := identifier.JobIDFromHex(lastJob)
		if err != nil {
			return nil, err
		}
		return h.commitsDB.FindCommitsBetween(ctx, repoID, start, end)
	}

	return nil, errors.New("this case is not handled")
}

var _ChangeMeasuresFields = structKeys(metrics.ChangeMeasures{})
var _FileMeasuresFields = structKeys(metrics.FileMeasures{})
var _FeedbackFields = structKeys(entity.Feedback{})

func writeChangeMeasuresCSV(ctx context.Context, w io.Writer, iter entity.ChangeMeasuresIter) error {
	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write(_ChangeMeasuresFields); err != nil {
		return err
	}

	err := iter.ForEach(ctx, func(m *metrics.ChangeMeasures) error {
		return csvWriter.Write(structValues(m))
	})
	if err != nil {
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func writeFileMeasuresCSV(ctx context.Context, w io.Writer, iter entity.FileMeasuresIter) error {
	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write(_FileMeasuresFields); err != nil {
		return err
	}

	err := iter.ForEach(ctx, func(m *metrics.FileMeasures) error {
		return csvWriter.Write(structValues(m))
	})
	if err != nil {
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func writeFeedbackCSV(ctx context.Context, w io.Writer, iter entity.FeedbackIter) error {
	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write(_FeedbackFields); err != nil {
		return err
	}

	err := iter.ForEach(ctx, func(m *entity.Feedback) error {
		return csvWriter.Write(structValues(m))
	})
	if err != nil {
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

type writeCommitsCSVOptions struct {
	FullDumb bool
}

func writeCommitsCSV(ctx context.Context, w io.Writer, iter entity.CommitIter, opts *writeCommitsCSVOptions) error {
	csvWriter := csv.NewWriter(w)

	header := []string{
		"commit_id",
		"author_date",
		"buggy",
	}

	if opts.FullDumb {
		header = append(header, []string{
			"repository_id",
			"commit_hash",
			"author_name",
			"author_email",
			"commit_message",
			"is_fix_commit",
			"fixed_by",
			"files_count",
			"files",
			"bug_potential",
		}...)
	}

	header = append(header, _ChangeMeasuresFields...)

	if err := csvWriter.Write(header); err != nil {
		return err
	}

	err := iter.ForEach(ctx, func(commit *entity.Commit) error {
		if commit.Metrics == nil {
			// ignore not analyzed commits
			return nil
		}
		raw := make([]string, 3, len(header))
		raw[0] = commit.ID.String()
		raw[1] = strconv.FormatInt(commit.Author.When.Unix(), 10)
		raw[2] = fmt.Sprint(len(commit.Fixes) > 0)

		if opts.FullDumb {
			var potential string
			if commit.Analysis != nil {
				potential = fmt.Sprint(commit.Analysis.BugPotential)
			}

			raw = append(raw, []string{
				commit.ID.RepoID.Hex(),
				commit.ID.CommitHash.Hex(),
				commit.Author.Name,
				commit.Author.Email,
				commit.Message,
				fmt.Sprint(commit.Fix),
				stringSliceToString(commit.Fixes, ";"),
				fmt.Sprint(len(commit.Files)),
				fileSliceToString(commit.Files, ";"),
				potential,
			}...)
		}

		raw = append(raw, structValues(commit.Metrics)...)

		return csvWriter.Write(raw)
	})
	if err != nil {
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func structValues(m interface{}) []string {
	val := reflect.ValueOf(m)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	n := val.NumField()
	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = fmt.Sprint(val.Field(i).Interface())
	}

	return r
}

func hashSliceToString(hashes []identifier.Hash, split string) string {
	var str strings.Builder

	for i := range hashes {
		str.WriteString(hashes[i].Hex())
		if i < len(hashes)-1 {
			str.WriteString(split)
		}
	}

	return str.String()
}

func stringSliceToString(ss []string, split string) string {
	var str strings.Builder

	for i := range ss {
		str.WriteString(ss[i])
		if i < len(ss)-1 {
			str.WriteString(split)
		}
	}

	return str.String()
}

func fileSliceToString(files []*entity.File, split string) string {
	var str strings.Builder

	for i := range files {
		str.WriteString(files[i].Path)
		if i < len(files)-1 {
			str.WriteString(split)
		}
	}

	return str.String()
}

func structKeys(m interface{}) []string {
	val := reflect.ValueOf(m)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	n := val.NumField()
	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = strings.ToLower(val.Type().Field(i).Name)
	}

	return r
}
