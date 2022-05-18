// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/graph"
	"github.com/repofuel/repofuel/ingest/internal/apientry"
	"github.com/repofuel/repofuel/ingest/internal/configs"
	"github.com/repofuel/repofuel/ingest/internal/entity/mongosrc"
	"github.com/repofuel/repofuel/ingest/internal/rest"
	"github.com/repofuel/repofuel/ingest/internal/version"
	"github.com/repofuel/repofuel/ingest/pkg/atlassian"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/ingest/pkg/usage"
	"github.com/repofuel/repofuel/pkg/mongocon"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const serviceName = "ingest"

func main() {
	log.Logger = logger()
	ctx := log.Logger.WithContext(context.Background())

	log.Info().
		Str("version", version.Version).
		Str("build_date", version.BuildDate).
		Str("version_summery", version.GitSummary).
		Msg("launch server")

	cfg, err := configs.Parse(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed parsing the configurations")
	}

	db, err := mongocon.NewDatabase(ctx, nil, &cfg.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed in connecting to database")
	}

	var (
		commitsDB       = mongosrc.NewCommitDataSource(ctx, db)
		montorDB        = mongosrc.NewMontorDataSource(db)
		reposDB         = mongosrc.NewRepositoryDataSource(db, montorDB)
		jobsDB          = mongosrc.NewJobDataSource(db)
		pullsDB         = mongosrc.NewPullDataSource(db)
		providersDB     = mongosrc.NewProviderDataSource(db, cfg.Keys.EncryptionKey)
		organizationsDB = mongosrc.NewOrganizationDataSource(db, cfg.Keys.EncryptionKey)
		verificationDB  = mongosrc.NewVerificationDataSource(ctx, db)
		visitDB         = mongosrc.NewVisitDataSource(db)
		feedbackDB      = mongosrc.NewFeedbackDataSource(db)
	)

	auth := jwtauth.NewAuthenticator(serviceName, cfg.Keys.PrivateKey, nil, nil)

	tokenSrc := auth.TokenSource(ctx, &permission.AccessInfo{
		Role:        permission.RoleService,
		ServiceInfo: nil,
	})
	rfc := repofuel.NewClient(oauth2.NewClient(ctx, tokenSrc), &cfg.Repofuel)

	authCheck := jwtauth.NewAuthCheck(jwtauth.NewLocalKeySource(cfg.Keys.PublicKeys))

	manager := manage.NewManager(ctx, manage.ManagerServices{
		Provider:     providersDB,
		Commit:       commitsDB,
		Repo:         reposDB,
		Job:          jobsDB,
		PullRequest:  pullsDB,
		Organization: organizationsDB,
		Verification: verificationDB,
		Monitor:      montorDB,
		Repofuel:     rfc,
	}, authCheck)
	if err := manager.Recover(ctx); err != nil {
		log.Fatal().Err(err).Msg("problem in recovering the stuck repositories")
	}

	jiraMgr := atlassian.NewJiraIntegrationManager(providersDB)

	tracker := usage.NewTracker(visitDB)
	graphqlHandler := graph.NewGraphQLServer(authCheck, tracker, &graph.Resolver{
		RepofuelClient: rfc,
		FeedbackDB:     feedbackDB,
		CommitDB:       commitsDB,
		RepositoryDB:   reposDB,
		PullRequestDB:  pullsDB,
		JobDB:          jobsDB,
		OrganizationDB: organizationsDB,
		VisitDB:        visitDB,
		MonitorDB:      montorDB,
		Manager:        manager,
		Observables:    manager.ProgressObservableRegistry(),
	})

	restHandler := rest.NewHandler(&log.Logger, authCheck, manager, reposDB, commitsDB, jobsDB, pullsDB, organizationsDB, feedbackDB, jiraMgr)
	apiEntry := apientry.NewHandler(restHandler, manager.Integrations, graphqlHandler)

	err = migrateData(ctx, cfg, providersDB, manager, db)
	if err != nil {
		log.Fatal().Err(err).Msg("error in data migration")
	}

	go manager.Run()

	port := 3002
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		Handler:     apiEntry.Routes(),
		BaseContext: func(listener net.Listener) context.Context { return ctx },
	}

	log.Info().Int("port", port).Msg("starting HTTP server")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed in starting HTTP server")
	}
}

func logger() zerolog.Logger {
	var dev = flag.Bool("dev", false, "to run in development mode")
	if !flag.Parsed() {
		flag.Parse()
	}

	level := zerolog.InfoLevel
	out := io.Writer(os.Stderr)

	if *dev {
		level = zerolog.TraceLevel
		out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC1123}
	}

	return zerolog.New(out).Level(level).With().Timestamp().Logger()
}
