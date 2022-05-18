package repofuel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/repofuel/repofuel/pkg/common"
)

type IngestService service

//deprecated
type InstallationEvent struct {
	Action string
	Repos  []*common.Repository
}

//deprecated
type AddCollaboratorEvent struct {
	User        common.User        `json:"user"`
	Permissions common.Permissions `json:"permissions"`
}

//deprecated
type PullRequestAction uint

const (
	ActionNewCommit PullRequestAction = iota + 1
	ActionEdited
)

//deprecated
type PullRequestEvent struct {
	Action      PullRequestAction   `json:"action"`
	Provider    string              `json:"provider"`
	SourceID    string              `json:"source_id"`
	PullRequest *common.PullRequest `json:"pull,omitempty"`
}

//deprecated
func (s *IngestService) UpdateInstallation(ctx context.Context, provider string, installation string, events *InstallationEvent) (*http.Response, error) {
	u := fmt.Sprintf("platforms/%s/installations/%s", provider, installation)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodPatch, u, events)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

//deprecated
func (s *IngestService) CreateInstallation(ctx context.Context, provider string, installation string, repos []*common.Repository) (*http.Response, error) {
	u := fmt.Sprintf("platforms/%s/installations/%s", provider, installation)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodPut, u, repos)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

//deprecated
func (s *IngestService) DeleteInstallation(ctx context.Context, provider string, installation string) (*http.Response, error) {
	u := fmt.Sprintf("platforms/%s/installations/%s", provider, installation)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *IngestService) DeleteCollaborator(ctx context.Context, provider string, repoId, userId string) (*http.Response, error) {
	u := fmt.Sprintf("platforms/%s/repositories/%s/collaborators/%s", provider, repoId, userId)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

//deprecated
func (s *IngestService) AddCollaborator(ctx context.Context, provider string, repoId string, event *AddCollaboratorEvent) (*http.Response, error) {
	u := fmt.Sprintf("platforms/%s/repositories/%s/collaborators", provider, repoId)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodPost, u, event)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *IngestService) GetMetricsCSV(ctx context.Context, repoId, startJob, lastJob string) (io.Reader, *http.Response, error) {
	params := url.Values{}
	if startJob != "" {
		params.Set("start_job", startJob)
	}
	if lastJob != "" {
		params.Set("last_job", lastJob)
	}

	u := fmt.Sprintf("repositories/%s/metrics.csv?%s", repoId, params.Encode())

	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	resp, err := s.client.Do(req, &buf)
	return &buf, resp, err
}

//deprecated
func (s *IngestService) FetchNewCommits(ctx context.Context, repo_id string) (*http.Response, error) {
	u := fmt.Sprintf("/repositories/%s/process/trigger", repo_id)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

//deprecated
func (s *IngestService) SendPullRequestEvent(ctx context.Context, e *PullRequestEvent) (*http.Response, error) {
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodPost, "webhooks/pulls", e)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
