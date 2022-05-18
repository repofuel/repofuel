package git2go

import (
	"context"
	"fmt"
	"time"

	git "github.com/libgit2/git2go"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

var diffOptions git.DiffOptions
var diffFindOptions git.DiffFindOptions

func init() {
	var err error
	diffOptions, err = git.DefaultDiffOptions()
	if err != nil {
		panic(err)
	}
	diffOptions.Flags = git.DiffIgnoreWhitespace
	diffOptions.ContextLines = 0
	diffOptions.OldPrefix = ""
	diffOptions.NewPrefix = ""

	diffFindOptions, err = git.DefaultDiffFindOptions()
	if err != nil {
		panic(err)
	}
	diffFindOptions.Flags = git.DiffFindRenames
}

type commit struct {
	id         git.Oid
	repo       *git.Repository
	developer  engine.Developer
	authorDate time.Time
	files      map[string]*engine.FileInfo
	children   []engine.Commit
	parents    []engine.Commit
	branches   engine.StringSet
}

func (c *commit) Branches() engine.StringSet {
	return c.branches
}

func (c *commit) NumBranches() int {
	return len(c.branches)
}

func (c *commit) AddBranch(branch string) {
	if c.branches == nil {
		c.branches = engine.NewStringSet()
	}
	c.branches.Add(branch)
}

func (c *commit) HasFile(path string) bool {
	_, ok := c.files[path]
	return ok
}

func (c *commit) Files() map[string]*engine.FileInfo {
	return c.files
}

func (c *commit) SetFiles(files map[string]*engine.FileInfo) {
	if len(files) == 0 {
		return
	}

	c.files = files
}

func newCommit(repo *git.Repository, id git.Oid) engine.Commit {
	return &commit{
		id:   id,
		repo: repo,
	}
}

func (c *commit) SetAuthorDate(date time.Time) {
	c.authorDate = date
}

func (c *commit) AuthorDate() time.Time {
	return c.authorDate
}

func (c *commit) SetDeveloper(dev engine.Developer) {
	c.developer = dev
}

func (c *commit) Object() (engine.CommitObject, error) {
	obj, err := c.repo.LookupCommit(&c.id)
	if err != nil {
		return nil, translateGitError(err)
	}

	return &commitObject{
		Repository: c.repo,
		Commit:     obj,
	}, nil
}

func translateGitError(err error) error {
	gitErr, ok := err.(*git.GitError)

	if ok && gitErr.Class == git.ErrClassOdb && gitErr.Code == git.ErrNotFound {
		return engine.ErrObjectNotFound
	}

	return err
}

func (c *commitObject) NumParents() int {
	return int(c.Commit.ParentCount())
}

func (c *commit) Developer() engine.Developer {
	return c.developer
}

type commitObject struct {
	*git.Repository
	*git.Commit
}

func (c *commitObject) FirstParentHash() identifier.Hash {
	return identifier.Hash(*c.Commit.ParentId(0))
}

func (c *commitObject) Hash() identifier.Hash {
	return identifier.Hash(*c.Commit.Id())
}

func (c *commit) Hash() identifier.Hash {
	return identifier.Hash(c.id)
}

func (c *commit) String() string {
	return c.id.String()
}

func (c *commitObject) Message() string {
	return c.Commit.Message()
}

func (c *commitObject) ParentHashes() []identifier.Hash {
	n := c.Commit.ParentCount()
	s := make([]identifier.Hash, n)
	for i := uint(0); i < n; i += 1 {
		s[i] = identifier.Hash(*c.Commit.ParentId(i))
	}
	return s
}

func (c *commit) NumChildren() int {
	return len(c.children)
}

func (c *commit) IsMerge() bool {
	return len(c.parents) > 1
}

func (c *commitObject) CommitterDate() time.Time {
	return c.Commit.Committer().When
}
func (c *commitObject) Author() *engine.Signature {
	return (*engine.Signature)(c.Commit.Author())
}

func (c *commitObject) AuthorName() string {
	return c.Commit.Author().Name
}

func (c *commitObject) AuthorDate() time.Time {
	return c.Commit.Author().When
}

func (c *commitObject) AuthorEmail() string {
	return c.Commit.Author().Email
}

func (c *commitObject) Developer() engine.Developer {
	return engine.Developer(c.Commit.Author().Email)
}

func (c *commit) Children() []engine.Commit {
	return c.children
}

func (c *commit) Parents() []engine.Commit {
	return c.parents
}

func (c *commit) NumParents() int {
	return len(c.parents)
}

func (c *commit) AddChild(child engine.Commit) {
	c.children = append(c.children, child)
}

func (c *commit) AddParent(parent engine.Commit) {
	c.parents = append(c.parents, parent)
}

func (c *commit) SetParents(parents []engine.Commit) {
	c.parents = parents
}

func (c *commitObject) ChangedFiles(ctx context.Context) ([]engine.DiffDelta, error) {
	panic("not implemented")
}

func (c *commitObject) diff(ctx context.Context, cbFile git.DiffForEachFileCallback, detail git.DiffDetail) error {
	var oldTree *git.Tree
	if c.ParentCount() > 0 {
		var err error
		p := c.Commit.Parent(0)
		defer p.Free()

		oldTree, err = p.Tree()
		if err != nil {
			return err
		}
		defer oldTree.Free()
	}

	newTree, err := c.Commit.Tree()
	if err != nil {
		return err
	}
	defer newTree.Free()

	diff, err := c.Repository.DiffTreeToTree(oldTree, newTree, &diffOptions)
	if err != nil {
		return err
	}
	defer diff.Free()

	err = diff.FindSimilar(&diffFindOptions)
	if err != nil {
		return err
	}

	return diff.ForEach(cbFile, detail) //fixme: could return git.ErrInvalid
}

func (c *commitObject) DiffFiles(ctx context.Context, cbFile engine.FileDiffCB) error {
	return c.diff(ctx, func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, cbFile(newDiffDelta(c.Repository, &delta))
	}, git.DiffDetailFiles)
}

func (c *commitObject) DiffHunks(ctx context.Context, cbFile engine.HunkDiffCB) error {
	return c.diff(ctx, func(delta git.DiffDelta, _ float64) (git.DiffForEachHunkCallback, error) {
		file, err := cbFile(newDiffDelta(c.Repository, &delta))

		return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			return nil, file.AnalyzeHunk(newDiffHunk(&hunk))
		}, err
	}, git.DiffDetailHunks)
}

func (c *commitObject) DiffLines(ctx context.Context, cbFile engine.LineDiffCB) error {
	return c.diff(ctx, func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
		file, err := cbFile(newDiffDelta(c.Repository, &delta))

		return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			err := file.AnalyzeHunk(newDiffHunk(&hunk))

			return func(line git.DiffLine) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				return file.AnalyzeLine(newDiffLine(&line))
			}, err
		}, err
	}, git.DiffDetailLines)
}

func (c *commitObject) Free() {
	c.Commit.Free()
}

func (c *commit) FirstParent() engine.Commit {
	return c.parents[0]
}

func (c *commit) HasParent() bool {
	return len(c.parents) > 0
}

func (c *commit) parent(i int) (*commit, error) {
	p, ok := c.parents[i].(*commit)
	if !ok {
		return nil, fmt.Errorf("the commit interface hold an unexpected underlyng type")
	}

	return p, nil
}
