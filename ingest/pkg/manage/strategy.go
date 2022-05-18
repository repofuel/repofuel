package manage

import (
	"context"
	"sync"
)

type Task func(context.Context, *process) error

type Pipeline [][]Task

const (
	Clone = iota + 1
	Analyze
	NotifyPlatform
)

var (
	pipelineUpdateRepository        Pipeline
	pipelineUpdatePullRequest       Pipeline
	pipelineNewRepository           Pipeline
	pipelineNewPublicRepository     Pipeline
	pipelineProcessPushCheck        Pipeline
	pipelineProcessPullRequestCheck Pipeline
)

func init() {
	pipelineUpdateRepository = Pipeline{
		{StatusInProgress},
		{PrepareGitRepository, RecoverFromLastPredicting},
		{ingestLocalBranchesTask},
		{analyze},
		{Predict},
		{StatusIsReady},
	}

	pipelineUpdatePullRequest = Pipeline{
		{StatusInProgress},
		{PrepareGitRepository},
		{PreparePullRequest},
		{ingestPullRequestTask},
		{analyze},
		{FinalizePullRequestAnalysis},
		{Predict},
		{StatusIsReady},
	}

	pipelineNewRepository = Pipeline{
		{StatusInProgress},
		{PrepareGitRepository, UpdateCollaborators},
		{ingestLocalBranchesTask, AddPullRequests},
		{analyze},
		{Predict},
		{StatusIsReady},
	}

	pipelineNewPublicRepository = Pipeline{
		{StatusInProgress},
		{PrepareGitRepository},
		{ingestLocalBranchesTask, AddPullRequests},
		{analyze},
		{Predict},
		{StatusIsReady},
	}

	pipelineProcessPushCheck = Pipeline{
		{ProcessCheckRun},
		{ReportPushCheckResult},
	}

	pipelineProcessPullRequestCheck = Pipeline{
		{RepoStatusInProgress},
		{ProcessCheckRun},
		{ReportPullRequestCheckResult},
		{RepoStatusIsReady},
	}
}

func (p *process) runPipeline(ctx context.Context, pip Pipeline) error {
	var wg sync.WaitGroup
	var err error

	p.tracker = p.mgr.observables.GetOrCreate(p.ObservableNodeID())

	for _, stage := range pip {
		numTasks := len(stage)

		if numTasks == 0 {
			panic("stage should have at least one task")
		}

		if numTasks == 1 {
			err = stage[0](ctx, p)
			if err != nil {
				return err
			}
			continue
		}

		wg.Add(numTasks)
		for i := range stage {
			go func(task Task) {
				var localErr error
				defer func() {
					if r := recover(); r != nil {
						localErr = errorFromRecovery(r)
					}
				}()

				localErr = task(ctx, p)
				if localErr != nil && err == nil {
					p.cancel()
					err = localErr
				}
				wg.Done()
			}(stage[i])
		}
		wg.Wait()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
