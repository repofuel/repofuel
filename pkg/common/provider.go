// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

// todo: rename and move to the ingest repository (common, abstract, universal, or generic)
package common

// todo: change the pkg name
import (
	"context"
	"encoding/json"

	"github.com/repofuel/repofuel/pkg/credentials"

	"net/http"
	"strconv"
	"time"
)

type AccountType uint8

const (
	AccountPersonal AccountType = iota + 1
	AccountOrganization
)

type System uint8

func (s System) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s System) String() string {
	switch s {
	case SystemGithub:
		return "Github"
	case SystemGithubEnterprise:
		return "Github Enterprise"
	case SystemBitbucketCloud:
		return "Bitbucket Cloud"
	case SystemBitbucketServer:
		return "Bitbucket Server"
	case SystemJiraCloud:
		return "Jira Cloud"
	case SystemJiraServer:
		return "Jira Server"
	}

	return strconv.FormatInt(int64(s), 10)
}

const (
	SystemGithub System = iota + 1
	SystemGithubEnterprise
	SystemBitbucketCloud
	SystemBitbucketServer
	SystemJiraCloud
	SystemJiraServer
)

//deprecated
const (
	Github    Name = "github"
	Bitbucket      = "bitbucket"
	Jira           = "jira"
)

//deprecated
type Name string

//deprecated
type Installation string

//	integration

// deprecated
// BasicAuth represent a HTTP basic auth
type BasicAuth struct {
	Username, Password string
}

func (i Installation) Int64() (int64, error) {
	return strconv.ParseInt(string(i), 10, 64)
}

//deprecated
type IssueService struct {
	Provider Name `json:"provider,omitempty"     bson:"provider,omitempty"`
}

// FetchAuthUserFunc fetch the authenticated user.
type FetchAuthUserFunc func(context.Context, *http.Client) (*User, error)

type User struct {
	//deprecated
	Provider  string                `json:"provider,omitempty"     bson:"provider,omitempty"`
	ID        string                `json:"id,omitempty"           bson:"id,omitempty"`
	Username  string                `json:"username,omitempty"     bson:"username,omitempty"`
	FullName  string                `json:"name,omitempty"         bson:"name,omitempty"`
	AvatarURL string                `json:"avatar_url,omitempty"   bson:"avatar_url,omitempty"`
	Location  string                `json:"location,omitempty"     bson:"location,omitempty"`
	HomePage  string                `json:"home_page,omitempty"    bson:"home_page,omitempty"`
	Cred      credentials.Interface `json:"-"                      bson:"cred,omitempty"`
}

type Permissions struct {
	Admin bool
	Read  bool
	Write bool
}

type OrgRole uint8

const (
	OrgAdmin OrgRole = iota + 1
	OrgMember
)

type Membership struct {
	Role OrgRole
}

var FullPermissions = Permissions{
	Admin: true,
	Read:  true,
	Write: true,
}

type Repository struct {
	ID            string    `json:"id"                        bson:"id,omitempty"`
	RepoName      string    `json:"name,omitempty"            bson:"name,omitempty"`
	Description   string    `json:"description,omitempty"     bson:"description,omitempty"`
	DefaultBranch string    `json:"default_branch,omitempty"  bson:"default_branch,omitempty"`
	HTMLURL       string    `json:"html_url,omitempty"        bson:"html_url,omitempty"`
	CloneURL      string    `json:"clone_url,omitempty"       bson:"clone_url,omitempty"`
	SSHURL        string    `json:"ssh_url,omitempty"         bson:"ssh_url,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"      bson:"created_at,omitempty"`
	Private       bool      `json:"private,omitempty"         bson:"private,omitempty"`
}

var emptyRepository = Repository{}

func (r Repository) IsZero() bool {
	return r == emptyRepository
}

type Account struct {
	ID   string      `json:"id,omitempty"         bson:"id,omitempty"`
	Slug string      `json:"slug,omitempty"       bson:"slug,omitempty"`
	Type AccountType `json:"type,omitempty"       bson:"type,omitempty"`
}

var emptyAccount = Account{}

func (r Account) IsZero() bool {
	return r == emptyAccount
}

type Issue struct {
	Id        string    `json:"id"                      bson:"id"`
	Fetched   bool      `json:"fetched"                 bson:"fetched,omitempty"`
	Bug       bool      `json:"bug"                     bson:"bug,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"    bson:"created_at,omitempty"`
}

type PullRequest struct {
	ID        string    `json:"id"                    bson:"id"`
	Number    int       `json:"number"                bson:"number"`
	Title     string    `json:"title"                 bson:"title"`
	Body      string    `json:"body"                  bson:"body"`
	ClosedAt  time.Time `json:"closed_at,omitempty"   bson:"closed_at,omitempty"`
	MergedAt  time.Time `json:"merged_at,omitempty"   bson:"merged_at,omitempty"`
	Head      *Branch   `json:"head,omitempty"        bson:"head,omitempty"`
	Base      *Branch   `json:"base,omitempty"        bson:"base,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"  bson:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"  bson:"updated_at,omitempty"`
}

func (c *PullRequest) Closed() bool {
	return !c.ClosedAt.IsZero()
}

func (c *PullRequest) Merged() bool {
	return !c.MergedAt.IsZero()
}

type PullRequestItr interface {
	ForEach(func(*PullRequest) error) error
}

type Branch struct {
	Name     string
	SHA      string
	CloneURL string
}
