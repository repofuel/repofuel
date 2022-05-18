package providers

import (
	"context"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

type CheckRunSummarizer interface {
	Title() string
	Summary() string
	DetailsText(providerName, providerUrl, ownerName, repoName string) string
}

type Integration interface{}

type SourceIntegration interface {
	FetchRepositoryInfo(ctx context.Context) (*common.Repository, error)
	FetchCollaborators(ctx context.Context) (map[string]common.Permissions, error)
	FetchRepositoryCollaborators(ctx context.Context) (map[string]common.Permissions, error)
	BasicAuth(context.Context) (*credentials.BasicAuth, error)
	ListOpenPullRequests(ctx context.Context) common.PullRequestItr
	StartCheckRun(ctx context.Context, jobID identifier.JobID, details jobinfo.Store) error
	FinishCheckRun(ctx context.Context, details jobinfo.Store, s status.Stage, summarizer CheckRunSummarizer) error
}

type IssuesIntegration interface {
	IssuesFromText(ctx context.Context, s string) ([]common.Issue, bool, error)
}
