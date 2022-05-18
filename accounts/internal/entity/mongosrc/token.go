// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package mongosrc

import (
	"context"
	"time"

	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type tokenDataSource struct {
	collection *mongo.Collection
}

func NewTokenDataSource(db *mongo.Database) *tokenDataSource {
	return &tokenDataSource{
		collection: db.Collection("tokens"),
	}
}

func (db tokenDataSource) Find(ctx context.Context, id string) (*entity.Token, error) {
	t := &entity.Token{}
	err := db.collection.FindOne(ctx, bson.M{"_id": id}).Decode(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (db tokenDataSource) GenerateToken(ctx context.Context, userId permission.UserID) (*entity.Token, error) {
	now := time.Now()
	t := &entity.Token{
		Id:        entity.RandomToken(80),
		UserId:    userId,
		CreatedAt: now,
		ExpiredAt: now.Add(time.Hour * 24 * 60),
	}
	_, err := db.collection.InsertOne(ctx, t)
	return t, err
}
