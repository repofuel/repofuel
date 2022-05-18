package entity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPullRequestNotExist = errors.New("pull request not exist")
)

type PullRequestDataSource interface {
	FindByID(context.Context, identifier.PullRequestID) (*PullRequest, error)
	//deprecated
	FindByRepoID(context.Context, identifier.RepositoryID, ...*options.FindOptions) (PullRequestIter, error)
	//deprecated
	Insert(context.Context, *PullRequest) error
	FindByNumber(context.Context, identifier.RepositoryID, int) (*PullRequest, error)
	//deprecated
	FindAnalyzedHeads(ctx context.Context, repoID identifier.RepositoryID) ([]identifier.Hash, error)
	//deprecated
	UpdateSource(context.Context, identifier.RepositoryID, *common.PullRequest) (bool, error)
	FindAndUpdateSource(context.Context, identifier.RepositoryID, *common.PullRequest) (*PullRequest, error)
	SaveStatus(context.Context, identifier.PullRequestID, status.Stage) error
	SaveAnalyzedHead(ctx context.Context, id identifier.PullRequestID, hash identifier.Hash) error
	StatusByID(ctx context.Context, repoID identifier.PullRequestID) (status.Stage, error)
	RepositoryPullRequestConnection(identifier.RepositoryID, *OrderDirection, *PaginationInput) PullRequestConnection

	PullRequestStatisticDataSource
}

type PullRequestStatisticDataSource interface {
	AnalyzedTotalCount(ctx context.Context, since time.Time) (int64, error)
	AnalyzedCountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)
}

type PullRequestSource = common.PullRequest

type PullRequest struct {
	ID           identifier.PullRequestID `json:"id"                       bson:"_id,omitempty"`
	RepoID       identifier.RepositoryID  `json:"repo_id"                  bson:"repo_id"`
	Source       common.PullRequest       `json:"source"                   bson:"source"`
	AnalyzedHead identifier.Hash          `json:"analyzed,omitempty"       bson:"analyzed,omitempty"`
	Status       status.Stage             `json:"status"                   bson:"status"`
	CreatedAt    time.Time                `json:"created_at,omitempty"     bson:"created_at,omitempty"`
	UpdatedAt    time.Time                `json:"updated_at,omitempty"     bson:"updated_at,omitempty"`
}

func NewPullRequest(repoID identifier.RepositoryID, s *common.PullRequest) *PullRequest {
	return &PullRequest{
		RepoID: repoID,
		Source: *s,
	}
}

func (p *PullRequest) IsSameOrigin() bool {
	return p.Source.Head.CloneURL == p.Source.Base.CloneURL
}

//deprecated
func (c *PullRequest) IsNode() {}

//deprecated
func (c *PullRequest) IsProgressable() {}

func (p *PullRequest) HeadBranchName() string {
	if p.IsSameOrigin() {
		return p.Source.Head.Name
	}

	var s strings.Builder
	s.WriteString(p.ID.Hex())
	s.WriteString("/")
	s.WriteString(p.Source.Head.Name)
	return s.String()
}

//todo: merge with the Branch type in common
type Branch struct {
	Name     string
	SHA      string
	CloneURL string
}
