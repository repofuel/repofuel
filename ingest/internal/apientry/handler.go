// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package apientry

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/repofuel/repofuel/ingest/internal/rest"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	restHandler     *rest.Handler
	graphqlHandler  http.Handler
	providerHandler *manage.IntegrationManager
}

func Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

				log.Ctx(r.Context()).Error().
					Interface(zerolog.ErrorStackFieldName, rvr).
					Msg("recover a panic in a request")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

var AccessHandler = hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
	log.Ctx(r.Context()).Info().
		Str("method", r.Method).
		Str("url", r.RequestURI).
		Int("status", status).
		Int("size", size).
		Dur("duration", duration).
		Msg("incoming request")
})

func NewHandler(restHandler *rest.Handler, providerHandler *manage.IntegrationManager, graphqlHandler http.Handler) *Handler {
	return &Handler{
		restHandler:     restHandler,
		graphqlHandler:  graphqlHandler,
		providerHandler: providerHandler,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(
		Recoverer,
		AccessHandler,
	)

	r.Mount("/debug", middleware.Profiler())

	r.Mount("/ingest", h.restHandler.Routes())
	r.Mount("/ingest/graphql", h.graphqlHandler)
	r.Mount("/ingest/playground", playground.Handler("GraphQL playground", "/ingest/graphql"))

	r.HandleFunc("/ingest/apps/{provider}/*", h.providerHandler.HandelProviders)

	return r
}
