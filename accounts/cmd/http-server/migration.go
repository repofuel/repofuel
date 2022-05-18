package main

import (
	"context"
	"log"
	"time"

	"github.com/repofuel/repofuel/accounts/internal/configs"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func MigrateData(ctx context.Context, cfg *configs.Configs, db *mongo.Database, authProvidersDB entity.AuthProviderDataSource) error {
	err := ChangeProviderShape(ctx, db)
	if err != nil {
		log.Println(err)
	}

	err = RegisterProvidersFromConfiguration(ctx, cfg, authProvidersDB)
	if err != nil {
		return err
	}

	return nil
}

func ChangeProviderShape(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")
	cur, err := collection.Find(ctx, bson.M{"providers": bson.M{"$not": bson.M{"$type": "array"}}})
	if err != nil {
		return err
	}

	docs := []struct {
		Id        permission.UserID       `json:"id"          bson:"_id,omitempty"`
		Username  string                  `json:"username"    bson:"username"`
		FirstName string                  `json:"first_name"  bson:"first_name"`
		LastName  string                  `json:"last_name"   bson:"last_name"`
		AvatarURL string                  `json:"avatar_url"  bson:"avatar_url,omitempty"`
		Email     string                  `json:"email"       bson:"email"`
		Password  string                  `json:"-"           bson:"password,omitempty"`
		Providers map[string]*common.User `json:"providers"   bson:"providers,omitempty"`
		Role      permission.Role         `json:"role"        bson:"role"`
		CreatedAt time.Time               `json:"-"           bson:"created_at"`
		UpdatedAt time.Time               `json:"-"           bson:"updated_at"`
	}{}

	err = cur.All(ctx, &docs)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		providers := make([]*common.User, 0, len(doc.Providers))
		for _, p := range doc.Providers {
			providers = append(providers, p)
		}

		_, err = collection.ReplaceOne(ctx, bson.M{"_id": doc.Id},
			&entity.User{
				Id:        doc.Id,
				Username:  doc.Username,
				FirstName: doc.FirstName,
				LastName:  doc.LastName,
				AvatarURL: doc.AvatarURL,
				Email:     doc.Email,
				Password:  doc.Password,
				Providers: providers,
				Role:      doc.Role,
				CreatedAt: doc.CreatedAt,
				UpdatedAt: doc.UpdatedAt,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

//deprecated, should be removed later
func RegisterProvidersFromConfiguration(ctx context.Context, cfg *configs.Configs, authProvidersDB entity.AuthProviderDataSource) error {
	gh := cfg.Oauth2Config(common.Github)
	_ = authProvidersDB.InsertOrUpdate(ctx, &entity.AuthProvider{
		ID:     "github",
		System: common.SystemGithub,
		Server: cfg.Providers.Github.Server,
		OAuth2: &credentials.OAuth2{
			ClientID:     gh.ClientID,
			ClientSecret: credentials.String(gh.ClientSecret),
			AuthURL:      gh.Endpoint.AuthURL,
			TokenURL:     gh.Endpoint.TokenURL,
			AuthStyle:    gh.Endpoint.AuthStyle,
			RedirectURL:  gh.RedirectURL,
			Scopes:       gh.Scopes,
		},
	})
	return nil
}
