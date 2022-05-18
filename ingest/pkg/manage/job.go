// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package manage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/repofuel/repofuel/ingest/pkg/insightgen"
	"github.com/repofuel/repofuel/pkg/metrics"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/analysis"
	"github.com/repofuel/repofuel/ingest/pkg/brancher"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/engine/git2go"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type process struct {
	*jobinfo.JobInfo

	repoEngine *engine.Repository
	//deprecated
	includeAfter identifier.JobID // include commits analyzed after the specified job in the prediction
	//deprecated
	startPoints engine.CommitSet

	mu      sync.Mutex
	logger  zerolog.Logger
	tracker *ProgressObservable
	mgr     *Manager
	cancel  context.CancelFunc
}

func RecoverFromLastPredicting(ctx context.Context, p *process) error {
	// clean up if needed
	lastJob, err := p.mgr.srv.Job.FindLast(ctx, p.RepoID)
	if err != nil && err != entity.ErrJobNotExist {
		return err
	}

	lastPredictedJob, err := p.mgr.srv.Job.FindLastWithStatus(ctx, p.RepoID, status.Predicting, status.Ready)
	if err != nil && err != entity.ErrJobNotExist {
		return err
	}

	if err == nil && !lastPredictedJob.IsSameID(lastJob) {
		p.includeAfter = lastPredictedJob.ID

	} else if err == entity.ErrJobNotExist && lastJob != nil {
		// in such case we will include all commits. Instead of getting the id of the first job,
		// we will use the repo id which is always older than the first job id.
		p.includeAfter = identifier.JobID(p.RepoID)
	}

	return nil
}

func (p *process) pullRequestEntities(ctx context.Context) ([]*entity.PullRequest, error) {
	pulls, ok := p.Cache[jobinfo.PullRequestEntities].([]*entity.PullRequest)
	if ok {
		return pulls, nil
	}

	numbers, err := getPullRequestNumbers(p.Details)
	if err != nil {
		return nil, err
	}

	pulls = make([]*entity.PullRequest, len(numbers))

	for i, num := range numbers {
		pull, err := p.mgr.srv.PullRequest.FindByNumber(ctx, p.RepoID, int(num))
		if err != nil {
			return nil, err
		}

		pulls[i] = pull
	}

	p.AddCache(jobinfo.PullRequestEntities, pulls)

	return pulls, nil
}

func (p *process) pullRequestEntity(ctx context.Context) (*entity.PullRequest, error) {
	pull, ok := p.Cache[jobinfo.PullRequestEntity].(*entity.PullRequest)
	if ok {
		return pull, nil
	}

	id, ok := p.JobInfo.Details[jobinfo.PullRequestID].(identifier.PullRequestID)
	if !ok {
		return nil, errors.New("messing a valid pull request ID")
	}

	pull, err := p.mgr.srv.PullRequest.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p.AddCache(jobinfo.PullRequestEntity, pull)

	return pull, nil
}

func (p *process) loadPullRequestEntity(ctx context.Context, repoID identifier.RepositoryID, number int) (*entity.PullRequest, error) {
	//todo: fetch the pull request from the provider if not added to the database
	pull, err := p.mgr.srv.PullRequest.FindByNumber(ctx, repoID, number)
	if err != nil {
		return nil, err
	}

	p.Details[jobinfo.PullRequestID] = pull.ID
	p.AddCache(jobinfo.PullRequestEntity, pull)

	return nil, nil
}

func (p *process) AddCache(key string, value interface{}) {
	if p.Cache != nil {
		p.Cache[key] = value
		return
	}

	p.Cache[key] = jobinfo.Store{
		key: value,
	}
}

func (p *process) RepoEngine(ctx context.Context) (*engine.Repository, error) {
	if p.repoEngine != nil {
		return p.repoEngine, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.repoEngine != nil {
		return p.repoEngine, nil
	}

	repoEntity, ok := p.Cache[jobinfo.RepoEntity].(*entity.Repository)
	if !ok {
		var err error
		repoEntity, err = p.mgr.srv.Repo.FindByID(ctx, p.RepoID)
		if err != nil {
			return nil, err
		}
		p.AddCache(jobinfo.RepoEntity, repoEntity)
	}

	source, issues, err := p.mgr.Integrations.RepositoryIntegrations(ctx, repoEntity)
	if err != nil {
		return nil, err
	}

	repoEngine := engine.NewRepository(p.RepoID, repoEntity.Path(), &engine.RepositoryOpts{
		Adapter:   git2go.NewAdapter(),
		Issues:    issues,
		Source:    source,
		OriginURL: repoEntity.Source.CloneURL,
	})

	p.repoEngine = repoEngine

	return repoEngine, nil
}

func ProcessCheckRun(ctx context.Context, p *process) error {
	repoEngine, err := p.RepoEngine(ctx)
	if err != nil {
		return err
	}

	err = repoEngine.SCM().StartCheckRun(ctx, p.JobID, p.Details)
	if err != nil {
		return err
	}

	err = routeCheckPipelines(ctx, p)
	if err != nil {
		if err := repoEngine.SCM().FinishCheckRun(ctx, p.Details, status.Failed, NewFailureSummary()); err != nil {
			p.logger.Err(err).Msg("fail in reporting check run error")
		}

		return err
	}

	return nil
}

func ReportPushCheckResult(ctx context.Context, p *process) error {
	pushCommits, err := jobinfo.GetPushCommits(p.Details)
	if err != nil {
		return err
	}

	repoEngine, err := p.RepoEngine(ctx)
	if err != nil {
		return err
	}

	var hashes []identifier.Hash

	if pushCommits.Before.IsZero() {
		// The first push in the branch
		// fixme: if the push has multiple commits, only the last one will be included
		hashes = []identifier.Hash{pushCommits.After}
	} else {
		after, ok := repoEngine.Commit(pushCommits.After)
		if !ok {
			return errors.New("after head commit is not ingested")
		}

		err = repoEngine.IngestHead(ctx, pushCommits.Before)
		if err != nil {
			// fixme: it will fail if the commit is not fetched
			return err
		}

		before, ok := repoEngine.Commit(pushCommits.Before)
		if !ok {
			return errors.New("the before commit is not ingested")
		}

		seen := engine.AncestorsList(before)
		hashes = engine.UnseenAncestors(seen, after).HashesSlice()
	}

	commitsItr, err := p.mgr.srv.Commit.FindCommitsByHash(ctx, p.RepoID, hashes...)
	if err != nil {
		return err
	}

	commits, err := commitsItr.Slice(ctx)
	if err != nil {
		return err
	}

	summary := NewPushSummery(commits)

	return repoEngine.SCM().FinishCheckRun(ctx, p.Details, p.tracker.Status(), summary)
}

func ReportPullRequestCheckResult(ctx context.Context, p *process) error {
	pulls, err := p.pullRequestEntities(ctx)
	if err != nil {
		pull, err := p.pullRequestEntity(ctx)
		if err != nil {
			return err
		}
		pulls = []*entity.PullRequest{pull}
	}

	summary := NewPullRequestSummary(len(pulls))
	for _, pull := range pulls {
		commitsItr, err := p.mgr.srv.Commit.FindPullRequestCommits(ctx, p.RepoID, pull.ID)
		if err != nil {
			return err
		}

		commits, err := commitsItr.Slice(ctx)
		if err != nil {
			return err
		}

		summary.AddPullRequest(pull, commits)
	}

	repoEngine, err := p.RepoEngine(ctx)
	if err != nil {
		return err
	}

	return repoEngine.SCM().FinishCheckRun(ctx, p.Details, p.tracker.Status(), summary)
}

func getPullRequestNumbers(details jobinfo.Store) ([]int, error) {
	switch numbers := details[jobinfo.PullRequestNumbers].(type) {
	case []int:
		return numbers, nil

	case bson.A:
		casted := make([]int, len(numbers))
		for i, v := range numbers {
			switch v := v.(type) {
			case int32:
				casted[i] = int(v)
			case int64:
				casted[i] = int(v)
			case int:
				casted[i] = v
			default:
				return nil, errors.New("unsupported pull request numbers")
			}
		}
		return casted, nil

	case []int64:
		casted := make([]int, len(numbers))
		for i := range numbers {
			casted[i] = int(numbers[i])
		}
		return casted, nil

	case []int32:
		casted := make([]int, len(numbers))
		for i := range numbers {
			casted[i] = int(numbers[i])
		}
		return casted, nil

	case nil:
		return nil, errors.New("missing pull request numbers")

	default:
		return nil, errors.New("unsupported pull request numbers")
	}
}

func routeCheckPipelines(ctx context.Context, p *process) error {
	switch p.Action {
	default:
		return errors.New("unsupported action for checks")

	case invoke.ActionPushCheck:
		return p.runPipeline(ctx, pipelineUpdateRepository)

	case invoke.ActionPullRequestCheck:
	}

	if _, ok := p.Details[jobinfo.PullRequestID]; ok {
		return p.runPipeline(ctx, pipelineUpdatePullRequest)
	}

	pulls, err := p.pullRequestEntities(ctx)
	if err != nil {
		return err
	}

	for _, pull := range pulls {
		p.Details[jobinfo.PullRequestID] = pull.ID
		p.AddCache(jobinfo.PullRequestEntity, pull)

		err = p.runPipeline(ctx, pipelineUpdatePullRequest)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateCollaborators(ctx context.Context, p *process) error {
	repoEngine, err := p.RepoEngine(ctx)
	if err != nil {
		return err
	}

	coll, err := repoEngine.SCM().FetchCollaborators(ctx)
	if err != nil {
		return err
	}

	return p.mgr.srv.Repo.UpdateCollaborators(ctx, p.RepoID, coll)
}

func PrepareGitRepository(ctx context.Context, p *process) error {
	repoEngine, err := p.RepoEngine(ctx)
	if err != nil {
		return err
	}

	err = repoEngine.Open()
	if err != nil {
		if err == engine.ErrLocalRepoNotExist {
			err = p.saveStatus(ctx, status.Cloning)
			if err != nil {
				return err
			}

			return repoEngine.Clone(ctx)
		}
		return err
	}

	err = p.saveStatus(ctx, status.Fetching)
	if err != nil {
		return err
	}

	return repoEngine.FetchOrigin(ctx)
}

func PreparePullRequest(ctx context.Context, p *process) error {
	pull, err := p.pullRequestEntity(ctx)
	if err != nil {
		return err
	}

	if pull.IsSameOrigin() {
		return nil
	}

	head := pull.Source.Head
	//todo: PR: should include the author id, and that can be used for the remote name instead of PR ID
	return p.repoEngine.Fetch(ctx, pull.ID.Hex(), head.CloneURL, head.Name)
}

func (p *process) saveStatus(ctx context.Context, s status.Stage) error {
	p.logger.Info().Str("status", s.String()).Msg("save status")

	defer p.tracker.SetNewStage(s, true)

	if s != status.Failed {
		err := p.mgr.srv.Job.SaveStatus(ctx, p.JobID, s)
		if err != nil {
			return err
		}
	}

	if p.IsPullRequest() {
		pull, err := p.pullRequestEntity(ctx)
		if err != nil {
			return err
		}
		return p.mgr.srv.PullRequest.SaveStatus(ctx, pull.ID, s)
	}

	return p.mgr.srv.Repo.SaveStatus(ctx, p.RepoID, s)
}

func (p *process) Stop() {
	p.cancel()
}

func (p *process) run(baseCtx context.Context) {
	var err error
	var ctx context.Context

	p.logger = log.Ctx(baseCtx).With().
		Hex("repo", p.RepoID[:]).
		Hex("job", p.JobID[:]).
		Logger()
	baseCtx = p.logger.WithContext(baseCtx)
	ctx, p.cancel = context.WithCancel(baseCtx)

	p.logger.Info().
		Str("action", p.Action.String()).
		Msg("start processing a job")

	defer func() {
		if r := recover(); r != nil {
			err = errorFromRecovery(r)
		}

		err = p.checkErrors(baseCtx, err)
		if err != nil {
			p.logger.Err(err).Msg("error while reporting an error")
		}

		select {
		case <-ctx.Done():
			p.CancelNext()
		default:
			p.cancel()
		}

		p.mgr.done <- p
	}()

	var pipe Pipeline
	switch p.Action {
	case invoke.ActionRepositoryAdded:
		pipe = pipelineNewRepository

	case invoke.ActionMonitorRepository:
		pipe = pipelineNewPublicRepository

	case invoke.ActionPullRequestAdded, invoke.ActionPullRequestUpdate:
		pipe = pipelineUpdatePullRequest

	case invoke.ActionPushCheck:
		pipe = pipelineProcessPushCheck

	case invoke.ActionPullRequestCheck:
		pipe = pipelineProcessPullRequestCheck

	default:
		pipe = pipelineUpdateRepository
	}

	err = p.runPipeline(ctx, pipe)
}

func errorFromRecovery(r interface{}) error {
	switch err := r.(type) {
	case error:
		return err
	default:
		return fmt.Errorf("recover process: %v", err)
	}
}

func (p *process) checkErrors(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	if err == context.Canceled {
		return p.saveStatus(ctx, status.Canceled)
	}

	p.logger.Error().Err(err).Msg("processing error")

	err = p.mgr.srv.Job.ReportError(ctx, p.JobID, err)
	if err != nil {
		return err
	}

	return p.saveStatus(ctx, status.Failed)
}

func ingestLocalBranches(ctx context.Context, reposDB entity.RepositoryDataSource, commitsDB entity.CommitDataSource, repo *engine.Repository) (engine.CommitSet, error) {
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	// ingest all branches, this will not include pull requests from external repositories
	for _, head := range branches {
		err := repo.IngestHead(ctx, head)
		if err != nil {
			return nil, err
		}
	}

	// cache the commits count in the repository entity
	err = reposDB.SaveCommitsCount(ctx, repo.ID, repo.CommitsCount())
	if err != nil {
		return nil, err
	}

	analyzed, err := reposDB.Branches(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	base := repo.HeadCommits(analyzed).Slice()
	heads := repo.HeadCommits(branches)
	seen := engine.AncestorsList(base...)
	var start engine.CommitSet
	if len(analyzed) == 0 {
		start = repo.Roots()
	} else {
		start, err = engine.UnseenRoots(ctx, seen, heads.Slice()...)
		if err != nil {
			return nil, err
		}
	}

	err = repo.TagBranchesOnCommits(branches, start)
	if err != nil {
		return nil, err
	}

	rearrange := brancher.NewRearrange(commitsDB, reposDB, repo, branches, analyzed, start.Slice())
	err = engine.RunForwardAnalysis(ctx, rearrange, start.Slice())
	if err != nil {
		return nil, err
	}

	if len(start) == 0 {
		// in this case no need to load the files, we will not analyze
		return start, nil
	}

	itr, err := commitsDB.RepositoryEngineFiles(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	numSeen := seen.Count()

	err = itr.ForEach(ctx, func(hash identifier.Hash, files map[string]*engine.FileInfo) error {
		if c, ok := repo.Commit(hash); ok {
			c.SetFiles(files)
			seen.Add(c)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if seen.Count() > numSeen {
		return engine.UnseenRoots(ctx, seen, heads.Slice()...)
	}

	return start, nil
}

var pullFilesOpts = options.Find().SetProjection(bson.M{"files": 1})

func ingestPullRequest(ctx context.Context, reposDB entity.RepositoryDataSource, commitsDB entity.CommitDataSource, repo *engine.Repository, pull *entity.PullRequest) (engine.CommitSet, error) {
	head := identifier.NewHash(pull.Source.Head.SHA)
	err := repo.IngestHead(ctx, head)
	if err != nil {
		return nil, err
	}

	err = repo.IngestHead(ctx, identifier.NewHash(pull.Source.Base.SHA))
	if err != nil {
		return nil, err
	}

	analyzed, err := reposDB.Branches(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	// seen
	bc := repo.HeadCommits(analyzed)
	bc.Update(repo.Commits(pull.AnalyzedHead))
	seen := engine.AncestorsList(bc.Slice()...)

	heads := repo.Commits(head)
	start, err := engine.UnseenRoots(ctx, seen, heads.Slice()...)
	if err != nil {
		return nil, err
	}

	if len(start) == 0 {
		// in this case no need to load the files, we will not analyze
		return start, nil
	}

	itr, err := commitsDB.RepositoryEngineFiles(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	numSeen := seen.Count()

	err = itr.ForEach(ctx, func(hash identifier.Hash, files map[string]*engine.FileInfo) error {
		if c, ok := repo.Commit(hash); ok {
			c.SetFiles(files)
			seen.Add(c)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if seen.Count() > numSeen {
		return engine.UnseenRoots(ctx, seen, heads.Slice()...)
	}

	return start, nil
}

func ingestLocalBranchesTask(ctx context.Context, p *process) error {
	err := p.saveStatus(ctx, status.Ingesting)
	if err != nil {
		return err
	}

	p.startPoints, err = ingestLocalBranches(ctx, p.mgr.srv.Repo, p.mgr.srv.Commit, p.repoEngine)
	return err
}

func ingestPullRequestTask(ctx context.Context, p *process) error {
	err := p.saveStatus(ctx, status.Ingesting)
	if err != nil {
		return err
	}

	pull, err := p.pullRequestEntity(ctx)
	if err != nil {
		return err
	}

	p.startPoints, err = ingestPullRequest(ctx, p.mgr.srv.Repo, p.mgr.srv.Commit, p.repoEngine, pull)
	if err != nil {
		return err
	}

	if len(p.startPoints) == 0 {
		//todo: make more chicks to make sure that the pull request needs this if needed
		return MarkPullRequestCommits(ctx, p.mgr.srv.Commit, pull, p.repoEngine)
	}

	return nil
}

func analyze(ctx context.Context, p *process) error {
	if len(p.startPoints) == 0 {
		p.logger.Info().Msg("nothing new to be analyzed")
		return nil
	}

	err := p.saveStatus(ctx, status.Analyzing)
	if err != nil {
		return err
	}

	p.logger.Info().Int("start_points_count", len(p.startPoints)).Msg("start analyzing")

	p.tracker.SetStageTotal(p.repoEngine.CommitsCount())
	analyzer := analysis.NewRepositoryAnalysis(p.repoEngine, p.JobID, p.mgr.srv.Commit, p.tracker)

	err = analyzer.Run(ctx, p.startPoints)
	if err != nil {
		return err
	}

	b, err := p.repoEngine.Branches()
	if err != nil {
		return err
	}
	err = p.mgr.srv.Repo.SaveBranches(ctx, p.RepoID, b)
	if err != nil {
		return err
	}

	count, err := p.mgr.srv.Commit.BugInducingCount(ctx, p.RepoID)
	if err != nil {
		return err
	}

	return p.mgr.srv.Repo.SaveBuggyCount(ctx, p.RepoID, int(count))
}

func FinalizePullRequestAnalysis(ctx context.Context, p *process) error {
	if len(p.startPoints) == 0 {
		return nil
	}

	pull, err := p.pullRequestEntity(ctx)
	if err != nil {
		return err
	}

	err = MarkPullRequestCommits(ctx, p.mgr.srv.Commit, pull, p.repoEngine)
	if err != nil {
		return err
	}

	// todo: could be improved
	if !pull.IsSameOrigin() {
		err = p.mgr.srv.PullRequest.SaveAnalyzedHead(ctx, pull.ID, identifier.NewHash(pull.Source.Head.SHA))
		if err != nil {
			return err
		}
	}

	return nil
}

func MarkPullRequestCommits(ctx context.Context, commitsDB entity.CommitDataSource, pull *entity.PullRequest, repo *engine.Repository) error {
	bases := repo.Commits(identifier.NewHash(pull.Source.Base.SHA))
	heads := repo.Commits(identifier.NewHash(pull.Source.Head.SHA))

	seen := engine.AncestorsList(bases.Slice()...)
	commitIDs := engine.UnseenAncestors(seen, heads.Slice()...).HashesSet()

	return commitsDB.ReTagPullRequest(ctx, repo.ID, pull.ID, commitIDs)
}

func Predict(ctx context.Context, p *process) error {
	if len(p.startPoints) == 0 && p.includeAfter.IsZero() {
		p.logger.Info().Msg("skip predicting")
		return nil
	}

	err := p.saveStatus(ctx, status.Predicting)
	if err != nil {
		return err
	}

	var oldestJob string
	if !p.includeAfter.IsZero() {
		oldestJob = p.includeAfter.Hex()
	}

	res, _, err := p.mgr.srv.Repofuel.AI.PredictByJob(ctx, p.RepoID.Hex(), p.JobID.Hex(), oldestJob)
	if err != nil {
		return err
	}

	ba, err := generateCommitAnalyses(ctx, p.mgr.srv.Commit, res.Predictions, res.Quantiles)
	if err != nil {
		return err
	}

	err = p.mgr.srv.Commit.SaveCommitAnalysis(ctx, ba...)
	if err != nil {
		return err
	}

	if res.Status == repofuel.PredictOk {
		return p.mgr.srv.Repo.SaveConfidence(ctx, p.RepoID, res.Confidence)
	}

	return p.mgr.srv.Repo.SaveQuality(ctx, p.RepoID, res.Status)
}

func generateCommitAnalyses(ctx context.Context, commitDB entity.CommitDataSource, predictions []repofuel.Prediction, quantiles *metrics.Quantiles) ([]*entity.CommitAnalysisHolder, error) {
	if len(predictions) == 0 {
		return nil, nil
	}

	gen, err := insightgen.NewGenerator(quantiles)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.CommitAnalysisHolder, len(predictions))

	for i, p := range predictions {
		id, err := identifier.CommitIDFromStr(p.CommitID, '_')
		if err != nil {
			return nil, err
		}

		c, err := commitDB.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		res[i] = &entity.CommitAnalysisHolder{
			ID: id,
			Analysis: entity.CommitAnalysis{
				BugPotential: p.Score,
				Indicators: entity.BugIndicators{
					Experience: p.Experience,
					History:    p.History,
					Size:       p.Size,
					Diffusion:  p.Diffusion,
				},
				Insights: gen.CommitInsights(c),
			},
			FileInsights: gen.FileInsights(c),
		}
	}

	return res, nil
}

func (p *process) Progress() *Progress {
	return p.tracker.Progress()
}

func StatusInProgress(ctx context.Context, p *process) error {
	return p.saveStatus(ctx, status.Progressing)
}

func StatusIsReady(ctx context.Context, p *process) error {
	return p.saveStatus(ctx, status.Ready)
}

func RepoStatusInProgress(ctx context.Context, p *process) error {
	return repoStatus(ctx, p, status.Progressing)
}

func RepoStatusIsReady(ctx context.Context, p *process) error {
	return repoStatus(ctx, p, status.Ready)
}

func repoStatus(ctx context.Context, p *process, s status.Stage) error {
	tracker := p.mgr.observables.GetOrCreate(p.RepoID.Hex())
	defer tracker.SetNewStage(s, true)

	return p.mgr.srv.Repo.SaveStatus(ctx, p.RepoID, s)
}
