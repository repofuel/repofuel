package entity

import (
	"context"
	"errors"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

var (
	ErrVerificationNotExist = errors.New("verification not exist")
)

type VerificationDataSource interface {
	FindByID(ctx context.Context, id string) (*Verification, error)
	Insert(ctx context.Context, v *Verification) error
}

type Verification struct {
	ID        string              `bson:"_id,omitempty"`
	ExpiredAt time.Time           `bson:"expired_at"`
	CreatedAt time.Time           `bson:"created_at"`
	Payload   VerificationPayload `bson:"payload"`
}

type VerificationPayload interface {
	PayloadType() string
}

type LinkingVerificationOauth1 struct {
	OrgID         identifier.OrganizationID `bson:"org_id"`
	RequestSecret string                    `bson:"secret"`
}

func (l *LinkingVerificationOauth1) PayloadType() string {
	return "org_oauth1"
}
