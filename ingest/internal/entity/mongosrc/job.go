package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/invoke"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const jobsCollection = "jobs"

type jobDataSource struct {
	collection *mongo.Collection
}

func NewJobDataSource(db *mongo.Database) *jobDataSource {
	return &jobDataSource{
		collection: db.Collection(jobsCollection),
	}
}

func (db *jobDataSource) TotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"_id": bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)}}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *jobDataSource) CountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{"_id": bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)}}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, bson.M{"$toDate": "$_id"}, frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

func (db *jobDataSource) RepositoryJobConnection(repoID identifier.RepositoryID, direction *entity.OrderDirection, pageCfg *entity.PaginationInput) entity.JobConnection {
	filter := bson.M{"repo_id": repoID}

	orderCfg := &orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: defaultDescIndex,
		AscIndex:  defaultAscIndex,
	}

	return newJobConnection(db.collection, filter, pageCfg, orderCfg, defaultCursorParser)
}

func (db *jobDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Job, error) {
	var doc entity.Job
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrJobNotExist
		}
		return nil, err
	}

	return &doc, nil
}

var findByRepoOpts = options.Find().SetSort(bson.M{"_id": -1}).SetLimit(50)

func (db *jobDataSource) FindByRepo(ctx context.Context, repoID identifier.RepositoryID) (entity.JobIter, error) {
	return db.find(ctx, bson.M{
		"repo_id": repoID,
	}, findByRepoOpts)
}

var sortJobsDisc = options.FindOne().SetSort(bson.M{"_id": -1})

func (db *jobDataSource) FindLast(ctx context.Context, repoID identifier.RepositoryID) (*entity.Job, error) {
	return db.findOne(ctx, bson.M{
		"repo_id":      repoID,
		"log":          bson.M{"$exists": true},
		"log.0.status": bson.M{"$ne": status.Failed},
	}, sortJobsDisc)
}

func (db *jobDataSource) FindLastWithStatus(ctx context.Context, repoID identifier.RepositoryID, status ...status.Stage) (*entity.Job, error) {
	query := make(bson.A, len(status))
	for i := range status {
		query[i] = bson.M{"log.status": status[i]}
	}

	return db.findOne(ctx, bson.M{
		"repo_id": repoID,
		"$and":    query,
	}, sortJobsDisc)
}

//todo: have a better name
func (db *jobDataSource) CreateJob(ctx context.Context, id identifier.RepositoryID, a invoke.Action, details map[string]interface{}) (identifier.JobID, error) {
	result, err := db.collection.InsertOne(ctx, &entity.Job{
		Invoker:    a,
		Repository: id,
		Details:    details,
	})
	if err != nil {
		return identifier.JobID{}, err
	}
	return identifier.JobID(result.InsertedID.(primitive.ObjectID)), nil

}

func (db *jobDataSource) SaveStatus(ctx context.Context, id identifier.JobID, s status.Stage) error {
	_, err := db.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$push": bson.M{"log": &entity.Update{
			Status:    s,
			StartedAt: time.Now(),
		}}})

	return err
}

func (db *jobDataSource) ReportError(ctx context.Context, id identifier.JobID, report error) error {
	_, err := db.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"error": report.Error()},
			"$push": bson.M{
				"log": &entity.Update{
					Status:    status.Failed,
					StartedAt: time.Now(),
				}},
		})

	return err
}

func (db *jobDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.JobIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return newEntityJobIter(cur), nil
}
