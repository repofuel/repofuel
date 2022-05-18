package permission

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const ViewerCtxKey = "viewer"

//go:generate stringer -type=Role -linecomment  -trimprefix Role
//go:generate jsonenums -type=Role
type Role uint8

const (
	RoleUser Role = iota // USER
	_
	_
	_
	_
	RoleSiteAdmin // SITE_ADMIN
	RoleService   // SERVICE
)

var _RoleEnumToActionValue = make(map[string]Role, len(_RoleValueToName))
var _RoleValueToQuotedActionEnumTo = make(map[Role]string, len(_RoleValueToName))

func init() {
	for k := range _RoleValueToName {
		name := strings.ReplaceAll(strings.ToUpper(k.String()), " ", "_")
		_RoleEnumToActionValue[name] = k
		_RoleValueToQuotedActionEnumTo[k] = strconv.Quote(name)
	}

}

func (t *Role) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	tag, ok := _RoleEnumToActionValue[s]
	if !ok {
		return fmt.Errorf("invalid Action %q", s)
	}
	*t = tag
	return nil
}

func (t Role) MarshalGQL(w io.Writer) {
	v, ok := _RoleValueToQuotedActionEnumTo[t]
	if !ok {
		fmt.Fprint(w, t.String())
		return
	}

	fmt.Fprint(w, v)
}

func (r Role) HasFullAccess() bool {
	return r == RoleSiteAdmin || r == RoleService
}

type UserID = primitive.ObjectID

type AccessInfo struct {
	Role Role `json:"r,omitempty"`
	*ServiceInfo
	*UserInfo
}

type UserInfo struct {
	UserID    UserID            `json:"id,omitempty"` // UserID is the user ID on Repofuel
	Providers map[string]string `json:"pids"`         // Providers is a mapping to the user ID on each provider
}

type ServiceInfo struct {
	ServiceID string `json:"sid,omitempty"`
}

func ViewerCtx(ctx context.Context) *AccessInfo {
	viewer, _ := ctx.Value(ViewerCtxKey).(*AccessInfo)
	return viewer
}

func OnlyServiceAccounts(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		access, ok := r.Context().Value(ViewerCtxKey).(*AccessInfo)
		if !ok || access.Role != RoleService {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func OnlyAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		access, ok := r.Context().Value(ViewerCtxKey).(*AccessInfo)
		if !ok || access.Role != RoleSiteAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// todo: use this function to refactor OnlyServiceAccounts and OnlyAdmin (after doing a benchmark)
func only(p Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		access, ok := r.Context().Value(ViewerCtxKey).(*AccessInfo)
		if !ok || access.Role != p {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
