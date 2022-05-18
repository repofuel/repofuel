package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const feedbackCollection = "feedback"

type feedbackDataSource struct {
	collection *mongo.Collection
}

func NewFeedbackDataSource(db *mongo.Database) *feedbackDataSource {
	return &feedbackDataSource{
		collection: db.Collection(feedbackCollection),
	}
}

func (db *feedbackDataSource) Insert(ctx context.Context, doc *entity.Feedback) error {
	doc.CreatedAt = time.Now()
	result, err := db.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	doc.ID = identifier.FeedbackID(result.InsertedID.(primitive.ObjectID))
	return nil
}

func (db *feedbackDataSource) FeedbackConnection(directionInput *entity.OrderDirection, pageCfg *entity.PaginationInput) entity.FeedbackConnection {
	filter := bson.M{}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(directionInput, entity.OrderDirectionDesc),
		DescIndex: defaultDescIndex,
		AscIndex:  defaultAscIndex,
	}

	return newFeedbackConnection(db.collection, filter, pageCfg, &orderCfg, defaultCursorParser)
}

func (db *feedbackDataSource) All(ctx context.Context) (entity.FeedbackIter, error) {
	filter := bson.M{}

	return db.find(ctx, filter)
}

func (db *feedbackDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.FeedbackIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return newEntityFeedbackIter(cur), nil
}
