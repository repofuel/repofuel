package graph

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/ingest/graph/generated"
	"github.com/repofuel/repofuel/ingest/pkg/usage"
	"github.com/rs/zerolog/log"
)

type middleware = func(http.Handler) http.Handler

func NewGraphQLServer(authCheck *jwtauth.AuthCheck, tracker *usage.Tracker, res *Resolver) http.Handler {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: res,
	}))

	srv.AddTransport(transport.Websocket{
		InitFunc:              WebsocketInit(authCheck),
		KeepAlivePingInterval: 10 * time.Second,
	})
	//srv.AddTransport(transport.Options{})
	//srv.AddTransport(transport.GET{})
	srv.AddTransport(POST{
		tracker:   tracker,
		authCheck: authCheck,
	})
	//srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}

type POST struct {
	tracker   *usage.Tracker
	authCheck *jwtauth.AuthCheck
	transport.POST
}

func (h POST) Do(w http.ResponseWriter, r *http.Request, exec graphql.GraphExecutor) {
	ctx, err := h.authCheck.AuthenticatedContext(r.Context(), jwtauth.StripBearerToken(r.Header.Get("Authorization")))
	if err != nil {
		log.Debug().Err(err).Msg("authenticate GraphQL POST request")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	go h.tracker.LogVisit(ctx, r.Header)

	h.POST.Do(w, r.WithContext(ctx), exec)
}

func WebsocketInit(authCheck *jwtauth.AuthCheck) transport.WebsocketInitFunc {
	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
		return authCheck.AuthenticatedContext(ctx, initPayload.Authorization())
	}
}

func AllowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		switch origin {
		case "https://www.graphqlbin.com":
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		next.ServeHTTP(w, r)
	})
}
