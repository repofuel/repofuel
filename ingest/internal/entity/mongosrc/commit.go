// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package mongosrc

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/classify"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const commitsCollection = "commits"

type commitDataSource struct {
	collection *mongo.Collection
}

func (db *commitDataSource) AnalyzedTotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"job": bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)}}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *commitDataSource) AnalyzedCountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{"job": bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)}}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, bson.M{"$toDate": "$job"}, frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

func (db *commitDataSource) PredictedTotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{
		"job":                    bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)},
		"analysis.bug_potential": bson.M{"$exists": true},
	}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *commitDataSource) RepositoryPredictionsTotalCount(ctx context.Context, id identifier.RepositoryID) (int64, error) {
	filter := bson.M{
		"_id.r":                  id,
		"analysis.bug_potential": bson.M{"$exists": true},
	}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *commitDataSource) PredictedCountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{
		"job":                    bson.M{"$gte": primitive.NewObjectIDFromTimestamp(since)},
		"analysis.bug_potential": bson.M{"$exists": true},
	}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, bson.M{"$toDate": "$job"}, frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

func (db *commitDataSource) SelectedCommitConnection(ctx context.Context, repoID identifier.RepositoryID, ids []identifier.Hash, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.CommitConnection, error) {
	filter := bson.M{
		"_id.r": repoID,
		"_id.h": bson.M{"$in": ids},
	}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: descCommitIndex,
		AscIndex:  ascCommitIndex,
	}

	return newCommitConnection(db.collection, filter, opts, &orderCfg, commitCursorParser(repoID)), nil
}

func NewCommitDataSource(ctx context.Context, db *mongo.Database) *commitDataSource {
	_, err := db.Collection(commitsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: ascCommitIndex,
	})
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("create index on commit collection")
	}

	return &commitDataSource{
		collection: db.Collection(commitsCollection),
	}
}

var insertOrReplaceOpts = options.Replace().SetUpsert(true)

func (db *commitDataSource) InsertOrReplace(ctx context.Context, c *entity.Commit) error {
	c.CreatedAt = time.Now()
	r, err := db.collection.ReplaceOne(ctx, bson.M{
		"_id": c.ID,
	}, c, insertOrReplaceOpts)
	if r != nil && r.MatchedCount > 0 {
		log.Ctx(ctx).Debug().Msg("commit already exists and has been replaced")
	}
	return err
}

func (db *commitDataSource) countOverTime(ctx context.Context, filter interface{}) ([]*entity.CountOverTime, error) {
	cur, err := db.overTime(ctx, filter, bson.E{Key: "count", Value: bson.M{"$sum": 1}})
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)

	return res, err
}

func (db *commitDataSource) avgOverTime(ctx context.Context, filter interface{}, field interface{}) ([]*entity.AvgOverTime, error) {
	cur, err := db.overTime(ctx, filter, bson.E{Key: "avg", Value: bson.M{"$avg": field}})
	if err != nil {
		return nil, err
	}

	var res []*entity.AvgOverTime
	err = cur.All(ctx, &res)

	return res, err
}

//todo: refactor: rename to e.g., authorDateOverTime and use the `overTime` function
func (db *commitDataSource) overTime(ctx context.Context, filter interface{}, summery bson.E) (*mongo.Cursor, error) {
	return db.collection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m",
					"date":   "$author.date",
				},
			}},
			summery,
		},
		}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	})
}

func frequencyToDateFormat(f entity.Frequency) string {
	switch f {
	case entity.FrequencyDaily:
		return "%Y-%m-%d"
	case entity.FrequencyMonthly:
		return "%Y-%m"
	case entity.FrequencyYearly:
		return "%Y"

	default:
		panic("this date frequency is not implemented")
	}
}

func overTime(ctx context.Context, collection *mongo.Collection, filter interface{}, summery bson.E, dateField interface{}, frequency entity.Frequency) (*mongo.Cursor, error) {
	return collection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.M{
				"$dateToString": bson.M{
					"format": frequencyToDateFormat(frequency),
					"date":   dateField,
				},
			}},
			summery,
		},
		}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	})
}

var listedTags = []classify.Tag{
	classify.Fix,
	classify.Feature,
	classify.Tests,
	classify.Documentations,
	classify.Refactor,
	classify.License,
	classify.CI,
	classify.TechnicalDebt,
}

func (db *commitDataSource) tagsCount(ctx context.Context, filter interface{}) ([]*entity.TagCount, error) {
	cur, err := db.collection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$unwind", Value: "$tags"}},
		{{Key: "$match", Value: bson.M{"tags": bson.M{"$in": listedTags}}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$tags",
			"count": bson.M{
				"$sum": 1,
			},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	})
	if err != nil {
		return nil, err
	}

	var res []*entity.TagCount
	err = cur.All(ctx, &res)

	return res, err
}

func (db *commitDataSource) BuggyCommitsOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*entity.CountOverTime, error) {
	return db.countOverTime(ctx, bson.M{
		"_id.r":   repoID,
		"fixes.0": bson.M{"$exists": true},
	})
}

func (db *commitDataSource) CommitsTagCount(ctx context.Context, repoID identifier.RepositoryID) ([]*entity.TagCount, error) {
	return db.tagsCount(ctx, bson.M{
		"_id.r": repoID,
	})
}

func (db *commitDataSource) CommitsOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*entity.CountOverTime, error) {
	return db.countOverTime(ctx, bson.M{
		"_id.r": repoID,
	})
}

func (db *commitDataSource) AvgEntropyOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*entity.AvgOverTime, error) {
	return db.avgOverTime(ctx, bson.M{
		"_id.r": repoID,
	}, "$metrics.entropy")
}

func (db *commitDataSource) AvgCommitFilesOverTime(ctx context.Context, repoID identifier.RepositoryID) ([]*entity.AvgOverTime, error) {
	return db.avgOverTime(ctx, bson.M{
		"_id.r": repoID,
	}, "$metrics.nf")
}

var ascCommitIndex = bson.D{
	{Key: "_id.r", Value: 1},
	{Key: "author.date", Value: 1},
	{Key: "_id.h", Value: 1},
}

var descCommitIndex = bson.D{
	{Key: "_id.r", Value: -1},
	{Key: "author.date", Value: -1},
	{Key: "_id.h", Value: -1},
}

func (db *commitDataSource) DeveloperEmails(ctx context.Context, repoID identifier.RepositoryID) ([]string, error) {
	filter := bson.M{"_id.r": repoID}

	res, err := db.collection.Distinct(ctx, "author.email", filter)
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(res))
	for i := range res {
		emails[i] = res[i].(string)
	}

	return emails, nil
}

func (db *commitDataSource) DeveloperNames(ctx context.Context, repoID identifier.RepositoryID) ([]string, error) {
	filter := bson.M{"_id.r": repoID}

	res, err := db.collection.Distinct(ctx, "author.name", filter)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(res))
	for i := range res {
		names[i] = res[i].(string)
	}

	return names, nil
}

func (db *commitDataSource) DevelopersAggregatedMetrics(ctx context.Context, repoID identifier.RepositoryID) (entity.ChangeMeasuresIter, error) {
	filter := bson.M{"_id.r": repoID}
	pipe := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":     "$author.email",
			"ns":      bson.M{"$avg": "$metrics.ns"},
			"nd":      bson.M{"$avg": "$metrics.nd"},
			"nf":      bson.M{"$avg": "$metrics.nf"},
			"entropy": bson.M{"$avg": "$metrics.entropy"},
			"la":      bson.M{"$avg": "$metric.la"},
			"ld":      bson.M{"$avg": "$metric.ld"},
			"ha":      bson.M{"$avg": "$metric.ha"},
			"lt":      bson.M{"$avg": "$metric.lt"},
			"ndev":    bson.M{"$max": "$metric.ndev"},
			"age":     bson.M{"$avg": "$metrics.age"},
			"nuc":     bson.M{"$max": "$metrics.nuc"},
			"exp":     bson.M{"$max": "$metrics.exp"},
			"rexp":    bson.M{"$max": "$metrics.rexp"},
			"sexp":    bson.M{"$max": "$metrics.sexp"},
		}}},
	}

	cur, err := db.collection.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	return newMetricsChangeMeasuresIter(cur), nil
}

func (db *commitDataSource) PullRequestCommitConnection(ctx context.Context, repoID identifier.RepositoryID, pullID identifier.PullRequestID, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.CommitConnection, error) {
	filter := bson.M{"_id.r": repoID, "pulls": pullID}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: descCommitIndex,
		AscIndex:  ascCommitIndex,
	}

	return newCommitConnection(db.collection, filter, opts, &orderCfg, commitCursorParser(repoID)), nil
}

func (db *commitDataSource) RepositoryCommitConnection(ctx context.Context, repoID identifier.RepositoryID, direction *entity.OrderDirection, filters *entity.CommitFilters, opts *entity.PaginationInput) (entity.CommitConnection, error) {
	filter := bson.M{"_id.r": repoID}
	if filters.Branch != nil {
		filter["branches"] = filters.Branch
	}

	//TODO: USE STRUCT TO CLEANUP
	if filters.MinRisk != nil || filters.MaxRisk != nil {
		f := make(bson.M)
		if filters.MinRisk != nil {
			f["$gte"] = filters.MinRisk
		}

		if filters.MaxRisk != nil {
			f["$lte"] = filters.MaxRisk
		}

		filter["analysis.bug_potential"] = f
	}

	if filters.DeveloperName != nil {
		filter["author.name"] = filters.DeveloperName
	}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: descCommitIndex,
		AscIndex:  ascCommitIndex,
	}

	return newCommitConnection(db.collection, filter, opts, &orderCfg, commitCursorParser(repoID)), nil
}

type orderDirectionConfig struct {
	Direction entity.OrderDirection
	DescIndex bson.D
	AscIndex  bson.D
}

func HashSetToFullIDs(repoID identifier.RepositoryID, set identifier.HashSet) []*identifier.CommitID {
	hashes := make([]*identifier.CommitID, 0, len(set))
	for h := range set {
		hashes = append(hashes, identifier.NewCommitID(repoID, h))
	}
	return hashes
}

func HashSliceToFullIDs(repoID identifier.RepositoryID, s []identifier.Hash) []*identifier.CommitID {
	hashes := make([]*identifier.CommitID, 0, len(s))
	for i := range s {
		hashes = append(hashes, identifier.NewCommitID(repoID, s[i]))
	}
	return hashes
}

func (db *commitDataSource) MarkBuggy(ctx context.Context, repoID identifier.RepositoryID, fix identifier.Hash, bugs identifier.HashSet) error {
	h := fix.Hex()
	_, err := db.collection.UpdateMany(ctx, bson.M{
		"_id":   bson.M{"$in": HashSetToFullIDs(repoID, bugs)},
		"fixes": bson.M{"$ne": h},
	}, bson.M{
		"$push": bson.M{"fixes": h},
	})

	return err
}

func (db *commitDataSource) RepositoryEngineFiles(ctx context.Context, id identifier.RepositoryID) (entity.EngineFileIter, error) {

	pipe := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"_id.r": id,
		}}},
		{{Key: "$project", Value: bson.M{
			"files": bson.M{
				"$filter": bson.M{
					"input": "$files",
					"cond":  bson.M{"$eq": bson.A{"$$this.type", classify.FileCode}},
				},
			},
		}}},
	}

	cur, err := db.collection.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	return newEngineFileIter(cur), nil
}

func (db *commitDataSource) FileAggregatedMetrics(ctx context.Context, id identifier.RepositoryID) (entity.FileMeasuresIter, error) {

	pipe := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"_id.r": id,
		}}},
		{{Key: "$unwind", Value: "$files"}},
		{{Key: "$match", Value: bson.M{
			"files.type": classify.FileCode,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":  "$files.path",
			"la":   bson.M{"$avg": "$files.metrics.la"},
			"ld":   bson.M{"$avg": "$files.metrics.ld"},
			"ha":   bson.M{"$avg": "$files.metrics.ha"},
			"hd":   bson.M{"$avg": "$files.metrics.hd"},
			"lt":   bson.M{"$avg": "$files.metrics.lt"},
			"age":  bson.M{"$avg": "$files.metrics.age"},
			"ndev": bson.M{"$max": "$files.metrics.ndev"},
			"nuc":  bson.M{"$max": "$files.metrics.nuc"},
			"nfc":  bson.M{"$max": "$files.metrics.nfc"},
			"exp":  bson.M{"$avg": "$files.metrics.exp"},
			"rexp": bson.M{"$avg": "$files.metrics.rexp"},
		}}},
	}

	cur, err := db.collection.Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}

	return newMetricsFileMeasuresIter(cur), nil
}

func (db *commitDataSource) SaveCommitAnalysis(ctx context.Context, analyses ...*entity.CommitAnalysisHolder) error {
	if len(analyses) == 0 {
		// nothing to save
		return nil
	}

	models := make([]mongo.WriteModel, len(analyses))

	for i, a := range analyses {
		fieldValues := make(bson.M, len(a.FileInsights))
		fieldValues["analysis"] = a.Analysis
		for ii, f := range a.FileInsights {
			fieldValues[fmt.Sprintf("files.%d.insights", ii)] = f
		}

		models[i] = mongo.NewUpdateManyModel().SetFilter(bson.M{
			"_id": a.ID,
		}).SetUpdate(bson.M{
			"$set": fieldValues,
		})
	}

	_, err := db.collection.BulkWrite(ctx, models)

	return err
}

func (db *commitDataSource) FindRepoCommits(ctx context.Context, repoID identifier.RepositoryID, opts ...*options.FindOptions) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
	}, opts...)
}

func (db *commitDataSource) FindPullRequestCommits(ctx context.Context, repoID identifier.RepositoryID, pull identifier.PullRequestID, opts ...*options.FindOptions) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
		"pulls": pull,
	}, opts...)
}

func (db *commitDataSource) FindByID(ctx context.Context, commitID *identifier.CommitID) (*entity.Commit, error) {
	var c entity.Commit
	err := db.collection.FindOne(ctx, bson.M{
		"_id": commitID,
	}).Decode(&c)
	return &c, err
}

func (db *commitDataSource) ContributorsCount(ctx context.Context, repoID identifier.RepositoryID) (int, error) {
	val, err := db.collection.Distinct(ctx, "author.email", bson.M{
		"_id.r": repoID,
	})

	return len(val), err
}

func (db *commitDataSource) BugInducingCount(ctx context.Context, repoID identifier.RepositoryID) (int64, error) {
	return db.collection.CountDocuments(ctx, bson.M{
		"_id.r":   repoID,
		"fixes.0": bson.M{"$exists": true},
	})
}

func (db *commitDataSource) BugFixingCount(ctx context.Context, repoID identifier.RepositoryID) (int, error) {
	res, err := db.collection.Distinct(ctx, "fixes", bson.M{
		"_id.r": repoID,
		"fixes": bson.M{"$exists": true},
	})

	return len(res), err
}

func (db *commitDataSource) FindJobCommits(ctx context.Context, repoID identifier.RepositoryID, job identifier.JobID) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
		"job":   job,
	})
}

func (db *commitDataSource) FindCommitsBetween(ctx context.Context, repoID identifier.RepositoryID, start identifier.JobID, end identifier.JobID) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
		"job": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
}

func (db *commitDataSource) FindCommitsByHash(ctx context.Context, repoID identifier.RepositoryID, hs ...identifier.Hash) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
		"_id.h": bson.M{
			"$in": hs,
		},
	})
}

func (db *commitDataSource) FindCommitsUntil(ctx context.Context, repoID identifier.RepositoryID, job identifier.JobID) (entity.CommitIter, error) {
	return db.find(ctx, bson.M{
		"_id.r": repoID,
		"job":   bson.M{"$lte": job},
	})
}

func (db *commitDataSource) DeleteRepoCommits(ctx context.Context, repoID identifier.RepositoryID) error {
	_, err := db.collection.DeleteMany(ctx, bson.M{
		"_id.r": repoID,
	})

	return err
}

var onlyIDs = options.Find().SetProjection(bson.M{"_id": 1})

type strCommitIDs []struct {
	Id identifier.CommitID `bson:"_id"`
}

func (ids strCommitIDs) ToHashes() []string {
	r := make([]string, len(ids))
	for i := range ids {
		r[i] = ids[i].Id.CommitHash.Hex()
	}

	return r
}

func (db *commitDataSource) ReTagBranch(ctx context.Context, repoID identifier.RepositoryID, branch string, commitHashes identifier.HashSet) error {
	_, err := db.collection.BulkWrite(ctx, []mongo.WriteModel{
		mongo.NewUpdateManyModel().SetFilter(bson.M{
			"_id.r": repoID,
		}).SetUpdate(bson.M{
			"$pull": bson.M{
				"branches": branch,
			},
		}),
		mongo.NewUpdateManyModel().SetFilter(bson.M{
			"_id": bson.M{
				"$in": HashSetToFullIDs(repoID, commitHashes),
			},
		}).SetUpdate(bson.M{
			"$push": bson.M{
				"branches": branch,
			},
		}),
	})

	return err
}

func (db *commitDataSource) ReTagPullRequest(ctx context.Context, repoID identifier.RepositoryID, pull identifier.PullRequestID, commitIDs identifier.HashSet) error {
	h := HashSetToFullIDs(repoID, commitIDs)
	_, err := db.collection.BulkWrite(ctx, []mongo.WriteModel{
		mongo.NewUpdateManyModel().SetFilter(bson.M{
			"$and": bson.A{
				bson.M{"_id.r": repoID},
				bson.M{"_id": bson.M{
					"$nin": h,
				}},
			},
			"pulls": pull,
		}).SetUpdate(bson.M{
			"$pull": bson.M{
				"pulls": pull,
			},
		}),
		mongo.NewUpdateManyModel().SetFilter(bson.M{
			"_id": bson.M{
				"$in": h,
			},
			"pulls": bson.M{"$ne": pull},
		}).SetUpdate(bson.M{
			"$push": bson.M{
				"pulls": pull,
			},
		}),
	})
	return err
}

func (db *commitDataSource) RemoveBranch(ctx context.Context, repoID identifier.RepositoryID, branch string) error {
	_, err := db.collection.UpdateMany(ctx, bson.M{
		"_id.r": repoID,
	}, bson.M{
		"$pull": bson.M{
			"branches": branch,
		},
	})
	return err
}

func (db *commitDataSource) Prune(ctx context.Context, repoID identifier.RepositoryID) error {
	_, err := db.collection.DeleteMany(ctx, bson.M{
		"_id.r":      repoID,
		"pulls.0":    bson.M{"$exists": false},
		"branches.0": bson.M{"$exists": false},
	})
	return err
}

func (db *commitDataSource) DeleteCommitTag(ctx context.Context, commitID *identifier.CommitID, tag classify.Tag) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{
		"_id": commitID,
	}, bson.M{
		"$pull": bson.M{
			"tags": tag,
		},
		"$addToSet": bson.M{
			"dtags": tag,
		},
	})

	return err
}

func (db *commitDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.CommitIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return newEntityCommitIter(cur), nil
}

const errorDuplicateKey = 11000

func isDuplicateKeyErr(err error) bool {
	switch err := err.(type) {
	case mongo.BulkWriteException:
		for i := range err.WriteErrors {
			if err.WriteErrors[i].Code == errorDuplicateKey {
				return true
			}
		}
	}

	return false
}

type engineFileIter struct {
	cur *mongo.Cursor
}

func newEngineFileIter(cur *mongo.Cursor) *engineFileIter {
	return &engineFileIter{
		cur: cur,
	}
}

func (iter *engineFileIter) ForEach(ctx context.Context, cb func(identifier.Hash, map[string]*engine.FileInfo) error) error {
	defer iter.cur.Close(ctx)

	for iter.cur.Next(ctx) {
		var doc struct {
			CommitID identifier.CommitID `bson:"_id"`
			Files    []*engine.FileInfo  `bson:"files"`
		}

		if err := iter.cur.Decode(&doc); err != nil {
			return err
		}

		files := make(map[string]*engine.FileInfo, len(doc.Files))
		for _, f := range doc.Files {
			files[f.Path] = f
		}

		return cb(doc.CommitID.CommitHash, files)
	}

	return iter.cur.Err()
}

func getStructTag(t interface{}) bson.D {
	val := reflect.TypeOf(engine.FileInfo{})
	count := val.NumField()

	res := make(bson.D, count)
	for i := 0; i < count; i++ {
		res[i] = bson.E{Key: val.Field(i).Tag.Get("bson"), Value: 1}
	}

	return res
}
