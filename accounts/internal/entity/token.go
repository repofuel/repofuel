// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package entity

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"
const charsetLen = int64(len(charset))

type TokenDataSource interface {
	Find(ctx context.Context, id string) (*Token, error)
	GenerateToken(ctx context.Context, userId permission.UserID) (*Token, error)
}

type Token struct {
	Id        string            `json:"id"          bson:"_id,omitempty"`
	UserId    permission.UserID `json:"user_id"     bson:"user_id"`
	CreatedAt time.Time         `json:"created_at"  bson:"created_at"`
	ExpiredAt time.Time         `json:"expired_at"  bson:"expired_at"`
	//deprecated
	RevokedAt time.Time `json:"revoked_at"  bson:"revoked_at,omitempty"`
}

func (t *Token) String() string {
	return t.Id
}

//deprecated, it should be removed using TTL when it is expired
func (t *Token) IsValid() bool {
	return t != nil && t.RevokedAt.IsZero() && t.ExpiredAt.After(time.Now())
}

func RandomToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
