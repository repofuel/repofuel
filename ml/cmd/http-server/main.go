package main

import (
	"context"
	"log"
	"net/http"

	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ml/internal/apientry"
	"github.com/repofuel/repofuel/ml/internal/configs"
	"github.com/repofuel/repofuel/ml/internal/entity/mongosrc"
	"github.com/repofuel/repofuel/ml/internal/rest"
	"github.com/repofuel/repofuel/ml/pkg/ml"
	"github.com/repofuel/repofuel/pkg/mongocon"
)

const ServiceName = "ai"

func main() {
	cfg, err := configs.Parse()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := mongocon.NewDatabase(ctx, nil, &cfg.DB)
	if err != nil {
		log.Fatal("database connection: ", err)
	}

	modelsDB := mongosrc.NewModelDataSource(db)

	auth := jwtauth.NewAuthenticator(ServiceName, cfg.Keys.PrivateKey, nil, nil)

	tokenSrc := auth.TokenSource(ctx, &permission.AccessInfo{
		Role:        permission.RoleService,
		ServiceInfo: nil,
	})

	ingestURL := cfg.Repofuel.BaseURLs["ingest"].String()
	mlServer := ml.NewModelServer(tokenSrc, modelsDB, ingestURL)
	authCheck := jwtauth.NewAuthCheck(jwtauth.NewLocalKeySource(cfg.Keys.PublicKeys))
	restAPI := rest.NewHandler(authCheck, mlServer, modelsDB)
	apiEntry := apientry.NewHandler(restAPI)

	log.Println("starting server at:", "3004")
	if err := http.ListenAndServe(":3004", apiEntry.Routes()); err != nil {
		log.Fatal(err)
	}
}
