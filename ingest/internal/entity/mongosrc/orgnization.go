package mongosrc

import (
	"context"
	"crypto/cipher"
	"fmt"
	"reflect"
	"time"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/codec"
	"github.com/repofuel/repofuel/pkg/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type organizationDataSource struct {
	gcm        cipher.AEAD
	collection *mongo.Collection
}

func NewOrganizationDataSource(db *mongo.Database, gcm cipher.AEAD) *organizationDataSource {
	interfaceCodec := codec.NewInterfaceCodec("ConfigType",
		&entity.InstallationConfig{},
		&entity.BitbucketOAuth1Config{},
		&entity.JiraOAuth1Config{},
		&entity.JiraBasicAuthConfig{},
	)
	rb := codec.NewRegistryWithEncryption(gcm)
	interfaceCodec.RegisterInterfaceCodec(rb, reflect.TypeOf((*entity.IntegrationConfig)(nil)).Elem())

	return &organizationDataSource{collection: db.Collection("organizations", &options.CollectionOptions{
		Registry: rb.Build(),
	})}
}

var orgFindOrCreateOpts = options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

func (db *organizationDataSource) FindOrCreate(ctx context.Context, org *entity.Organization) (*entity.Organization, error) {
	now := time.Now()

	var doc entity.Organization
	err := db.collection.FindOneAndUpdate(ctx, bson.M{
		"provider_scm": org.ProviderSCM,
		"owner.id":     org.Owner.ID,
	}, bson.M{
		"$set": &entity.Organization{
			ProvidersConfig: org.ProvidersConfig,
			UpdatedAt:       now,
		},
		"$setOnInsert": &entity.Organization{
			Owner:       org.Owner,
			ProviderSCM: org.ProviderSCM,
			ProviderITS: org.ProviderITS,
			AvatarURL:   org.AvatarURL,
			Members:     org.Members,
			CreatedAt:   now,
		},
	}, orgFindOrCreateOpts).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, err
}

func (db *organizationDataSource) UpdateOwner(ctx context.Context, id identifier.OrganizationID, owner *common.Account) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"owner": owner,
		},
	})

	return err
}

func (db *organizationDataSource) UpdateMembers(ctx context.Context, id identifier.OrganizationID, members map[string]common.Membership) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{
		"$set": bson.M{
			"members": members,
		},
	})

	return err
}

func (db *organizationDataSource) SetProviderConfig(ctx context.Context, id identifier.OrganizationID, provider string, cfg entity.IntegrationConfig) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.M{"$set": map[string]entity.IntegrationConfig{
		fmt.Sprintf("config.%s", provider): cfg,
	}})

	return err
}

func (db *organizationDataSource) FindByID(ctx context.Context, id identifier.OrganizationID) (*entity.Organization, error) {
	return db.findOne(ctx, bson.M{"_id": id})
}

func (db *organizationDataSource) FindBySlug(ctx context.Context, provider, slug string) (*entity.Organization, error) {
	return db.findOne(ctx, bson.M{
		"provider_scm": provider,
		"owner.slug":   slug,
	})
}

func (db *organizationDataSource) ListUserOrganizations(ctx context.Context, providers map[string]string) (entity.OrganizationIter, error) {
	var queries = make(bson.A, len(providers))
	var i int
	for p, userId := range providers {
		queries[i] = bson.M{
			"provider_scm":                    p,
			fmt.Sprintf("members.%s", userId): bson.M{"$exists": true},
		}
		i++
	}

	return db.find(ctx, bson.M{"$or": queries})
}

func (db *organizationDataSource) TotalCount(ctx context.Context, since time.Time) (int64, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	return db.collection.CountDocuments(ctx, filter)
}

func (db *organizationDataSource) CountOverTime(ctx context.Context, since time.Time, frequency entity.Frequency) ([]*entity.CountOverTime, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}

	cur, err := overTime(ctx, db.collection, filter, CountSummery, "$created_at", frequency)
	if err != nil {
		return nil, err
	}

	var res []*entity.CountOverTime
	err = cur.All(ctx, &res)
	return res, err
}

func (db *organizationDataSource) FindAllOrgsConnection(ctx context.Context, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.OrganizationConnection, error) {
	filter := bson.M{}

	return db.authenticatedOrganizationConnection(ctx, filter, direction, opts)
}

func (db *organizationDataSource) authenticatedOrganizationConnection(ctx context.Context, filter bson.M, direction *entity.OrderDirection, opts *entity.PaginationInput) (entity.OrganizationConnection, error) {
	err := attachOrganizationAuthorizationFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	orderCfg := orderDirectionConfig{
		Direction: getOrderDirection(direction, entity.OrderDirectionDesc),
		DescIndex: defaultDescIndex,
		AscIndex:  defaultAscIndex,
	}

	return newOrganizationConnection(db.collection, filter, opts, &orderCfg, defaultCursorParser), nil
}

//todo: it is important to write tests for this
func attachOrganizationAuthorizationFilter(ctx context.Context, filter bson.M) error {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return entity.ErrMissedViewerAccessInfo
	}

	if viewer.Role == permission.RoleSiteAdmin {
		return nil
	}

	providers := viewer.UserInfo.Providers

	var queries = make(bson.A, len(providers))
	var i int
	for p, userID := range providers {
		queries[i] = bson.M{
			"provider_scm":                    p,
			fmt.Sprintf("members.%s", userID): bson.M{"$exists": true},
		}
		i++
	}

	existingAnd, ok := filter["$and"].(bson.A)
	if !ok {
		existingAnd = make(bson.A, 0, 1)
	}

	filter["$and"] = append(existingAnd, bson.M{"$or": queries})

	return nil
}

func (db *organizationDataSource) All(ctx context.Context) (entity.OrganizationIter, error) {
	return db.find(ctx, bson.D{})
}

func (db *organizationDataSource) Delete(ctx context.Context, id identifier.OrganizationID) error {
	_, err := db.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (db *organizationDataSource) DeleteByOwnerID(ctx context.Context, provider, ownerID string) error {
	_, err := db.collection.DeleteOne(ctx, bson.M{
		"provider_scm": provider,
		"owner.id":     ownerID})
	return err
}

func (db *organizationDataSource) DeleteProviderConfig(ctx context.Context, orgID identifier.OrganizationID, providerID string) error {
	_, err := db.collection.UpdateOne(ctx, bson.M{"_id": orgID}, bson.M{"$unset": bson.M{"config." + providerID: ""}})
	return err
}

func (db *organizationDataSource) findOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*entity.Organization, error) {
	var doc entity.Organization
	err := db.collection.FindOne(ctx, filter, opts...).Decode(&doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (db *organizationDataSource) find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (entity.OrganizationIter, error) {
	cur, err := db.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return newEntityOrganizationIter(cur), nil
}
