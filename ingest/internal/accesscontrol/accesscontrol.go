package accesscontrol

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/rs/zerolog/log"
)

const (
	RepositoryCtxKey   = "repository"
	OrganizationCtxKey = "organization"
)

func OnlyOrganizationAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if MemberRole(r.Context()) != common.OrgAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func OnlyCollaborators(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !UserPermissions(r.Context()).Read {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func OnlyRepositoryAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !UserPermissions(r.Context()).Admin {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func UserPermissions(ctx context.Context) common.Permissions {
	viewer := permission.ViewerCtx(ctx)
	if viewer.Role == permission.RoleService || viewer.Role == permission.RoleSiteAdmin {
		// todo: maybe not all services need full permissions
		return common.FullPermissions
	}

	repo := ctx.Value(RepositoryCtxKey).(*entity.Repository)

	// get the user JobID for the repository provider
	userId := viewer.UserInfo.Providers[repo.ProviderSCM]

	p, ok := repo.Collaborators[userId]
	if !ok && !repo.Source.Private {
		return common.Permissions{
			Read: true,
		}
	}

	return p
}

func MemberRole(ctx context.Context) common.OrgRole {
	viewer := permission.ViewerCtx(ctx)
	if viewer.Role == permission.RoleService || viewer.Role == permission.RoleSiteAdmin {
		// todo: maybe not all services need full permissions
		return common.OrgAdmin
	}

	org := ctx.Value(OrganizationCtxKey).(*entity.Organization)

	userId := viewer.UserInfo.Providers[org.ProviderSCM]

	return org.Members[userId].Role
}

func CtxRepositoryBySourceID(reposDB entity.RepositoryDataSource) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			repo, err := reposDB.FindByProviderID(ctx,
				chi.URLParam(r, "platform"),
				chi.URLParam(r, "source_id"))
			if err != nil {
				if err == entity.ErrRepositoryNotExist {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				log.Ctx(ctx).Err(err).Msg("find repository")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx = context.WithValue(ctx, RepositoryCtxKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CtxOrganizationByID(organizationsDB entity.OrganizationDataSource) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			orgID, err := identifier.OrganizationIDFromHex(chi.URLParam(r, "org_id"))
			if err != nil {
				err = fmt.Errorf("invalid organization ID: %w", err)
				w.WriteHeader(http.StatusBadRequest)
				log.Ctx(ctx).Err(err).Msg("parse organization id")
				return
			}

			org, err := organizationsDB.FindByID(ctx, orgID)
			if err != nil {
				if err == entity.ErrRepositoryNotExist {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				log.Ctx(ctx).Err(err).Msg("fetch organization form DB")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx = context.WithValue(ctx, OrganizationCtxKey, org)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
