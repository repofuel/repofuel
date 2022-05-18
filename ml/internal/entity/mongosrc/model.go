package mongosrc

import (
	"context"
	"github.com/repofuel/repofuel/ml/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type modelDataSource struct {
	collection *mongo.Collection
}

func NewModelDataSource(db *mongo.Database) *modelDataSource {
	return &modelDataSource{
		collection: db.Collection("models"),
	}
}

func (db *modelDataSource) Insert(ctx context.Context, m *entity.Model) error {
	result, err := db.collection.InsertOne(ctx, m)
	if err != nil {
		return err
	}
	m.ID = result.InsertedID.(entity.ModelID)
	return nil
}

func (db *modelDataSource) FindRepoModels(ctx context.Context, repoId string, opts ...*options.FindOptions) (entity.ModelIter, error) {
	return db.find(ctx, bson.M{
		"repo_id": repoId,
	}, opts...)

}

var findLastModelOpts = options.FindOne().SetSort(bson.M{"_id": -1})

func (db *modelDataSource) FindLatestModel(ctx context.Context, repoId string) (*entity.Model, error) {
	return db.findOne(ctx, bson.M{"repo_id": repoId}, findLastModelOpts)
}

func (db *modelDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Model, error) {
	var model entity.Model
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrModelNotExist
		}
		return nil, err
	}

	return &model, nil
}

func (db *modelDataSource) LogUsage(ctx context.Context, id entity.ModelID) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"last_use": time.Now(),
	}})
	if err != nil {
		return err
	}
	return nil
}

type modelIter struct {
	cur *mongo.Cursor
}

func newModelIter(cur *mongo.Cursor) *modelIter {
	return &modelIter{
		cur: cur,
	}
}

func (db *modelDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.ModelIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return newModelIter(cur), nil
}

func (iter *modelIter) ForEach(ctx context.Context, fun func(*entity.Model) error) error {
	defer iter.cur.Close(ctx)
	for iter.cur.Next(ctx) {
		var doc entity.Model
		if err := iter.cur.Decode(&doc); err != nil {
			return err
		}

		if err := fun(&doc); err != nil {
			return err
		}
	}
	return iter.cur.Err()
}
