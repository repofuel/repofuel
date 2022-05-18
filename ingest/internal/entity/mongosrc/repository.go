// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package mongosrc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const repositoriesCollection = "repositories"

type repositoryDataSource struct {
	collection *mongo.Collection
	monitorDB  entity.MonitorDataSource
}

func (db *repositoryDataSource) TotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	return db.collection.CountDocuments(ctx, filter)
}

//todo: use this variable where the value is used to reduce duplication
var CountSummery = bson.E{Key: "count", Value: bson.M{"$sum": 1}}

func (db *repositoryDataSource) CountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, "$created_at", frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

func NewRepositoryDataSource(db *mongo.Database, monitorDB entity.MonitorDataSource) *repositoryDataSource {
	return &repositoryDataSource{
		collection: db.Collection(repositoriesCollection),
		monitorDB:  monitorDB,
	}
}

func (db *repositoryDataSource) FindWhereStatusNot(ctx context.Context, s ...status.Stage) (entity.RepositoryIter, error) {
	return db.find(ctx, bson.M{
		"status": bson.M{"$nin": s},
	})
}

func (db *repositoryDataSource) FindByOwnerID(ctx context.Context, platform string, ownerID string) (entity.RepositoryIter, error) {
	return db.find(ctx, bson.M{
		"provider_scm": platform,
		"owner.id":     ownerID,
	})
}

func (db *repositoryDataSource) FindByCollaborator(ctx context.Context, providers map[string]string) (entity.RepositoryIter, error) {
	var queries = make(bson.A, len(providers))
	var i int
	for p, userId := range providers {
		queries[i] = bson.D{
			{Key: "provider_scm", Value: p},
			{Key: fmt.Sprintf("collaborators.%s.read", userId), Value: true},
		}
		i++
	}

	return db.find(ctx, bson.D{{Key: "$or", Value: queries}})
}

func (db *repositoryDataSource) FindAllReposConnection(ctx context.Context, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.RepositoryConnection, error) {
	return db.authenticatedRepositoryConnection(ctx, bson.M{}, direction, opts)
}

func (db *repositoryDataSource) FindUserRepos(ctx context.Context, provider, owner string) (entity.RepositoryIter, error) {
	return db.find(ctx, bson.D{
		{Key: "provider_scm", Value: provider},
		{Key: "owner.slug", Value: owner},
	})

}

func (db *repositoryDataSource) FindUserReposByCollaborator(ctx context.Context, provider string, owner, collaboratorID string) (entity.RepositoryIter, error) {
	filter := bson.M{
		"provider_scm": provider,
		"owner.slug":   owner,
		fmt.Sprintf("collaborators.%s.read", collaboratorID): true,
	}

	return db.find(ctx, filter)
}

//todo: it is important to write tests for this
func attachRepositoryAuthorizationFilter(ctx context.Context, filter bson.M) error {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return entity.ErrMissedViewerAccessInfo
	}

	if viewer.Role == permission.RoleSiteAdmin {
		return nil
	}

	providers := viewer.UserInfo.Providers

	var queries = make(bson.A, len(providers)+1)
	var i int
	for p, userID := range providers {
		queries[i] = bson.M{
			"provider_scm": p,
			fmt.Sprintf("collaborators.%s.read", userID): true,
		}
		i++
	}
	queries[len(providers)] = publicReposFillter

	existingAnd, ok := filter["$and"].(bson.A)
	if !ok {
		existingAnd = make(bson.A, 0, 1)
	}

	filter["$and"] = append(existingAnd, bson.M{"$or": queries})

	return nil
}

func (db *repositoryDataSource) authenticatedRepositoryConnection(ctx context.Context, filter bson.M, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.RepositoryConnection, error) {
	err := attachRepositoryAuthorizationFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: defaultDescIndex,
		AscIndex:  defaultAscIndex,
	}

	return newRepositoryConnection(db.collection, filter, opts, &orderCfg, defaultCursorParser), nil
}

func toRepoCollaboratorsFilter(providers []*common.User, permission string) bson.M {
	if len(providers) == 1 { //this is just an optimization
		p := providers[0]
		return bson.M{
			"provider_scm": p.Provider,
			fmt.Sprintf("collaborators.%s.%s", p.ID, permission): true,
		}
	}

	var queries = make(bson.A, len(providers))
	for i, p := range providers {
		queries[i] = bson.M{
			"provider_scm": p.Provider,
			fmt.Sprintf("collaborators.%s.%s", p.ID, permission): true,
		}
	}
	return bson.M{"$or": queries}
}

func (db *repositoryDataSource) FindOrgReposConnection(ctx context.Context, orgID identifier.OrganizationID, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.RepositoryConnection, error) {
	filter := bson.M{
		"org_id": orgID,
	}

	return db.authenticatedRepositoryConnection(ctx, filter, direction, opts)
}

var publicReposFillter = bson.M{
	"source.private": bson.M{"$ne": true},
}

func (db *repositoryDataSource) FindUserReposConnection(ctx context.Context, inputs []*entity.UserAffiliationInput, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.RepositoryConnection, error) {

	var queries bson.A

	for _, input := range inputs {
		affiliations := input.Affiliations
		if len(affiliations) == 0 {
			affiliations = []entity.RepositoryAffiliation{
				entity.RepositoryAffiliationAccess,
				entity.RepositoryAffiliationMonitor,
			}
		}

		for _, affiliation := range affiliations {
			switch affiliation {
			case entity.RepositoryAffiliationAccess:
				queries = append(queries, toRepoCollaboratorsFilter(input.Providers, "read"))

			case entity.RepositoryAffiliationCollaborator:
				queries = append(queries, toRepoCollaboratorsFilter(input.Providers, "write"))

			case entity.RepositoryAffiliationOwner:
				queries = append(queries, toRepoCollaboratorsFilter(input.Providers, "admin"))

			case entity.RepositoryAffiliationMonitor:
				ids, err := db.monitorDB.UserReposIDs(ctx, input.UserID)
				if err != nil {
					return nil, err
				}

				queries = append(queries, bson.M{
					"_id": bson.M{"$in": ids},
				})

			default:
				return nil, errors.New("unsupported affiliation input")
			}
		}

	}

	var filter bson.M
	switch len(queries) {
	case 0:
		return nil, errors.New("missing user affiliation input")
	case 1:
		filter = queries[0].(bson.M)
	default:
		filter = bson.M{"$or": queries}
	}

	return db.authenticatedRepositoryConnection(ctx, filter, direction, opts)
}

func (db *repositoryDataSource) FindByID(ctx context.Context, id identifier.RepositoryID) (*entity.Repository, error) {
	return db.findOne(ctx, bson.M{"_id": id})
}

type NotFoundError struct {
	entity string
}

var statusByIdOptions = options.FindOne().SetProjection(bson.M{"status": 1})

func (db *repositoryDataSource) StatusByID(ctx context.Context, id identifier.RepositoryID) (status.Stage, error) {
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

func (db *repositoryDataSource) FindByName(ctx context.Context, provider string, owner string, repo string) (*entity.Repository, error) {
	return db.findOne(ctx, bson.M{
		"provider_scm": provider,
		"owner.slug":   owner,
		"source.name":  repo,
	})
}

func (db *repositoryDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Repository, error) {
	var repository entity.Repository
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&repository)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrRepositoryNotExist
		}
		return nil, err
	}

	return &repository, nil
}

func (db *repositoryDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.RepositoryIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return newEntityRepositoryIter(cur), nil
}

var insertOrUpdateOpts = options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

func (db *repositoryDataSource) InsertOrUpdate(ctx context.Context, r *entity.Repository) error {
	now := time.Now()

	filter := bson.M{
		"source.id":    r.Source.ID,
		"provider_scm": r.ProviderSCM,
	}

	update := bson.M{
		"$set": &entity.Repository{
			Source:        r.Source,
			Organization:  r.Organization,
			Owner:         r.Owner,
			Collaborators: r.Collaborators,
			UpdatedAt:     now,
		},
		"$setOnInsert": &entity.Repository{
			ProviderSCM: r.ProviderITS,
			ProviderITS: r.ProviderITS,
			MonitorMode: r.MonitorMode,
			DataVersion: entity.CurrentDataVersion,
			CreatedAt:   now,
		},
	}

	if !r.MonitorMode {
		update["$unset"] = bson.M{
			"monitor_mode": "",
		}
	}

	return db.collection.FindOneAndUpdate(ctx, filter, update, insertOrUpdateOpts).Decode(r)
}

func (db *repositoryDataSource) FindByProviderIDs(ctx context.Context, provider string, ids []string) (entity.RepositoryIter, error) {
	return db.find(ctx, bson.M{
		"provider_scm": provider,
		"source.id":    bson.M{"$in": ids},
	})
}

func (db *repositoryDataSource) FindByProviderID(ctx context.Context, provider, id string) (*entity.Repository, error) {
	return db.findOne(ctx, bson.M{
		"provider_scm": provider,
		"source.id":    id,
	})
}

func (db *repositoryDataSource) SaveStatus(ctx context.Context, id identifier.RepositoryID, status status.Stage) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}})
}

func (db *repositoryDataSource) SaveCommitsCount(ctx context.Context, id identifier.RepositoryID, count int) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"commits_count": count}})
}

func (db *repositoryDataSource) SaveBuggyCount(ctx context.Context, id identifier.RepositoryID, count int) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"buggy_count": count}})
}

var branchesOpts = options.FindOne().SetProjection(bson.M{"_id": 0, "branches": 1})

func (db *repositoryDataSource) Branches(ctx context.Context, id identifier.RepositoryID) (map[string]identifier.Hash, error) {
	r, err := db.findOne(ctx, bson.M{"_id": id}, branchesOpts)
	if err != nil {
		return nil, err
	}
	return r.Branches, nil
}

func (db *repositoryDataSource) UpgradeDataVersion(ctx context.Context, id identifier.RepositoryID, v uint32) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"v": v,
		}})
}

func (db *repositoryDataSource) SaveQuality(ctx context.Context, id identifier.RepositoryID, quality repofuel.PredictionStatus) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{
			"$set":   bson.M{"quality": quality},
			"$unset": bson.M{"confidence": ""},
		})
}

func (db *repositoryDataSource) UpdateOwner(ctx context.Context, id identifier.OrganizationID, owner *common.Account) error {
	_, err := db.collection.UpdateMany(ctx,
		bson.M{"org_id": id},
		bson.M{"$set": bson.M{
			"owner": owner,
		}})

	return err
}

func (db *repositoryDataSource) SaveConfidence(ctx context.Context, id identifier.RepositoryID, confidence float32) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{
			"$set":   bson.M{"confidence": confidence},
			"$unset": bson.M{"quality": ""},
		})
}

func (db *repositoryDataSource) SaveBranches(ctx context.Context, id identifier.RepositoryID, bs map[string]identifier.Hash) error {
	return db.updateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"branches": bs,
		}})
}

func (db *repositoryDataSource) updateOne(ctx context.Context, filter interface{}, update bson.M, opts ...*options.UpdateOptions) error {
	return updateOne(ctx, db.collection, filter, update, opts...)
}

func (db *repositoryDataSource) Delete(ctx context.Context, id identifier.RepositoryID) error {
	r, err := db.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if r.DeletedCount != 1 {
		return errors.New("cannot delete the repository")
	}
	return nil
}

//todo: outdated
func (db *repositoryDataSource) DeleteCollaborator(ctx context.Context, provider string, repo string, user string) error {
	return db.updateOne(ctx,
		bson.M{"provider_scm": provider, "source.id": repo},
		bson.M{"$unset": `collaborators.${user}`})
}

//todo: outdated
func (db *repositoryDataSource) AddCollaborator(ctx context.Context, provider string, repo string, user string, p common.Permissions) error {
	return db.updateOne(ctx,
		bson.M{"provider_scm": provider, "source.id": repo},
		bson.M{"$set": bson.M{`collaborators.${user}`: p}})
}

func (db *repositoryDataSource) UpdateSource(ctx context.Context, id identifier.RepositoryID, source *common.Repository) error {
	return db.updateOne(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"source": source,
		}})
}

func (db *repositoryDataSource) UpdateCollaborators(ctx context.Context, id identifier.RepositoryID, col map[string]common.Permissions) error {
	return db.updateOne(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"collaborators": col,
		}})
}

func (db *repositoryDataSource) FindAndUpdateChecksConfig(ctx context.Context, id identifier.RepositoryID, cfg *entity.ChecksConfig) (*entity.Repository, error) {
	var doc entity.Repository
	var opts = options.FindOneAndUpdate().SetReturnDocument(options.After)
	var filter = bson.M{
		"_id": id,
	}
	var update = bson.M{
		"$set": bson.M{
			"checks_config": cfg,
		},
	}

	err := db.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

type sharedAccountIter struct {
	cur *mongo.Cursor
}

func newSharedAccountIter(cur *mongo.Cursor) *sharedAccountIter {
	return &sharedAccountIter{
		cur: cur,
	}
}

func (iter *sharedAccountIter) ForEach(ctx context.Context, fun func(*entity.SharedAccount) error) error {
	defer iter.cur.Close(ctx)
	for iter.cur.Next(ctx) {
		var doc struct {
			Unique entity.SharedAccount `bson:"_id"`
		}
		if err := iter.cur.Decode(&doc); err != nil {
			return err
		}
		if err := fun(&doc.Unique); err != nil {
			return err
		}
	}
	return iter.cur.Err()
}
