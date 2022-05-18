package manage

import (
	"context"

	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/pkg/common"
)

//p.Action == invoke.ActionRepositoryAdded || p.Action == invoke.ActionPullRequestRefreshing
func AddPullRequests(ctx context.Context, p *process) error {

	return p.repoEngine.SCM().ListOpenPullRequests(ctx).ForEach(func(pr *common.PullRequest) error {
		pull, err := p.mgr.srv.PullRequest.FindAndUpdateSource(ctx, p.RepoID, pr)
		if err != nil {
			return err
		}

		return p.mgr.ProcessRepository(&jobinfo.JobInfo{
			Action: invoke.ActionPullRequestAdded,
			RepoID: p.RepoID,
			Cache: jobinfo.Store{
				jobinfo.PullRequestEntity: pull,
			},
			Details: jobinfo.Store{
				jobinfo.PullRequestID: pull.ID,
			},
		})
	})
}
