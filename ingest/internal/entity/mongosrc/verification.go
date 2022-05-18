package mongosrc

import (
	"context"
	"reflect"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/pkg/codec"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type verificationDataSource struct {
	collection *mongo.Collection
}

func NewVerificationDataSource(ctx context.Context, db *mongo.Database) *verificationDataSource {
	interfaceCodec := codec.NewInterfaceCodec("PayloadType",
		&entity.LinkingVerificationOauth1{},
	)
	rb := bson.NewRegistryBuilder()
	interfaceCodec.RegisterInterfaceCodec(rb, reflect.TypeOf((*entity.VerificationPayload)(nil)).Elem())

	c := db.Collection("verification", options.Collection().SetRegistry(rb.Build()))

	_, err := c.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"expired_at": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("create index for verification")
	}

	return &verificationDataSource{collection: c}
}

func (db *verificationDataSource) FindByID(ctx context.Context, id string) (*entity.Verification, error) {
	return db.findOne(ctx, bson.M{
		"_id": id,
	})
}

func (db *verificationDataSource) Insert(ctx context.Context, v *entity.Verification) error {
	v.CreatedAt = time.Now()
	_, err := db.collection.InsertOne(ctx, v)
	return err
}

func (db *verificationDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Verification, error) {
	var doc entity.Verification
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrVerificationNotExist
		}
		return nil, err
	}

	return &doc, nil
}
