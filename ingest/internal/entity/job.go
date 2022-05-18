package entity

import (
	"context"
	"errors"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/status"
)

type JobDataSource interface {
	FindByRepo(context.Context, identifier.RepositoryID) (JobIter, error)
	FindLast(context.Context, identifier.RepositoryID) (*Job, error)
	FindLastWithStatus(context.Context, identifier.RepositoryID, ...status.Stage) (*Job, error)
	SaveStatus(context.Context, identifier.JobID, status.Stage) error
	CreateJob(context.Context, identifier.RepositoryID, invoke.Action, map[string]interface{}) (identifier.JobID, error)
	ReportError(context.Context, identifier.JobID, error) error

	RepositoryJobConnection(identifier.RepositoryID, *OrderDirection, *PaginationInput) JobConnection

	JobStatisticDataSource
}

type JobStatisticDataSource interface {
	TotalCount(ctx context.Context, since time.Time) (int64, error)
	CountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*CountOverTime, error)
}

type Job struct {
	ID         identifier.JobID        `json:"id"                 bson:"_id,omitempty"`
	Invoker    invoke.Action           `json:"invoker"            bson:"invoker"` //todo: rename to invoke to event
	Details    map[string]interface{}  `json:"details,omitempty"  bson:"details,omitempty"`
	Repository identifier.RepositoryID `json:"repo_id"            bson:"repo_id"`
	StatusLog  []Update                `json:"log"                bson:"log,omitempty"`
	Error      string                  `json:"error"              bson:"error,omitempty"`
}

//deprecated
func (j *Job) LastWorkingStatus() status.Stage {
	if j == nil {
		return status.Queued
	}

	lastStatus := j.Status()
	if lastStatus == status.Failed || lastStatus == status.Canceled {
		return j.PreviousStatus()
	}

	return lastStatus
}

func (j *Job) IsSameID(j2 *Job) bool {
	if j2 == nil {
		return false
	}

	return j.ID == j2.ID
}

func (j *Job) Status() status.Stage {
	return j.StatusLog[len(j.StatusLog)-1].Status
}

func (j *Job) PreviousStatus() status.Stage {
	switch len(j.StatusLog) {
	case 0, 1:
		return status.Queued
	default:
		return j.StatusLog[len(j.StatusLog)-2].Status
	}
}

func (j *Job) CreatedAt() time.Time {
	return j.ID.Timestamp()
}

type Update struct {
	Status    status.Stage `json:"status"       bson:"status"`
	StartedAt time.Time    `json:"started_at"   bson:"started_at"`
}

func (j *Update) StatusText() string {
	return j.Status.String()
}

var (
	ErrJobNotExist = errors.New("job not exist")
)
