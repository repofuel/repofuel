package brancher

import (
	"context"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
)

type Rearrange struct {
	commitsDB       entity.CommitDataSource
	reposDB         entity.RepositoryDataSource
	repo            *engine.Repository
	affected        engine.StringSet
	newCommits      engine.CommitSet
	branches        engine.Branches
	analyzeBranches engine.Branches
}

func NewRearrange(commitsDB entity.CommitDataSource, reposDB entity.RepositoryDataSource, repo *engine.Repository, b engine.Branches, analyzeBranches engine.Branches, startPoints []engine.Commit) *Rearrange {
	affected := engine.NewStringSet()
	for name, head := range b {
		base, ok := analyzeBranches[name]
		if !ok || !repo.IsAncestor(base, head) {
			affected.Add(name)
		}
	}

	return &Rearrange{
		commitsDB:       commitsDB,
		reposDB:         reposDB,
		repo:            repo,
		affected:        affected,
		newCommits:      engine.SuccessorList(startPoints...),
		branches:        b,
		analyzeBranches: analyzeBranches,
	}
}

func (r *Rearrange) AnalyzeCommit(ctx context.Context, c engine.Commit) error {
	if !c.IsMerge() {
		return nil
	}

	ancestor, err := engine.FindCommonAncestor(c.Parents())
	if err != nil {
		return nil
	}

	// ancestor's commits always includes all of their successors branches, if it the
	// same counts, it means it is the same branches. If the ancestor is not analyzed
	// yet, it means it will not affect the commits that are stored in the database.
	if c.NumBranches() == ancestor.NumBranches() || r.newCommits.Has(ancestor) {
		return nil
	}

	ancestorBranches := ancestor.Branches()
	for b := range c.Branches() {
		if ancestorBranches.Has(b) {
			continue
		}

		r.affected.Add(b)
	}

	return nil
}

func (r *Rearrange) Finish(ctx context.Context) error {
	// update commit branches (removed branches)
	deletedBranches := engine.DeletedBranches(r.analyzeBranches, r.branches)
	for _, name := range deletedBranches {
		err := r.commitsDB.RemoveBranch(ctx, r.repo.ID, name)
		if err != nil {
			return err
		}
	}

	if len(deletedBranches) > 0 {
		err := r.commitsDB.Prune(ctx, r.repo.ID)
		if err != nil {
			return err
		}
	}

	addedBranches := engine.AddedBranches(r.analyzeBranches, r.branches)
	for _, name := range addedBranches {
		r.affected.Add(name)
	}

	if r.newCommits.Count() == 0 && (len(deletedBranches) > 0 || len(addedBranches) > 0) {
		err := r.reposDB.SaveBranches(ctx, r.repo.ID, r.branches)
		if err != nil {
			return err
		}
	}

	for branch := range r.affected {
		commits, err := r.repo.BranchCommits(r.branches[branch])
		if err != nil {
			return err
		}

		err = r.commitsDB.ReTagBranch(ctx, r.repo.ID, branch, commits.HashesSet())
		if err != nil {
			return nil
		}
	}

	return nil
}
