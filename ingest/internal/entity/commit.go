// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

//go:generate mockgen -destination=$PWD/internal/mock/entity/commit_mock.go . CommitDataSource,RepositoryDataSource,CommitIter

package entity

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/classify"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/insights"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/metrics"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	messageLimit = 70
)

type ChangeMeasures = metrics.ChangeMeasures

type CommitDataSource interface {
	//fixme: should update, should not delete the analysis
	InsertOrReplace(context.Context, *Commit) error
	MarkBuggy(context.Context, identifier.RepositoryID, identifier.Hash, identifier.HashSet) error
	FindRepoCommits(context.Context, identifier.RepositoryID, ...*options.FindOptions) (CommitIter, error)
	FindPullRequestCommits(context.Context, identifier.RepositoryID, identifier.PullRequestID, ...*options.FindOptions) (CommitIter, error)
	FindByID(ctx context.Context, commitID *identifier.CommitID) (*Commit, error)
	FindJobCommits(context.Context, identifier.RepositoryID, identifier.JobID) (CommitIter, error)
	FindCommitsByHash(context.Context, identifier.RepositoryID, ...identifier.Hash) (CommitIter, error)
	FindCommitsBetween(ctx context.Context, repoID identifier.RepositoryID, start identifier.JobID, end identifier.JobID) (CommitIter, error)
	FindCommitsUntil(context.Context, identifier.RepositoryID, identifier.JobID) (CommitIter, error)
	DeleteRepoCommits(context.Context, identifier.RepositoryID) error
	DeleteCommitTag(ctx context.Context, commitID *identifier.CommitID, tag classify.Tag) error
	SaveCommitAnalysis(ctx context.Context, analyses ...*CommitAnalysisHolder) error
	RemoveBranch(ctx context.Context, repoID identifier.RepositoryID, branch string) error
	ReTagBranch(ctx context.Context, repoID identifier.RepositoryID, branch string, commitIDs identifier.HashSet) error
	ReTagPullRequest(ctx context.Context, repoID identifier.RepositoryID, pull identifier.PullRequestID, commitIDs identifier.HashSet) error
	Prune(ctx context.Context, repoID identifier.RepositoryID) error

	RepositoryCommitConnection(ctx context.Context, repoID identifier.RepositoryID, direction *OrderDirection, filters *CommitFilters, opts *PaginationInput) (CommitConnection, error)
	PullRequestCommitConnection(ctx context.Context, repoID identifier.RepositoryID, pullID identifier.PullRequestID, direction *OrderDirection, opts *PaginationInput) (CommitConnection, error)
	SelectedCommitConnection(ctx context.Context, repoID identifier.RepositoryID, ids []identifier.Hash, direction *OrderDirection, opts *PaginationInput) (CommitConnection, error)

	DeveloperEmails(ctx context.Context, repoID identifier.RepositoryID) ([]string, error) // fixme: should be advanced: using connection and all the dev emails
	DeveloperNames(ctx context.Context, repoID identifier.RepositoryID) ([]string, error)  // fixme: should be advanced: using connection and all the dev emails
	DevelopersAggregatedMetrics(ctx context.Context, repoID identifier.RepositoryID) (ChangeMeasuresIter, error)
	RepositoryEngineFiles(context.Context, identifier.RepositoryID) (EngineFileIter, error)

	CommitStatisticsDataSource
}

type CommitStatisticsDataSource interface {
	BugInducingCount(ctx context.Context, repoID identifier.RepositoryID) (int64, error)
	ContributorsCount(ctx context.Context, repoID identifier.RepositoryID) (int, error)
	BugFixingCount(ctx context.Context, repoID identifier.RepositoryID) (int, error)

	BuggyCommitsOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*CountOverTime, error)
	CommitsOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*CountOverTime, error)
	CommitsTagCount(ctx context.Context, repoID identifier.RepositoryID) ([]*TagCount, error)
	AvgEntropyOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*AvgOverTime, error)
	AvgCommitFilesOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*AvgOverTime, error)

	AnalyzedTotalCount(ctx context.Context, since time.Time) (int64, error)
	AnalyzedCountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)
	PredictedTotalCount(ctx context.Context, since time.Time) (int64, error)
	RepositoryPredictionsTotalCount(ctx context.Context, id identifier.RepositoryID) (int64, error)
	PredictedCountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)

	FileAggregatedMetrics(ctx context.Context, id identifier.RepositoryID) (FileMeasuresIter, error)
}

type CommitFilters struct {
	Branch        *string  `json:"branch"`
	DeveloperName *string  `json:"developerName"`
	MinRisk       *float32 `json:"minRisk"`
	MaxRisk       *float32 `json:"maxRisk"`
}

type DeveloperExp struct {
	Email string `bson:"_id"`
	Exp   int    `bson:"exp"`
}

type Developer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

//deprecated
type CommitError struct {
	Stage   string
	Message string
}

type CountOverTime struct {
	Date  string `bson:"_id"`
	Count int    `bson:"count"`
}

type AvgOverTime struct {
	Date string  `bson:"_id"`
	Avg  float64 `bson:"avg"`
}

//deprecated
func NewCommitError(stage string, err error) *CommitError {
	return &CommitError{
		Stage:   stage,
		Message: err.Error(),
	}
}

func (err *CommitError) Error() string {
	return fmt.Sprintf("%s: %s", err.Stage, err.Message)
}

//todo: risk should changed to *float32 to differentiate between zero risk and no predication
type Commit struct {
	ID           *identifier.CommitID       `json:"id"                       bson:"_id,omitempty"`
	Author       *engine.Signature          `json:"author"                   bson:"author"`
	Message      string                     `json:"message"                  bson:"message"`
	Files        []*File                    `json:"files"                    bson:"files,omitempty"`
	Fixes        []string                   `json:"fixes,omitempty"          bson:"fixes,omitempty"`
	Fix          bool                       `json:"fix"                      bson:"fix"`
	Metrics      *metrics.ChangeMeasures    `json:"metrics,omitempty"        bson:"metrics,omitempty"`
	Job          identifier.JobID           `json:"job"                      bson:"job"`
	Analysis     *CommitAnalysis            `json:"analysis"                 bson:"analysis,omitempty"`
	Tags         []classify.Tag             `json:"tags,omitempty"           bson:"tags,omitempty"`
	DeletedTags  []classify.Tag             `json:"deleted_tags,omitempty"   bson:"deleted_tags,omitempty"`
	Merge        bool                       `json:"merge"                    bson:"merge,omitempty"`
	Issues       []common.Issue             `json:"issues,omitempty"         bson:"issues,omitempty"`
	Branches     []string                   `json:"branches,omitempty"       bson:"branches,omitempty"`
	PullRequests []identifier.PullRequestID `json:"pulls,omitempty"          bson:"pulls,omitempty"`
	CreatedAt    time.Time                  `json:"created_at"               bson:"created_at,omitempty"`
}

type CommitAnalysisHolder struct {
	ID           *identifier.CommitID `bson:"-"`
	Analysis     CommitAnalysis       `bson:"analysis,omitempty"`
	FileInsights [][]insights.Reason
}

type CommitAnalysis struct {
	BugPotential float32           `bson:"bug_potential"`
	Indicators   BugIndicators     `bson:"indicators,omitempty"`
	Insights     []insights.Reason `bson:"insights,omitempty"`
}

type BugIndicators struct {
	Experience float32 `bson:"experience"`
	History    float32 `bson:"history"`
	Size       float32 `bson:"size"`
	Diffusion  float32 `bson:"diffusion"`
}

func (c *Commit) IsNode() {}

// Fixed indicate if the commit was fixed be other commits.
func (c *Commit) Fixed() bool {
	return len(c.Fixes) > 0
}

//todo: json tags should be small letter
type Indicators struct {
	Experience float32 `json:"Experience"  bson:"exp"`
	History    float32 `json:"History"     bson:"history"`
	Size       float32 `json:"Size"        bson:"size"`
	Diffusion  float32 `json:"Diffusion"   bson:"dif"`
}

func (c *Commit) HasBranch(branch string) bool {
	for i := range c.Branches {
		if c.Branches[i] == branch {
			return true
		}
	}
	return false
}

func (c *Commit) Hash() identifier.Hash {
	return c.ID.CommitHash
}

func LimitedMessage(s string) string {
	i := strings.IndexByte(s, '\n')

	if i > messageLimit {
		return strings.TrimRight(s[:messageLimit-1], " ") + "…"
	}
	if i >= 0 {
		return s[:i]
	}
	if len(s) > messageLimit {
		return strings.TrimRight(s[:messageLimit-1], " ") + "…"
	}

	return s
}

type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "ASC"
	OrderDirectionDesc OrderDirection = "DESC"
)

var AllOrderDirection = []OrderDirection{
	OrderDirectionAsc,
	OrderDirectionDesc,
}

func (e OrderDirection) IsValid() bool {
	switch e {
	case OrderDirectionAsc, OrderDirectionDesc:
		return true
	}
	return false
}

func (e OrderDirection) String() string {
	return string(e)
}

func (e *OrderDirection) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OrderDirection(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OrderDirection", str)
	}
	return nil
}

func (e OrderDirection) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TagCount struct {
	Tag   classify.Tag `bson:"_id"`
	Count int          `bson:"count"`
}

type FileMeasures = metrics.FileMeasures

type EngineFileIter interface {
	ForEach(ctx context.Context, cb func(identifier.Hash, map[string]*engine.FileInfo) error) error
}

type File struct {
	*engine.FileInfo `bson:",inline"`

	Type          classify.FileType     `json:"type"              bson:"type"`
	Language      string                `json:"language"          bson:"language,omitempty"`
	Fixing        []identifier.Hash     `json:"fixing"            bson:"fixing,omitempty"`
	Metrics       *metrics.FileMeasures `json:"metrics,omitempty" bson:"metrics,omitempty"`
	SameDeveloper bool                  `json:"same_developer"    bson:"same_developer"`
	Insights      []insights.Reason     `json:"insights"          bson:"insights"`
}

type CommitFiles struct {
	CommitHash identifier.Hash
	Files      map[string]*File
}
