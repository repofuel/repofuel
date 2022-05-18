package mongosrc

import (
	"context"

	"github.com/cheekybits/genny/generic"
	"go.mongodb.org/mongo-driver/mongo"
)


type item generic.Type

type itemIter struct {
	cur *mongo.Cursor
}

func newitemIter(cur *mongo.Cursor) *itemIter {
	return &itemIter{cur: cur}
}

func (iter *itemIter) ForEach(ctx context.Context, fun func(*item) error) error {
	defer iter.cur.Close(ctx)
	for iter.cur.Next(ctx) {
		var doc item
		if err := iter.cur.Decode(&doc); err != nil {
			return err
		}

		if err := fun(&doc); err != nil {
			return err
		}
	}
	return iter.cur.Err()
}

func (iter *itemIter) Slice(ctx context.Context) ([]*item, error) {
	s := make([]*item, 0, iter.cur.RemainingBatchLength())

	err := iter.ForEach(ctx, func(i *item) error {
		s = append(s, i)
		return nil
	})

	return s, err
}
