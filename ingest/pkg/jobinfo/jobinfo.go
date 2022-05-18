package jobinfo

import (
	"errors"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
)

const RepoEntity = "repo_entity"
const PullRequestEntity = "pr_entity"
const PullRequestEntities = "pr_entities"
const PullRequestID = "pr_id"
const PullRequestNumbers = "pr_numbers"
const KeyPushBeforeSHA = "push_before"
const KeyPushAfterSHA = "push_after"

var ErrMissingCommitPushInfo = errors.New("missing commit push info")

type Store = map[string]interface{}

type PushCommits struct {
	Before identifier.Hash
	After  identifier.Hash
}

func GetPushCommits(details Store) (*PushCommits, error) {
	before, ok := details[KeyPushBeforeSHA].(string)
	if !ok {
		return nil, ErrMissingCommitPushInfo
	}

	after, ok := details[KeyPushAfterSHA].(string)
	if !ok {
		return nil, ErrMissingCommitPushInfo
	}

	return &PushCommits{
		Before: identifier.NewHash(before),
		After:  identifier.NewHash(after),
	}, nil
}

type JobInfo struct {
	JobID   identifier.JobID
	RepoID  identifier.RepositoryID
	Action  invoke.Action
	Details Store
	Cache   Store
	IsEqual func(*JobInfo) bool
	next    *JobInfo
}

func (info *JobInfo) CancelNext() {
	info.next = nil
}

func (info *JobInfo) Next() *JobInfo {
	return info.next
}

func (info *JobInfo) HasNext() bool {
	return info.next != nil
}

func (info *JobInfo) ObservableNodeID() string {
	if id, ok := info.Details[PullRequestID].(identifier.PullRequestID); ok {
		return id.NodeID()
	}

	return info.RepoID.NodeID()
}

func DefaultEqualFunc(info1, info2 *JobInfo) bool {
	if info1.Action != info2.Action {
		return false
	}

	if !isSamePullRequest(info1, info2) {
		return false
	}

	return true
}

func isSamePullRequest(info1, info2 *JobInfo) bool {
	id1, ok := info1.Details[PullRequestID].(identifier.PullRequestID)
	if !ok {
		return false
	}

	id2, ok := info2.Details[PullRequestID].(identifier.PullRequestID)
	if !ok {
		return false
	}

	return id1 == id2
}

//deprecated
func (info *JobInfo) IsPullRequest() bool {
	_, ok := info.Details[PullRequestID]
	return ok
}

// IMPOTENT: should be called with the manager mutex locked
func (info *JobInfo) Append(newInfo *JobInfo) bool {
	oldInfo := info

	for oldInfo.next != nil {
		oldInfo = oldInfo.next

		if newInfo.IsEqual != nil && newInfo.IsEqual(oldInfo) {
			//todo: could update the cache
			return false
		}
	}

	oldInfo.next = newInfo
	return true
}
