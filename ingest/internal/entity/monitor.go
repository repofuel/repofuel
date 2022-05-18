package entity

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

type MonitorDataSource interface {
	InsertMonitor(context.Context, *identifier.MonitorID) error
	//fixme: should me random innstead of latest
	LastRepositoryMonitorUserID(context.Context, identifier.RepositoryID) (identifier.UserID, error)
	UserReposIDs(context.Context, identifier.UserID) ([]identifier.RepositoryID, error)
	IsMonitor(context.Context, *identifier.MonitorID) (bool, error)
	MonitorCount(context.Context, identifier.RepositoryID) (int, error)
	RemoveMonitor(context.Context, *identifier.MonitorID) error
}

type Monitor struct {
	ID        *identifier.MonitorID `json:"id,omitempty"               bson:"_id,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitempty"       bson:"created_at,omitempty"`
}
