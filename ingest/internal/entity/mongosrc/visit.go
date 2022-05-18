package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const visitCollection = "visits"

type visitDataSource struct {
	collection *mongo.Collection
}

func NewVisitDataSource(db *mongo.Database) *visitDataSource {
	return &visitDataSource{collection: db.Collection(visitCollection)}
}

func (db *visitDataSource) Insert(ctx context.Context, visit *entity.Visit) error {
	visit.CreatedAt = time.Now()
	_, err := db.collection.InsertOne(ctx, visit)
	return err
}

func (db *visitDataSource) VisitorsTotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	res, err := db.collection.Distinct(ctx, "user_id", filter)
	return int64(len(res)), err
}

func (db *visitDataSource) ViewsTotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	return db.collection.CountDocuments(ctx, filter)
}

var CountUniqueSummery = bson.E{Key: "count", Value: bson.M{"$sum": 1}}

func (db *visitDataSource) CountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.VisitOverTime, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}

	cur, err := db.collection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.M{
				"date": bson.M{
					"$dateToString": bson.M{
						"format": frequencyToDateFormat(frequency),
						"date":   "$created_at",
					},
				},
				"visitor": "$user_id",
			}},
			{Key: "count", Value: bson.M{"$sum": 1}},
		},
		}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id.date"},
			{Key: "visitors", Value: bson.M{"$sum": 1}},
			{Key: "views", Value: bson.M{"$sum": "$count"}},
		},
		}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	})
	if err != nil {
		return nil, err
	}

	var res []*entity.VisitOverTime
	err = cur.All(ctx, &res)
	return res, err
}
