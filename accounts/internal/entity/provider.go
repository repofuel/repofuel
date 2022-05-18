package entity

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

type AuthProviderDataSource interface {
	FindByID(context.Context, string) (*AuthProvider, error)
	InsertOrUpdate(context.Context, *AuthProvider) error
}

type AuthProvider struct {
	ID        string              `json:"id"                   bson:"_id"`
	System    common.System       `json:"system"               bson:"system"`
	Server    string              `json:"server"               bson:"server"`
	OAuth1    *credentials.OAuth1 `json:"oauth1,omitempty"     bson:"oauth1,omitempty"`
	OAuth2    *credentials.OAuth2 `json:"oauth2,omitempty"     bson:"oauth2,omitempty"`
	CreatedAt time.Time           `json:"created_at,omitempty"    bson:"created_at,omitempty"`
	UpdatedAt time.Time           `json:"updated_at,omitempty"    bson:"updated_at,omitempty"`
}
