// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/repofuel/repofuel/accounts/internal/apientry"
	"github.com/repofuel/repofuel/accounts/internal/configs"
	"github.com/repofuel/repofuel/accounts/internal/entity/mongosrc"
	"github.com/repofuel/repofuel/accounts/internal/rest"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/pkg/mongocon"
)

const ServiceName = "accounts"

func main() {
	cfg, err := configs.Parse()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := mongocon.NewDatabase(ctx, cfg.Keys.EncryptionKey, &cfg.DB)
	if err != nil {
		log.Fatal("database connection: ", err)
	}

	var (
		usersDB         = mongosrc.NewUserDataSource(db)
		tokensDB        = mongosrc.NewTokenDataSource(db)
		authProvidersDB = mongosrc.NewAuthProviderDataSource(db)
	)

	err = MigrateData(ctx, cfg, db, authProvidersDB)
	if err != nil {
		log.Fatal(err)
	}

	auth := jwtauth.NewAuthenticator(ServiceName, cfg.Keys.PrivateKey, usersDB, tokensDB)
	authCheck := jwtauth.NewAuthCheck(jwtauth.NewLocalKeySource(cfg.Keys.PublicKeys))
	restHandler := rest.NewHandler(authCheck, usersDB)

	providerManager := apientry.NewAuthProvidersManage(auth, authProvidersDB, usersDB)
	apiEntry := apientry.NewHandler(auth, providerManager, restHandler)

	log.Println("starting server at:", "3003")
	if err := http.ListenAndServe(":3003", apiEntry.Routes()); err != nil {
		log.Fatal(err)
	}
}
