package graph

import (
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/classify"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

func tagsToStrings(org []classify.Tag) []string {
	tags := make([]string, len(org))
	for i, t := range org {
		tags[i] = t.String()
	}
	return tags
}

func hashToBranches(org map[string]identifier.Hash) []*entity.Branch {
	branches := make([]*entity.Branch, 0, len(org))
	for name, h := range org {
		b := &entity.Branch{
			Name: name,
			SHA:  h.Hex(),
		}
		branches = append(branches, b)
	}
	return branches
}
