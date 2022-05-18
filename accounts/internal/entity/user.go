package entity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/pkg/common"
	"golang.org/x/oauth2"
)

var (
	ErrUserNotExist = errors.New("user not exist")
)

type UsersDataSource interface {
	Find(ctx context.Context, id permission.UserID) (*User, error)
	ProviderToken(ctx context.Context, id permission.UserID, provider string) (*oauth2.Token, error)
	// deprecated
	// GenerateToken will assign an ID to the user if added successfully
	Insert(ctx context.Context, u *User) error
	// deprecated
	Update(ctx context.Context, u *User) error
	FindAndModifyProvider(ctx context.Context, pu *common.User) (*User, error)
}

type User struct {
	Id        permission.UserID `json:"id"          bson:"_id,omitempty"`
	Username  string            `json:"username"    bson:"username"`
	FirstName string            `json:"first_name"  bson:"first_name"`
	LastName  string            `json:"last_name"   bson:"last_name"`
	AvatarURL string            `json:"avatar_url"  bson:"avatar_url,omitempty"`
	Email     string            `json:"email"       bson:"email"`
	Password  string            `json:"-"           bson:"password,omitempty"`
	Providers []*common.User    `json:"providers"   bson:"providers,omitempty"`
	Role      permission.Role   `json:"role"        bson:"role"`
	CreatedAt time.Time         `json:"-"           bson:"created_at"`
	UpdatedAt time.Time         `json:"-"           bson:"updated_at"`
}

func NewUser(username string, fullName string, avatar string) *User {
	now := time.Now()
	first, last := splitFullName(fullName, username)

	return &User{
		Id:        permission.UserID{},
		Username:  username,
		FirstName: first,
		LastName:  last,
		AvatarURL: avatar,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) UserInfo() *permission.AccessInfo {
	m := make(map[string]string, len(u.Providers))
	for _, pu := range u.Providers {
		m[pu.Provider] = pu.ID
	}

	return &permission.AccessInfo{
		Role: u.Role,
		UserInfo: &permission.UserInfo{
			UserID:    u.Id,
			Providers: m,
		},
	}
}

func splitFullName(fullName, alt string) (string, string) {
	if fullName == "" {
		return alt, ""
	}
	a := strings.SplitN(fullName, " ", 2)

	if len(a) == 2 {
		return a[0], a[1]
	}

	return fullName, ""
}
