package entity

import (
	"context"

	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=connection_generic.go -out=connection_gnerated.go                   gen "Item=Commit,PullRequest,Repository,Job,Feedback,Organization"
//go:generate genny -in=mongosrc/connection_generic.go -out=mongosrc/connection_gnerated.go gen "Item=Commit,PullRequest,Repository,Job,Feedback,Organization"

type Item generic.Type

type ItemConnection interface {
	TotalCount(context.Context) (int64, error)
	Edges(context.Context) ([]*ItemEdge, error)
	PageInfo(context.Context) (*PageInfo, error)
	Nodes(context.Context) ([]*Item, error)
}

type ItemEdge struct {
	Node Item
}

func (e *ItemEdge) Cursor() *string {
	return nodeToItemCursor(&e.Node)
}

func PageInfoFromItemEdges(edges []*ItemEdge, hasNext bool, opts *PaginationInput) *PageInfo {
	if len(edges) == 0 {
		return &PageInfo{
			HasNextPage:     opts.Last != nil && opts.Before != nil,
			HasPreviousPage: opts.First != nil && opts.After != nil,
		}
	}

	var page PageInfo

	if opts.Last != nil {
		page.HasPreviousPage = len(edges) == *opts.Last && hasNext
		page.HasNextPage = opts.Before != nil
	} else {
		page.HasPreviousPage = opts.After != nil
		page.HasNextPage = len(edges) == *opts.First && hasNext
	}

	if len(edges) > 0 {
		page.StartEdge = edges[0]
		page.EndEdge = edges[len(edges)-1]
	}

	return &page
}
