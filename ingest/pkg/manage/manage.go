// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package manage

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/jobinfo"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	run          bool
	ctx          context.Context
	logger       *zerolog.Logger
	numWorkers   int
	srv          ManagerServices
	done         chan *process
	Integrations *IntegrationManager
	queues       map[QueueID]*Queue
	observables  *ProgressObservableRegistry
	processes    map[identifier.RepositoryID]*process
	queued       map[identifier.RepositoryID]*jobinfo.JobInfo
	mu           sync.Mutex
}

// deprecated: should be moved to entity
type ManagerServices struct {
	Provider     entity.ProviderDataSource
	Commit       entity.CommitDataSource
	Repo         entity.RepositoryDataSource
	Job          entity.JobDataSource
	PullRequest  entity.PullRequestDataSource
	Organization entity.OrganizationDataSource
	Verification entity.VerificationDataSource
	Monitor      entity.MonitorDataSource
	//deprecated: should be moved
	Repofuel *repofuel.Client
}

func NewManager(ctx context.Context, services ManagerServices, authCheck *jwtauth.AuthCheck) *Manager {
	mgr := &Manager{
		ctx:          ctx,
		logger:       log.Ctx(ctx),
		numWorkers:   DefaultManagerWorkersCount,
		srv:          services,
		done:         make(chan *process),
		Integrations: nil,
		queues:       make(map[QueueID]*Queue),
		observables:  nil,
		processes:    make(map[identifier.RepositoryID]*process),
		queued:       make(map[identifier.RepositoryID]*jobinfo.JobInfo),
		mu:           sync.Mutex{},
	}
	mgr.Integrations = NewIntegrationManager(mgr, services.Provider, services.Organization, services.Verification, authCheck, services.Repofuel)
	mgr.observables = newProgressObservableRegistry(mgr)

	return mgr
}

// IMPOTENT: should be called with the manager mutex locked
func (mgr *Manager) getOrCreateQueue(id QueueID) *Queue {
	q, ok := mgr.queues[id]
	if !ok {
		q = &Queue{NumWorkers: DefaultQueueWorkersCount}
		mgr.queues[id] = q
	}
	return q
}

func (mgr *Manager) ProgressObservableRegistry() *ProgressObservableRegistry {
	return mgr.observables
}

func (mgr *Manager) Run() {
	mgr.mu.Lock()

	mgr.run = true
	mgr.processQueuedRepositories()

	mgr.mu.Unlock()

	for p := range mgr.done {
		mgr.runNext(p)
	}
}

func (mgr *Manager) GetJobQueue(info *jobinfo.JobInfo) (*Queue, bool) {
	q, ok := mgr.queues[FindQueueID(info)]
	return q, ok
}

func FindQueueID(info *jobinfo.JobInfo) QueueID {
	switch info.Action {
	case invoke.ActionRepositoryRecovering:
		return QueueRecovered

	case invoke.ActionRepositoryAdded, invoke.ActionMonitorRepository:
		return QueueNewRepos

	default:
		return QueueNewCommits

	}
}

func (mgr *Manager) runNext(p *process) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	defer mgr.processQueuedRepositories()

	queue, ok := mgr.queues[FindQueueID(p.JobInfo)]
	if !ok {
		mgr.logger.Error().Msg("missed queue entry")
		return
	}
	queue.processDone()

	if !p.HasNext() {
		mgr.deleteProcess(p.RepoID)
		mgr.processJobsFromQueue(queue)
		return
	}

	p.JobInfo = p.Next()
	queue = mgr.getOrCreateQueue(FindQueueID(p.JobInfo))
	go p.run(mgr.ctx)
	queue.processStarted()
}

var notWorkingStages = []status.Stage{
	status.Ready,
	status.Failed,
	status.Canceled,
	status.Watched,
}

func (mgr *Manager) Recover(ctx context.Context) error {
	// fixme: the recover process does not scale to multiple nodes
	itr, err := mgr.srv.Repo.FindWhereStatusNot(ctx, notWorkingStages...)
	if err != nil {
		return err
	}

	return itr.ForEach(ctx, func(repo *entity.Repository) error {
		return mgr.ProcessRepository(&jobinfo.JobInfo{
			Action: invoke.ActionRepositoryRecovering,
			RepoID: repo.ID,
			Cache: jobinfo.Store{
				jobinfo.RepoEntity: repo,
			},
		})
	})
}

// IMPOTENT: should be called with the manager mutex locked
func (mgr *Manager) processQueuedRepositories() {
	if !mgr.HasFreeWorkers() {
		return
	}

	seen := NewQueueSet()

	for _, info := range mgr.queued {
		queue := mgr.getOrCreateQueue(FindQueueID(info))
		if seen.Has(queue) {
			continue
		}

		mgr.processJobsFromQueue(queue)
		seen.Add(queue)
	}
}

// IMPOTENT: should be called with the manager mutex locked
func (mgr *Manager) processJobsFromQueue(queue *Queue) {
	for queue.HasFreeWorkers() && queue.NumWaiting() > 0 && mgr.HasFreeWorkers() {
		repoId := queue.pup()
		process, err := mgr.addProcess(repoId)
		if err != nil {
			mgr.logger.Err(err).
				Hex("repo", repoId[:]).
				Msg("cannot add the process")
			continue
		}
		go process.run(mgr.ctx)
		queue.processStarted()
	}
}

func (mgr *Manager) HasFreeWorkers() bool {
	return mgr.numWorkers > mgr.NumProcessing()
}

func (mgr *Manager) NumProcessing() int {
	count := 0

	for _, q := range mgr.queues {
		count += q.NumProcessing()
	}

	return count
}

// IMPOTENT: should be called with the manager mutex locked
func (mgr *Manager) deleteProcess(id identifier.RepositoryID) {
	p, ok := mgr.processes[id]
	if !ok {
		return
	}

	mgr.observables.RemoveEmpty(p.ObservableNodeID())
	delete(mgr.processes, id)
}

// IMPOTENT: should be called with the manager mutex locked
func (mgr *Manager) addProcess(repoID identifier.RepositoryID) (*process, error) {
	info, ok := mgr.queued[repoID]
	if !ok {
		return nil, errors.New("the repository should be queued")
	}
	delete(mgr.queued, repoID)

	runningProcess, ok := mgr.processes[repoID]
	if ok {
		runningProcess.Append(info)
		return nil, errors.New("if the repository has a process, the job should be appended and not queued")
	}

	p := &process{
		JobInfo: info,
		mgr:     mgr,
	}
	mgr.processes[repoID] = p

	return p, nil
}

func (mgr *Manager) ProcessRepository(info *jobinfo.JobInfo) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if info.JobID.IsZero() {
		jobID, err := mgr.srv.Job.CreateJob(mgr.ctx, info.RepoID, info.Action, info.Details)
		if err != nil {
			return err
		}
		info.JobID = jobID
	}

	if job, ok := mgr.processes[info.RepoID]; ok {
		isAppended := job.Append(info)
		if !isAppended {
			return mgr.srv.Job.SaveStatus(mgr.ctx, info.JobID, status.Ignored)
		}
		return nil
	}

	if job, ok := mgr.queued[info.RepoID]; ok {
		isAppended := job.Append(info)
		if !isAppended {
			return mgr.srv.Job.SaveStatus(mgr.ctx, info.JobID, status.Ignored)
		}
		return nil
	}

	// queue the job
	queue := mgr.getOrCreateQueue(FindQueueID(info))
	queue.push(info.RepoID)
	mgr.queued[info.RepoID] = info

	if mgr.run {
		mgr.processJobsFromQueue(queue)
	}

	return nil
}

func (mgr *Manager) StopRepository(id identifier.RepositoryID) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if _, ok := mgr.queued[id]; ok {
		delete(mgr.queued, id)

		// remove the repository from all queues
		for _, q := range mgr.queues {
			q.removeRepository(id)
		}
	}

	if p, ok := mgr.processes[id]; ok {
		delete(mgr.processes, id)
		p.Stop()
	}
}

func (mgr *Manager) DeleteRepository(ctx context.Context, repo *entity.Repository) error {
	mgr.StopRepository(repo.ID)

	// delete the commits
	err := mgr.srv.Commit.DeleteRepoCommits(ctx, repo.ID)
	if err != nil {
		return err
	}

	// delete the git assets
	err = os.RemoveAll(repo.Path())
	if err != nil {
		return err
	}
	//todo: delete jobs, models, pulls

	return mgr.srv.Repo.Delete(ctx, repo.ID)
}

func (mgr *Manager) AddRepository(ctx context.Context, repo *entity.Repository) error {
	err := mgr.srv.Repo.InsertOrUpdate(ctx, repo)
	if err != nil {
		return err
	}

	return mgr.ProcessRepository(&jobinfo.JobInfo{
		Action: invoke.ActionRepositoryAdded,
		RepoID: repo.ID,
		Cache: jobinfo.Store{
			jobinfo.RepoEntity: repo,
		},
	})
}

//deprecated: this mix between jobs for pull requests and repository
func (mgr *Manager) Progress(id identifier.RepositoryID) *Progress {
	p, ok := mgr.processes[id]
	if !ok {
		return nil
	}
	return p.Progress()
}

//fixme: unified this const variables, we have others in the resolver an engine (RepositoryID)
const (
	NodeTypeCommit      = "Commit"
	NodeTypeRepository  = "Repository"
	NodeTypePullRequest = "PullRequest"
)

//todo: need to be improved
func (mgr *Manager) Node(ctx context.Context, id string) (interface{ IsNode() }, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}

	i := bytes.IndexByte(b, ':')
	nodeType, idBytes := string(b[:i]), b[i+1:]
	switch nodeType {
	case NodeTypeCommit:
		return mgr.srv.Commit.FindByID(ctx, identifier.CommitIDFromBytes(idBytes))
	case NodeTypeRepository:
		return mgr.srv.Repo.FindByID(ctx, identifier.RepositoryIDFromBytes(idBytes))
	case NodeTypePullRequest:
		return mgr.srv.PullRequest.FindByID(ctx, identifier.PullRequestIDFromBytes(idBytes))
	default:

	}
	panic(fmt.Errorf("not implemented"))
}

//todo: need to be improved
func (mgr *Manager) NodeStatus(ctx context.Context, id string) (status.Stage, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return 0, err
	}

	i := bytes.IndexByte(b, ':')
	nodeType, idBytes := string(b[:i]), b[i+1:]
	switch nodeType {
	case NodeTypeRepository:
		return mgr.srv.Repo.StatusByID(ctx, identifier.RepositoryIDFromBytes(idBytes))
	case NodeTypePullRequest:
		return mgr.srv.PullRequest.StatusByID(ctx, identifier.PullRequestIDFromBytes(idBytes))
	default:

	}
	panic(fmt.Errorf("not implemented"))
}
