package mongosrc

import (
	"context"
	"sync"

	"github.com/cheekybits/genny/generic"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Item generic.Type

type ItemConnection struct {
	collection   *mongo.Collection
	filter       bson.M
	pgInput      *entity.PaginationInput
	orderCfg     *orderDirectionConfig
	cursorParser FuncCursorParser

	edges   []*entity.ItemEdge
	hasNext bool
	once    sync.Once
}

func newItemConnection(collection *mongo.Collection, filter bson.M, pgInput *entity.PaginationInput, orderCfg *orderDirectionConfig, cursorParser FuncCursorParser) *ItemConnection {
	return &ItemConnection{
		collection:   collection,
		filter:       filter,
		pgInput:      pgInput,
		orderCfg:     orderCfg,
		cursorParser: cursorParser,
	}
}

func (c *ItemConnection) TotalCount(ctx context.Context) (int64, error) {
	return c.collection.CountDocuments(ctx, c.filter)
}

func (c *ItemConnection) Edges(ctx context.Context) ([]*entity.ItemEdge, error) {
	var err error

	c.once.Do(func() {
		c.edges, c.hasNext, err = findItemEdges(ctx, c.collection, c.filter, c.pgInput, c.orderCfg, c.cursorParser)
	})

	return c.edges, err
}

func (c *ItemConnection) PageInfo(ctx context.Context) (*entity.PageInfo, error) {
	edges, err := c.Edges(ctx)
	if err != nil {
		return nil, err
	}

	return entity.PageInfoFromItemEdges(edges, c.hasNext, c.pgInput), nil
}

func (c *ItemConnection) Nodes(ctx context.Context) ([]*entity.Item, error) {
	edges, err := c.Edges(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*entity.Item, len(edges))
	for i := range edges {
		nodes[i] = &edges[i].Node
	}
	return nodes, nil
}

func findItemEdges(ctx context.Context, collection *mongo.Collection, filter bson.M, pgInput *entity.PaginationInput, orderCfg *orderDirectionConfig, cursorParser FuncCursorParser) ([]*entity.ItemEdge, bool, error) {
	err := pgInput.Validate("Items", 100)
	if err != nil {
		return nil, false, err
	}

	mongoOpts := options.Find()

	filter = copyBsonM(filter) //fixme: should have a better solution
	err = applyPaginationOptions(mongoOpts, filter, pgInput, orderCfg, cursorParser)
	if err != nil {
		return nil, false, err
	}

	//todo: apply projection

	cur, err := collection.Find(ctx, filter, mongoOpts)
	if err != nil {
		if err, ok := err.(mongo.CommandError); ok && err.Code == 51175 {
			// no results
			return make([]*entity.ItemEdge, 0), false, nil
		}
		return nil, false, err
	}
	defer cur.Close(ctx)

	edges, err := getSortedItemEdges(ctx, cur, pgInput)
	if err != nil {
		return nil, false, err
	}

	return edges, cur.Next(ctx), err
}

func getSortedItemEdges(ctx context.Context, cur *mongo.Cursor, opts *entity.PaginationInput) ([]*entity.ItemEdge, error) {
	if opts.Last != nil {
		return backwardItemEdges(ctx, cur, opts)
	}

	return forwardItemEdges(ctx, cur, opts)
}

func forwardItemEdges(ctx context.Context, cur *mongo.Cursor, opts *entity.PaginationInput) ([]*entity.ItemEdge, error) {
	var limit = *opts.First
	var edges = make([]*entity.ItemEdge, limit)
	var index = 0

	for index < limit && cur.Next(ctx) {
		var edge entity.ItemEdge
		err := cur.Decode(&edge.Node)
		if err != nil {
			return nil, err
		}
		edges[index] = &edge
		index++
	}
	edges = edges[:index]

	return edges, nil
}

func backwardItemEdges(ctx context.Context, cur *mongo.Cursor, opts *entity.PaginationInput) ([]*entity.ItemEdge, error) {
	var limit = *opts.Last
	var edges = make([]*entity.ItemEdge, limit)
	var index = limit - 1

	for index >= 0 && cur.Next(ctx) {
		var c entity.ItemEdge
		err := cur.Decode(&c.Node)
		if err != nil {
			return nil, err
		}
		edges[index] = &c
		index--
	}
	edges = edges[index+1:]

	return edges, nil
}
