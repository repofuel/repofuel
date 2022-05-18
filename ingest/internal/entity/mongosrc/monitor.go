package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const montorCollection = "monitor"

type montorDataSource struct {
	collection *mongo.Collection
}

func NewMontorDataSource(db *mongo.Database) *montorDataSource {
	return &montorDataSource{collection: db.Collection(montorCollection)}
}

var lastMonitorOpts = options.FindOne().SetSort(bson.M{"create_at": -1}).SetProjection(bson.M{"user_id": 1})

func (db *montorDataSource) LastRepositoryMonitorUserID(ctx context.Context, repoID identifier.RepositoryID) (identifier.UserID, error) {
	var doc entity.Monitor
	err := db.collection.FindOne(ctx, bson.M{
		"_id.r": repoID,
	}, lastMonitorOpts).Decode(&doc)
	if err != nil {
		return identifier.UserID{}, err
	}

	return doc.ID.UserID, nil
}

var isUserMonitoringOpts = options.Count().SetLimit(1)

func (db *montorDataSource) IsMonitor(ctx context.Context, id *identifier.MonitorID) (bool, error) {
	count, err := db.collection.CountDocuments(ctx, bson.M{
		"_id": id,
	}, isUserMonitoringOpts)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (db *montorDataSource) RemoveMonitor(ctx context.Context, id *identifier.MonitorID) error {
	_, err := db.collection.DeleteOne(ctx, bson.M{
		"_id": id,
	})
	return err
}
func (db *montorDataSource) MonitorCount(ctx context.Context, repoID identifier.RepositoryID) (int, error) {
	count, err := db.collection.CountDocuments(ctx, bson.M{"_id.r": repoID})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

var userReposIDOpts = options.Find().SetProjection(bson.M{"_id": 1})

func (db *montorDataSource) UserReposIDs(ctx context.Context, userID identifier.UserID) ([]identifier.RepositoryID, error) {
	cur, err := db.collection.Find(ctx, bson.M{
		"_id.u": userID,
	}, userReposIDOpts)
	if err != nil {
		return nil, err
	}

	var docs []struct {
		ID struct {
			RepoID identifier.RepositoryID `bson:"r"`
		} `bson:"_id"`
	}

	err = cur.All(ctx, &docs)
	if err != nil {
		return nil, err
	}

	ids := make([]identifier.RepositoryID, len(docs))
	for i := range docs {
		ids[i] = docs[i].ID.RepoID
	}

	return ids, nil
}

func (db *montorDataSource) InsertMonitor(ctx context.Context, id *identifier.MonitorID) error {
	_, err := db.collection.InsertOne(ctx, &entity.Monitor{
		ID:        id,
		CreatedAt: time.Now(),
	})

	return err
}
