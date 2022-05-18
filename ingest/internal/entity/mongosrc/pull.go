package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const pullCollection = "pulls"

type pullDataSource struct {
	collection *mongo.Collection
}

func (db *pullDataSource) FindByID(ctx context.Context, id identifier.PullRequestID) (*entity.PullRequest, error) {
	return db.findOne(ctx, bson.M{"_id": id})
}

func NewPullDataSource(db *mongo.Database) *pullDataSource {
	return &pullDataSource{
		collection: db.Collection(pullCollection),
	}
}

func (db *pullDataSource) Insert(ctx context.Context, pull *entity.PullRequest) error {
	pull.CreatedAt = time.Now()
	pull.UpdatedAt = pull.CreatedAt
	result, err := db.collection.InsertOne(ctx, pull)
	if err != nil {
		return err
	}
	pull.ID = identifier.PullRequestID(result.InsertedID.(primitive.ObjectID))
	return nil
}

func (db *pullDataSource) FindByNumber(ctx context.Context, repoID identifier.RepositoryID, number int) (*entity.PullRequest, error) {
	return db.findOne(ctx, bson.M{
		"repo_id":       repoID,
		"source.number": number,
	})
}

func (db *pullDataSource) StatusByID(ctx context.Context, id identifier.PullRequestID) (status.Stage, error) {
	var obj struct {
		Status status.Stage
	}

	err := db.collection.FindOne(ctx, bson.M{"_id": id}, statusByIdOptions).Decode(&obj)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, entity.ErrRepositoryNotExist
		}
		return 0, err
	}

	return obj.Status, nil
}

func (db *pullDataSource) AnalyzedTotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *pullDataSource) AnalyzedCountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, "$created_at", frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

var findAnalyzedHeadsOpts = options.Find().SetProjection(bson.M{"analyzed": 1})

type analyzedHashes struct {
	Analyzed identifier.Hash `bson:"analyzed"`
}

func (db *pullDataSource) FindAnalyzedHeads(ctx context.Context, repoID identifier.RepositoryID) ([]identifier.Hash, error) {
	var all []identifier.Hash
	var r analyzedHashes

	cur, err := db.collection.Find(ctx, bson.M{
		"repo_id":       repoID,
		"analyzed":      bson.M{"$exists": true},
		"source.merged": false,
	}, findAnalyzedHeadsOpts)
	if err != nil {
		return nil, err
	}

	for cur.Next(ctx) {
		err = cur.Decode(&r)
		if err != nil {
			return nil, err
		}
		all = append(all, r.Analyzed)
	}

	return all, nil
}

var defaultAscIndex = bson.D{
	{Key: "_id", Value: 1},
}

var defaultDescIndex = bson.D{
	{Key: "_id", Value: -1},
}

func (db *pullDataSource) RepositoryPullRequestConnection(repoID identifier.RepositoryID, directionInput *entity.OrderDirection, pageCfg *entity.PaginationInput) entity.PullRequestConnection {
	filter := bson.M{"repo_id": repoID}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(directionInput, entity.OrderDirectionDesc),
		DescIndex: defaultDescIndex,
		AscIndex:  defaultAscIndex,
	}

	return newPullRequestConnection(db.collection, filter, pageCfg, &orderCfg, defaultCursorParser)
}

var updateSourceOpts = options.Update().SetUpsert(true)

func (db *pullDataSource) SaveStatus(ctx context.Context, id identifier.PullRequestID, s status.Stage) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": s}})
}

func (db *pullDataSource) SaveAnalyzedHead(ctx context.Context, id identifier.PullRequestID, hash identifier.Hash) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"analyzed": hash}})
}

func (db *pullDataSource) updateOne(ctx context.Context, filter interface{}, update bson.M, opts ...*options.UpdateOptions) error {
	return updateOne(ctx, db.collection, filter, update, opts...)
}

func updateOne(ctx context.Context, c *mongo.Collection, filter interface{}, update bson.M, opts ...*options.UpdateOptions) error {
	if set, ok := update["$set"]; ok {
		switch value := set.(type) {
		case bson.M:
			set.(bson.M)["updated_at"] = time.Now()
		case bson.D:
			update["$set"] = append(set.(bson.D), bson.E{Key: "updated_at", Value: time.Now()})
		case entity.Updatable:
			value.SetUpdatedNow()
		}
	}
	_, err := c.UpdateOne(ctx, filter, update, opts...)
	return err
}

var findAndUpdateSourceOpts = options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

func (db *pullDataSource) FindAndUpdateSource(ctx context.Context, repoID identifier.RepositoryID, source *common.PullRequest) (*entity.PullRequest, error) {
	r := db.collection.FindOneAndUpdate(ctx, bson.M{
		"repo_id":       repoID,
		"source.number": source.Number,
	}, bson.M{
		"$set": bson.M{
			"source": source,
		},
	}, findAndUpdateSourceOpts)

	var doc entity.PullRequest
	err := r.Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (db *pullDataSource) UpdateSource(ctx context.Context, repoID identifier.RepositoryID, source *common.PullRequest) (bool, error) {
	r, err := db.collection.UpdateOne(ctx, bson.M{
		"repo_id":       repoID,
		"source.number": source.Number,
	}, bson.M{
		"$set": bson.M{
			"source": source,
		},
	}, updateSourceOpts)
	if err != nil {
		return false, wrapError(err)
	}

	return r.UpsertedCount > 0, nil
}

func (db *pullDataSource) FindByRepoID(ctx context.Context, id identifier.RepositoryID, opts ...*options.FindOptions) (entity.PullRequestIter, error) {
	return db.find(ctx, bson.M{
		"repo_id": id,
	}, opts...)
}

func (db *pullDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.PullRequest, error) {
	var doc entity.PullRequest
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		return nil, wrapError(err)
	}

	return &doc, nil
}

func (db *pullDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.PullRequestIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return newEntityPullRequestIter(cur), nil
}

func wrapError(err error) error {
	if err == mongo.ErrNoDocuments {
		return entity.ErrPullRequestNotExist
	}
	return err
}
