package entity

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

type FeedbackDataSource interface {
	Insert(context.Context, *Feedback) error
	FeedbackConnection(*OrderDirection, *PaginationInput) FeedbackConnection
	All(context.Context) (FeedbackIter, error)
}

type Feedback struct {
	ID        identifier.FeedbackID `bson:"_id,omitempty"`
	Sender    identifier.UserID     `bson:"sender"`
	CommitID  identifier.CommitID   `bson:"commit_id"`
	Message   string                `bson:"message"`
	CreatedAt time.Time             `bson:"created_at"`
}
