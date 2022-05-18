package mongosrc

import (
	"context"
	"crypto/cipher"
	"reflect"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/pkg/codec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type providerDataSource struct {
	gcm        cipher.AEAD
	collection *mongo.Collection
}

func NewProviderDataSource(db *mongo.Database, gcm cipher.AEAD) *providerDataSource {
	interfaceCodec := codec.NewInterfaceCodec("Driver",
		&entity.BitbucketAppLinkConfig{},
		&entity.JiraAppLinkConfig{},
		&entity.GithubAppConfig{},
	)
	rb := codec.NewRegistryWithEncryption(gcm)
	interfaceCodec.RegisterInterfaceCodec(rb, reflect.TypeOf((*entity.ProviderConfig)(nil)).Elem())

	return &providerDataSource{collection: db.Collection("providers", &options.CollectionOptions{
		Registry: rb.Build(),
	})}
}

func (db *providerDataSource) Insert(ctx context.Context, p *entity.Provider) error {
	_, err := db.collection.InsertOne(ctx, p)
	return err
}

func (db *providerDataSource) FindByID(ctx context.Context, id string) (*entity.Provider, error) {
	return db.findOne(ctx, bson.M{"_id": id})
}

func (db *providerDataSource) FindByServer(ctx context.Context, server string) (*entity.Provider, error) {
	return db.findOne(ctx, bson.M{"server": server})
}

func (db *providerDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Provider, error) {
	var doc entity.Provider
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}
