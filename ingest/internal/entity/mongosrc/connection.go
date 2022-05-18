package mongosrc

import (
	"encoding/base64"
	"errors"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FuncCursorParser func(*string) (interface{}, error)

func copyBsonM(oldM bson.M) bson.M {
	newM := make(bson.M, len(oldM))
	for k, v := range oldM {
		newM[k] = v
	}
	return newM
}

func applyPaginationOptions(f *options.FindOptions, filter bson.M, opts *entity.PaginationInput, orderCfg *orderDirectionConfig, cursorParser FuncCursorParser) (err error) {
	if orderCfg.Direction == entity.OrderDirectionDesc {
		f.Min, err = cursorParser(opts.Before)
		if err != nil {
			return err
		}

		f.Max, err = cursorParser(opts.After)
		if err != nil {
			return err
		}
	} else {
		f.Min, err = cursorParser(opts.After)
		if err != nil {
			return err
		}

		f.Max, err = cursorParser(opts.Before)
		if err != nil {
			return err
		}
	}

	if opts.First != nil {
		first := (int64)(*opts.First)
		f.SetLimit(first + 1)
	}

	if opts.Last != nil {
		last := (int64)(*opts.Last)
		f.SetLimit(last + 1)
	}

	if f.Min != nil || f.Max != nil {
		var nor bson.A
		if f.Min != nil {
			nor = append(nor, f.Min)
		}
		if f.Max != nil {
			nor = append(nor, f.Max)
		}
		filter["$nor"] = nor

		f.Hint = orderCfg.AscIndex
	}

	switch adjustOrderDirection(opts, orderCfg.Direction) {
	case entity.OrderDirectionAsc:
		f.SetSort(orderCfg.AscIndex)
	case entity.OrderDirectionDesc:
		f.SetSort(orderCfg.DescIndex)
	}

	return nil
}

func commitCursorParser(repoID identifier.RepositoryID) func(cursor *string) (interface{}, error) {
	return func(cursor *string) (interface{}, error) {
		if cursor == nil {
			return nil, nil
		}

		var c entity.CommitCursor
		err := c.UnmarshalBase64([]byte(*cursor))
		if err != nil {
			return nil, err
		}

		c.RepoID = repoID

		return c, nil
	}
}

func base64ToObjectID(dst *primitive.ObjectID, src []byte) error {
	if base64.StdEncoding.DecodedLen(len(src)) != len(dst) {
		return errors.New("unexpected base64 length")
	}

	_, err := base64.StdEncoding.Decode(dst[:], src)
	return err
}

func defaultCursorParser(cursor *string) (interface{}, error) {
	if cursor == nil {
		return nil, nil
	}

	var id primitive.ObjectID
	err := base64ToObjectID(&id, []byte(*cursor))
	if err != nil {
		return nil, err
	}

	return bson.M{"_id": id}, nil
}

func getOrderDirection(input *entity.OrderDirection, defaultDirection entity.OrderDirection) entity.OrderDirection {
	if input == nil {
		return defaultDirection
	}

	return *input
}

func adjustOrderDirection(opts *entity.PaginationInput, d entity.OrderDirection) entity.OrderDirection {
	if opts.First != nil {
		return d
	}

	//flip the direction in case of `last` is provided
	if d == entity.OrderDirectionDesc {
		return entity.OrderDirectionAsc
	}
	return entity.OrderDirectionDesc
}
