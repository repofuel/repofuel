// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package engine

import "github.com/repofuel/repofuel/ingest/pkg/identifier"

func (s CommitSet) HashesSet() identifier.HashSet {
	h := make(identifier.HashSet, len(s))
	for k := range s {
		h.Add(k.Hash())
	}
	return h
}

func (s CommitSet) HashesSlice() []identifier.Hash {
	h := make([]identifier.Hash, 0, len(s))
	for k := range s {
		h = append(h, k.Hash())
	}
	return h
}
