package mongosrc

import (
	"context"
	"errors"
	"time"

	"github.com/repofuel/repofuel/accounts/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type authProviderDataSource struct {
	collection *mongo.Collection
}

func NewAuthProviderDataSource(db *mongo.Database) *authProviderDataSource {
	return &authProviderDataSource{collection: db.Collection("authenticators")}
}

var syncOpts = options.Update().SetUpsert(true)

func (db *authProviderDataSource) InsertOrUpdate(ctx context.Context, au *entity.AuthProvider) error {
	if au.ID == "" {
		return errors.New("authenticator ID cannot be empty")
	}

	now := time.Now()
	au.UpdatedAt = now

	_, err := db.collection.UpdateOne(ctx, bson.M{
		"_id": au.ID,
	}, bson.M{
		"$set": au,
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}, syncOpts)
	return err
}

func (db *authProviderDataSource) FindByID(ctx context.Context, id string) (*entity.AuthProvider, error) {
	return db.findOne(ctx, bson.M{"_id": id})
}

func (db *authProviderDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.AuthProvider, error) {
	var doc entity.AuthProvider
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}
