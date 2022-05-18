package model

import (
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/classify"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/insights"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/metrics"
)

type (
	ProgressEvent     = manage.ProgressObservable
	Progress          = manage.Progress
	Tag               = classify.Tag
	Stage             = status.Stage
	Signature         = engine.Signature
	ChangeMeasures    = metrics.ChangeMeasures
	FileMeasures      = metrics.FileMeasures
	RepositorySource  = common.Repository
	Owner             = common.Account
	PullRequestSource = common.PullRequest
	JobInvoker        = invoke.Action
	JobLogEntry       = entity.Update
	Issue             = common.Issue
	Role              = permission.Role
	CommitFile        = entity.File
	DeltaType         = engine.DeltaType
	Insight           = insights.Reason
	UserProviderInfo  = common.User
)

type User struct {
	ID        identifier.UserID `json:"id"        `
	Username  string            `json:"username"  `
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name" `
	AvatarURL string            `json:"avatar_url"`
	Email     string            `json:"email"     `
	Providers []*common.User    `json:"providers" `
	Role      permission.Role   `json:"role"      `
}

func (_ *User) IsRepositoryOwner() {}
func (_ *User) IsNode()            {}

//deprecated
type Activity struct{}
