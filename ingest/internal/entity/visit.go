package entity

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VisitDataSource interface {
	Insert(context.Context, *Visit) error
	VisitStatisticsDataSource
}

type VisitStatisticsDataSource interface {
	ViewsTotalCount(ctx context.Context, since time.Time) (int64, error)
	VisitorsTotalCount(ctx context.Context, since time.Time) (int64, error)
	CountOverTime(ctx context.Context, since time.Time, frequency Frequency) ([]*VisitOverTime, error)
}

//deprecated
type Visit struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Referer   string             `bson:"referer"`
	CreatedAt time.Time          `bson:"created_at"`
}

type VisitOverTime struct {
	Date     string `bson:"_id"`
	Views    int    `bson:"views"`
	Visitors int    `bson:"visitors"`
}
