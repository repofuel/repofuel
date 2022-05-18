package mongosrc

import (
	"context"
	"errors"
	"time"

	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
)

type userDataSource struct {
	collection *mongo.Collection
}

func NewUserDataSource(db *mongo.Database) *userDataSource {
	//todo: make sure that the registry configured to encrypt the values
	return &userDataSource{
		collection: db.Collection("users"),
	}
}

func (db *userDataSource) Update(ctx context.Context, u *entity.User) error {
	u.UpdatedAt = time.Now()

	_, err := db.collection.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": u})
	return err
}

func (db *userDataSource) ProviderToken(ctx context.Context, id permission.UserID, provider string) (*oauth2.Token, error) {
	providerTokenOpts := options.FindOne().SetProjection(bson.M{
		"_id":            0,
		"providers.cred": 1,
		"providers": bson.M{
			"$elemMatch": bson.M{"provider": provider},
		},
	})

	var usr entity.User
	err := db.collection.FindOne(ctx, bson.M{
		"_id": id,
	}, providerTokenOpts).Decode(&usr)
	if err != nil {
		return nil, err
	}

	if len(usr.Providers) != 1 {
		return nil, errors.New("user do not have the provider")
	}

	cred, ok := usr.Providers[0].Cred.(credentials.String)
	if !ok {
		return nil, errors.New("unexpected credential type")
	}

	return &oauth2.Token{
		AccessToken: string(cred),
		TokenType:   "Bearer",
	}, nil
}

var findOrCreateByProviderOpts = options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

func (db *userDataSource) FindAndModifyProvider(ctx context.Context, pu *common.User) (*entity.User, error) {
	user := entity.NewUser(pu.Username, pu.FullName, pu.AvatarURL)

	res := db.collection.FindOneAndUpdate(ctx, bson.M{
		"providers": bson.M{
			"$elemMatch": bson.M{
				"provider": pu.Provider,
				"id":       pu.ID,
			},
		},
	}, bson.M{
		"$pull": bson.M{
			"providers": bson.M{
				"provider": pu.Provider,
				"id":       pu.ID,
			},
		},
		"$setOnInsert": user,
	}, findOrCreateByProviderOpts)

	err := res.Decode(user)
	if err != nil {
		return nil, err
	}

	//fixme: this is a workaround because I was not able to update the provider array in one automic query
	res = db.collection.FindOneAndUpdate(ctx, bson.M{
		"_id": user.Id,
		"providers": bson.M{
			"$not": bson.M{
				"$elemMatch": bson.M{
					"provider": pu.Provider,
					"id":       pu.ID,
				},
			},
		},
	}, bson.M{
		"$set":  bson.M{"updated_at": time.Now()},
		"$push": bson.M{"providers": pu},
	})

	err = res.Decode(user)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	return user, nil
}

func (db *userDataSource) Insert(ctx context.Context, u *entity.User) error {
	_, err := db.collection.InsertOne(ctx, u)
	return err
}

func (db *userDataSource) Find(ctx context.Context, id permission.UserID) (*entity.User, error) {
	usr := &entity.User{}
	err := db.collection.FindOne(ctx, bson.M{"_id": id}).Decode(usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
