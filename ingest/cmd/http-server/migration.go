package main

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/keys"
	"github.com/repofuel/repofuel/ingest/internal/configs"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/ghapp"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func migrateData(ctx context.Context, cfg *configs.Configs, providersDB entity.ProviderDataSource, mgr *manage.Manager, db *mongo.Database) error {
	err := registerProvidersFromConfiguration(ctx, cfg, providersDB)
	if err != nil {
		return err
	}

	err = realyzeRepositoires(ctx, db, mgr)
	if err != nil {
		return err
	}

	count, err := db.Collection("organizations").EstimatedDocumentCount(ctx)
	if err != nil {
		return err
	}

	//if there are organization, no need to re add repositories
	if count != 0 {
		return nil
	}

	db.Collection("repositories").DeleteMany(ctx, bson.M{})
	db.Collection("pulls").DeleteMany(ctx, bson.M{})
	db.Collection("jobs").DeleteMany(ctx, bson.M{})
	db.Collection("commits").DeleteMany(ctx, bson.M{})
	db.Collection("models").DeleteMany(ctx, bson.M{})
	db.Collection("tokens").DeleteMany(ctx, bson.M{})

	it, err := mgr.Integrations.ServiceProvider(ctx, "github")
	if err != nil {
		return err
	}
	ghint := it.(*ghapp.GithubApp)
	err = ghint.AddAllRepository(ctx)
	if err != nil {
		return err
	}

	return nil
}

// deprecated, should be removed later
func registerProvidersFromConfiguration(ctx context.Context, cfg *configs.Configs, providersDB entity.ProviderDataSource) error {
	gh := cfg.Providers.Github
	s, err := url.Parse(gh.Server)
	if err != nil {
		return err
	}

	u := s.String()
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}

	_ = providersDB.Insert(ctx, &entity.Provider{
		ID:          "github",
		Name:        "Github",
		Server:      u,
		Platform:    common.SystemGithub,
		SourceCode:  true,
		Issues:      true,
		Webhook:     true,
		LoginWith:   entity.LoginOauth2,
		AuthMethods: nil,
		Config: &entity.GithubAppConfig{
			Server:        s,
			AppID:         gh.AppID,
			AppName:       gh.AppName,
			WebhookSecret: credentials.String(gh.WebhookSecret),
			PrivateKey:    gh.PrivateKey.Key(),
			OAuth2: &credentials.OAuth2{
				ClientID:     "",
				ClientSecret: "",
				AuthURL:      "",
				TokenURL:     "",
				AuthStyle:    0,
				RedirectURL:  "",
				Scopes:       nil,
			},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	})

	_ = providersDB.Insert(ctx, &entity.Provider{
		ID:       "jira_server",
		Name:     "Jira",
		Server:   "http://localhost:8080/",
		Platform: common.SystemJiraServer,
	})

	_ = providersDB.Insert(ctx, &entity.Provider{
		ID:       "jira_cloud",
		Name:     "Jira",
		Server:   "http://localhost:8080/",
		Platform: common.SystemJiraCloud,
	})

	pk, err := keys.LoadRSAPrivateKey("../repofuel-containers/configs/localhost/keys/jira-rsa.pem")
	if err != nil {
		// no need to register the rest
		return nil
	}

	_ = providersDB.Insert(ctx, &entity.Provider{
		ID:       "jira_localhost",
		Name:     "Jira LocalHost",
		Server:   "http://localhost:8080/",
		Platform: common.SystemJiraServer,
		//SourceCode:  true,
		Issues: true,
		//Webhook:     true,
		LoginWith: entity.LoginDisabled,
		AuthMethods: []entity.AuthMethod{
			entity.OAuth,
			entity.BasicAuth,
		},
		Config: &entity.JiraAppLinkConfig{
			Server:       "http://localhost:8080/",
			ConsumerName: "oauth-sample-consumer",
			OAuth1: credentials.OAuth1{
				ConsumerKey:     "oauth-sample-consumer",
				ConsumerSecret:  "",
				CallbackURL:     "http://localhost:3002",
				RequestTokenURL: "http://localhost:8080/plugins/servlet/oauth/request-token",
				AuthorizeURL:    "http://localhost:8080/plugins/servlet/oauth/authorize",
				AccessTokenURL:  "http://localhost:8080/plugins/servlet/oauth/access-token",
				Realm:           "",
				PrivateKey:      pk,
			},
			PublicKey: "",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	})

	return nil

}

func realyzeRepositoires(ctx context.Context, db *mongo.Database, mgr *manage.Manager) error {
	repos := db.Collection("repositories")

	cur, err := repos.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	for cur.Next(ctx) {
		var repo entity.Repository
		err = cur.Decode(&repo)
		if err != nil {
			return err
		}
		if repo.DataVersion == entity.CurrentDataVersion {
			continue
		}

		repo.CreatedAt = time.Time{}
		repo.Branches = nil
		repo.Confidence = 0
		repo.Quality = 0
		repo.DataVersion = 0
		repo.Status = 0

		_, err = db.Collection("pulls").DeleteMany(ctx, bson.M{
			"repo_id": repo.ID,
		})
		if err != nil {
			return err
		}

		_, err = db.Collection("commits").DeleteMany(ctx, bson.M{
			"_id.r": repo.ID,
		})
		if err != nil {
			return err
		}

		_, err = db.Collection("models").DeleteMany(ctx, bson.M{
			"repo_id": repo.ID.Hex(),
		})
		if err != nil {
			return err
		}

		db.Collection("repositories").UpdateOne(ctx, bson.M{
			"_id": repo.ID,
		}, bson.M{
			"$unset": bson.M{
				"branches":   "",
				"quality":    "",
				"status":     "",
				"confidence": "",
			},
			"$set": bson.M{
				"version": entity.CurrentDataVersion,
			},
		})

		err = mgr.AddRepository(ctx, &repo)
		if err != nil {
			return err
		}
	}

	return nil
}
