// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package mongocon

import (
	"context"
	"crypto/cipher"
	"time"

	"github.com/repofuel/repofuel/pkg/codec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DatabaseOptions struct {
	URI  string `yaml:"uri"`
	Name string `yaml:"name"`
}

func NewDatabase(ctx context.Context, gcm cipher.AEAD, opts *DatabaseOptions) (*mongo.Database, error) {
	clientOpts := options.Client().
		ApplyURI(opts.URI).
		SetMaxPoolSize(10)

	if gcm != nil {
		clientOpts.SetRegistry(codec.NewRegistryWithEncryption(gcm).Build())
	}

	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client.Database(opts.Name), nil
}
