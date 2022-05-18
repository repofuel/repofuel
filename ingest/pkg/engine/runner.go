// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package engine

import (
	"context"
	"errors"
	"fmt"
)

type Analyzer interface {
	AnalyzeCommit(context.Context, Commit) error
	Finish(context.Context) error
}

func RunForwardAnalysis(ctx context.Context, analyzer Analyzer, roots []Commit) (err error) {
	stack := NewCommitStack(roots...)
	seen := NewCommitSet()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("analysis panic: %+v", r)
		}
	}()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) {
			continue
		}

		err = analyzer.AnalyzeCommit(ctx, c)
		if err != nil {
			return err
		}

		seen.Add(c)
		stack.Push(c.Children()...)
	}
	return analyzer.Finish(ctx)
}

func LastFileChange(ctx context.Context, c Commit, path string) (Commit, error) {
	seen := NewCommitSet()         // only for the commits that have ore than one child
	commitsStack := CommitsStack{} //todo: maybe a queue will find last change faster
	var lastChange Commit

	commitsStack.Push(c.Parents()...)

	for !commitsStack.IsEmpty() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		c = commitsStack.Pop()

		if len(c.Children()) > 1 {
			if seen.Has(c) {
				continue
			}
			seen.Add(c)
		}

		if c.HasFile(path) {
			if lastChange == nil || lastChange.AuthorDate().Before(c.AuthorDate()) {
				lastChange = c
			}
			continue
		}

		commitsStack.Push(c.Parents()...)
	}

	if lastChange == nil {
		return nil, errors.New("cannot find last change")
	}

	return lastChange, nil
}

type FileInfo struct {
	Path      string    `bson:"path"`
	OldPath   string    `bson:"old_path,omitempty"`
	Subsystem string    `bson:"subsystem"`
	Fix       bool      `bson:"fix"`
	Action    DeltaType `bson:"action"`
}
