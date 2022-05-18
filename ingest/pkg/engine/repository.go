// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package engine

//go:generate mockgen -destination=$PWD/internal/mock/providers/providers_mock.go github.com/repofuel/repofuel/ingest/pkg/providers SourceIntegration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/providers"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/repofuel/repofuel/pkg/credentials"
)

const DefaultRemote = "origin"

var (
	ErrLocalRepoNotExist = errors.New("local repository path is not exist")
	ErrCommitNotIngested = errors.New("commit should be ingested")
	ErrObjectNotFound    = errors.New("object is not founded in the git repository")
	ErrBranchNotFound    = errors.New("branch is not founded in the git repository")
	ErrBranchNotIngested = errors.New("branch is not ingested")
)

type BasicAuthFunc func(context.Context) (*credentials.BasicAuth, error)

type RepositoryAdapter interface {
	Open(path string) error
	Clone(ctx context.Context, url string, path string, auth BasicAuthFunc) error
	Fetch(ctx context.Context, auth BasicAuthFunc, remote, url string, branches ...string) error
	Branches() (map[string]identifier.Hash, error)
	Commit(id identifier.Hash) (Commit, error)
	InducingCommits(ctx context.Context, id identifier.Hash, path string, chunks ...ChunkAddr) (identifier.HashSet, error)
}

type Commit interface {
	Object() (CommitObject, error)
	AddChild(Commit)
	SetParents([]Commit)
	AddParent(Commit)
	Hash() identifier.Hash
	SetDeveloper(Developer)
	Developer() Developer
	NumChildren() int
	Children() []Commit
	AuthorDate() time.Time
	SetAuthorDate(time.Time)
	Parents() []Commit
	NumParents() int
	FirstParent() Commit
	HasParent() bool
	IsMerge() bool
	Files() map[string]*FileInfo
	HasFile(path string) bool
	SetFiles(files map[string]*FileInfo)
	AddBranch(branch string)
	Branches() StringSet
	NumBranches() int
}

type CommitObject interface {
	Hash() identifier.Hash
	FirstParentHash() identifier.Hash
	Message() string
	NumParents() int
	ParentHashes() []identifier.Hash
	AuthorDate() time.Time
	Author() *Signature
	CommitterDate() time.Time
	AuthorEmail() string
	AuthorName() string
	//DiffFiles(ctx context.Context, cbFile FileDiffCB) error
	DiffHunks(ctx context.Context, cbFile HunkDiffCB) error
	//DiffLines(ctx context.Context, cbFile AnalyzeFileCB) error
	Free()
}

type DiffDelta interface {
	//OldFile() DiffFile
	//NewFile() DiffFile
	FromPath() string
	ToPath() string
	Action() DeltaType
	IsBinary() bool
	IsSymlink() bool
	OldContent() ([]byte, error)
	NewContent() ([]byte, error)
	//Patch() (FilePatch, error)
}

type DiffFile interface {
	Content() ([]byte, error)
	Filename() string
}

type DiffHunk interface {
	LinesAdded() int
	LinesDeleted() int
	AddressDeleted() ChunkAddr
}

type DiffLine interface {
}

type FilePatch interface {
	Chunks() []Chunk
	IsBinary() bool
}

type Chunk interface {
	Content() string
	Action() DeltaType
	Size() int
}

type RepositoryOpts struct {
	Adapter   RepositoryAdapter
	Source    providers.SourceIntegration
	Issues    providers.IssuesIntegration
	OriginURL string
}

type Repository struct {
	ID        identifier.RepositoryID
	adapter   RepositoryAdapter
	scm       providers.SourceIntegration
	its       providers.IssuesIntegration
	originURL string
	path      string
	roots     CommitSet
	commits   map[identifier.Hash]Commit
}

func (r *Repository) ITS() providers.IssuesIntegration {
	return r.its
}

func (r *Repository) SCM() providers.SourceIntegration {
	return r.scm
}

func (r *Repository) Branches() (map[string]identifier.Hash, error) {
	return r.adapter.Branches()
}

func (r *Repository) Commits(all ...identifier.Hash) CommitSet {
	set := NewCommitSet()

	for _, h := range all {
		c, ok := r.commits[h]
		if ok {
			set.Add(c)
		}
	}

	return set
}

func (r *Repository) Commit(h identifier.Hash) (Commit, bool) {
	c, ok := r.commits[h]
	return c, ok
}

func (r *Repository) HeadCommits(branches map[string]identifier.Hash) CommitSet {
	set := NewCommitSet()

	for _, h := range branches {
		c, ok := r.commits[h]
		if ok {
			set.Add(c)
		}
	}

	return set
}

func (r *Repository) BranchCommits(h identifier.Hash) (CommitSet, error) {
	c, ok := r.commits[h]
	if !ok {
		return nil, errors.New("branch head is not ingested")
	}

	return AncestorsList(c), nil
}

func (r *Repository) Adapter() RepositoryAdapter {
	return r.adapter
}

func NewRepository(id identifier.RepositoryID, path string, opts *RepositoryOpts) *Repository {
	return &Repository{
		ID:        id,
		adapter:   opts.Adapter,
		scm:       opts.Source,
		its:       opts.Issues,
		originURL: opts.OriginURL,
		path:      path,
		roots:     NewCommitSet(),
		commits:   make(map[identifier.Hash]Commit),
	}
}

func (r *Repository) InducingCommits(ctx context.Context, id identifier.Hash, path string, chunks ...ChunkAddr) (identifier.HashSet, error) {
	return r.adapter.InducingCommits(ctx, id, path, chunks...)
}

func (r *Repository) Open() error {
	return r.adapter.Open(r.path)
}

func (r *Repository) Clone(ctx context.Context) error {
	return r.adapter.Clone(ctx, r.originURL, r.path, r.scm.BasicAuth)
}

func (r *Repository) IssuesFromText(ctx context.Context, s string) ([]common.Issue, bool, error) {
	return r.its.IssuesFromText(ctx, s)
}

func (r *Repository) FetchOrigin(ctx context.Context, branches ...string) error {
	return r.adapter.Fetch(ctx, r.scm.BasicAuth, DefaultRemote, r.originURL, branches...)
}

func (r *Repository) Fetch(ctx context.Context, remote, url string, branches ...string) error {
	return r.adapter.Fetch(ctx, r.scm.BasicAuth, remote, url, branches...)
}

type Branches map[string]identifier.Hash

// deprecated
func StartCommits(ctx context.Context, bases, heads CommitSet) (CommitSet, error) {
	seen := AncestorsList(bases.Slice()...)

	roots := NewCommitSet()
	for c := range heads {
		unseenRoots, err := UnseenRoots(ctx, seen, c)
		if err != nil {
			return nil, err
		}
		roots.Update(unseenRoots)
	}

	return roots, nil
}

//deprecated
func StartCommitsFromSeen(ctx context.Context, seen, heads CommitSet) (CommitSet, error) {
	roots := NewCommitSet()
	for c := range heads {
		unseenRoots, err := UnseenRoots(ctx, seen, c)
		if err != nil {
			return nil, err
		}
		roots.Update(unseenRoots)
	}

	return roots, nil
}

func DeletedBranches(analyzed, current Branches) []string {
	var deleted []string
	for name := range analyzed {
		if _, ok := current[name]; !ok {
			deleted = append(deleted, name)
		}
	}
	return deleted
}

func (r *Repository) TagBranchesOnCommits(b Branches, until CommitSet) error {
	for name, head := range b {
		c, ok := r.commits[head]
		if !ok {
			return ErrCommitNotIngested
		}

		for c := range AncestorsListUntil(until, c) {
			c.AddBranch(name)
		}
	}
	return nil
}

func AddedBranches(analyzed, current Branches) []string {
	var added []string
	for name := range current {
		if _, ok := analyzed[name]; !ok {
			added = append(added, name)
		}
	}
	return added
}

func AncestorsList(heads ...Commit) CommitSet {
	stack := NewCommitStack(heads...)
	seen := NewCommitSet()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) {
			continue
		}

		seen.Add(c)
		stack.Push(c.Parents()...)
	}
	return seen
}

func SuccessorList(heads ...Commit) CommitSet {
	stack := NewCommitStack(heads...)
	seen := NewCommitSet()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) {
			continue
		}

		seen.Add(c)
		stack.Push(c.Children()...)
	}
	return seen
}

func (r *Repository) IsAncestor(base, head identifier.Hash) bool {
	baseCommit, ok := r.commits[base]
	if !ok {
		return false
	}

	headCommit, ok := r.commits[head]
	if !ok {
		return false
	}

	return IsAncestor(baseCommit, headCommit)
}

func IsAncestor(base, head Commit) bool {
	stack := CommitsStack{head}
	seen := NewCommitSet()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) {
			continue
		}

		if base == c {
			return true
		}

		seen.Add(c)
		stack.Push(c.Parents()...)
	}
	return false
}

func AncestorsListUntil(stopPoints CommitSet, heads ...Commit) CommitSet {
	stack := NewCommitStack(heads...)
	seen := NewCommitSet()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) {
			continue
		}

		seen.Add(c)

		if stopPoints.Has(c) {
			continue
		}

		stack.Push(c.Parents()...)
	}
	return seen
}

func UnseenAncestors(seen CommitSet, heads ...Commit) CommitSet {
	stack := NewCommitStack(heads...)
	unseen := NewCommitSet()

	for !stack.IsEmpty() {
		c := stack.Pop()
		if seen.Has(c) || unseen.Has(c) {
			continue
		}

		unseen.Add(c)
		stack.Push(c.Parents()...)
	}
	return unseen
}

//UnseenRoots return the unseen roots of the head commits.
func UnseenRoots(ctx context.Context, seen CommitSet, heads ...Commit) (CommitSet, error) {
	stack := NewCommitStack(heads...)
	if seen == nil {
		return nil, errors.New("missing the seen commits list")
	}

	scanned := NewCommitSet()
	ancestors := NewCommitSet()
	for !stack.IsEmpty() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		c := stack.Pop()
		if !seen.Has(c) {
			if !scanned.Has(c) {
				stack.Push(c.Parents()...)
				scanned.Add(c)
			}
			continue
		}

		// if seen, we add unseen children
		for _, child := range c.Children() {
			if seen.Has(child) {
				continue
			}

			var seenParents int
			for _, p := range child.Parents() {
				if seen.Has(p) {
					seenParents += 1
				}
			}

			if child.NumParents() == seenParents {
				ancestors.Add(child)
			}
		}
	}
	return ancestors, nil
}

func (r *Repository) String() string {
	return fmt.Sprintf("<Repository path=%s>", r.path)
}

func (r *Repository) Roots() CommitSet {
	return r.roots
}

func (r *Repository) NumRoots() int {
	return len(r.roots)
}

func (r *Repository) CommitsCount() int {
	return len(r.commits)
}

func (r *Repository) getOrCreateCommit(h identifier.Hash) (Commit, bool, error) {
	if c, ok := r.commits[h]; ok {
		return c, true, nil
	}

	c, err := r.adapter.Commit(h)
	if err != nil {
		return nil, false, err
	}

	// register the commit pointer in the repository
	r.commits[h] = c

	return c, false, nil
}

func (r *Repository) addRoot(c Commit) {
	r.roots.Add(c)
}

type HashesStack []identifier.Hash

func (s *HashesStack) Push(hashes ...identifier.Hash) {
	*s = append(*s, hashes...)
}

func (s *HashesStack) Pop() identifier.Hash {
	n := len(*s) - 1
	// ge the last item
	h := (*s)[n]
	// delete it from the stack
	*s = (*s)[:n]

	return h
}

func (s *HashesStack) Len() int {
	return len(*s)
}

type CommitsStack []Commit

func NewCommitStack(init ...Commit) CommitsStack {
	// copy the `init` slice to avoid corrupting its data
	stack := make(CommitsStack, len(init))
	copy(stack, init)
	return stack
}

func (s *CommitsStack) Pop() Commit {
	n := len(*s) - 1
	// ge the last item
	c := (*s)[n]
	// avoid memory leak
	(*s)[n] = nil
	// delete the item from the stack
	*s = (*s)[:n]
	return c
}

func (s *CommitsStack) Push(c ...Commit) {
	*s = append(*s, c...)
}

func (s CommitsStack) IsEmpty() bool {
	return len(s) == 0
}

func (r *Repository) IngestHead(ctx context.Context, head identifier.Hash) error {
	headCommit, exist, err := r.getOrCreateCommit(head)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	stack := CommitsStack{headCommit}

	for !stack.IsEmpty() {
		c := stack.Pop()
		obj, err := c.Object()
		if err != nil {
			return err
		}
		c.SetAuthorDate(obj.AuthorDate())
		c.SetDeveloper(Developer(obj.AuthorEmail())) // todo: should be based on developer aggregation

		parents := obj.ParentHashes()
		if len(parents) == 0 {
			r.addRoot(c)
		}
		obj.Free()

		for i := range parents {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			p, exist, err := r.getOrCreateCommit(parents[i])
			if err != nil {
				return err
			}

			p.AddChild(c)
			c.AddParent(p)

			if !exist {
				stack.Push(p)
			}
		}
	}
	return nil
}

type ChunkAddr struct {
	// The indexes of the first and last lines the chunkAddr
	Start, End int
}

type DeltaType int8

var _StageEnumToStageValue = make(map[string]DeltaType, len(_DeltaTypeNameToValue))
var _StageValueToStageQuotedEnum = make(map[DeltaType]string, len(_DeltaTypeNameToValue))

func init() {
	for k := range _DeltaTypeValueToName {
		name := strings.ReplaceAll(strings.ToUpper(k.String()), " ", "_")
		_StageEnumToStageValue[name] = k
		_StageValueToStageQuotedEnum[k] = strconv.Quote(name)
	}
}

//go:generate stringer -type=DeltaType -trimprefix=Delta -linecomment
//go:generate jsonenums -type=DeltaType
const (
	DeltaDeleted DeltaType = iota - 1
	DeltaUnmodified
	DeltaAdded
	DeltaModified
	DeltaRenamed
	DeltaCopied
	DeltaIgnored
	DeltaUntracked
	DeltaTypeChange
	DeltaUnreadable
	DeltaConflicted
	DeltaOther
)

func (t *DeltaType) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	val, ok := _StageEnumToStageValue[s]
	if !ok {
		return fmt.Errorf("invalid delta type: %q", s)
	}
	*t = val
	return nil
}

func (t DeltaType) MarshalGQL(w io.Writer) {
	v, ok := _StageValueToStageQuotedEnum[t]
	if !ok {
		fmt.Fprint(w, strconv.Quote(t.String()))
		return
	}

	fmt.Fprint(w, v)
}

type Developer string

type Subsystem string

func (sub Subsystem) Len() int {
	return len(string(sub))
}

func SubsystemFromPath(p string) string {
	i := strings.Index(p, "/")
	return p[:i+1]
}

func FindCommonAncestor(commits []Commit) (Commit, error) {
	seen := make(map[Commit]int)
	queue := append(commits[:0:0], commits...)

	for len(queue) > 0 {
		commit, queue := queue[0], queue[1:]

		i := seen[commit]
		i += 1
		if i == len(commits) {
			return commit, nil
		}
		seen[commit] = i

		queue = append(queue, commit.Parents()...)
	}

	return nil, fmt.Errorf("cannot find common ancestor")
}

//deprecated
func FindCommonFirstParent(commits []Commit) (Commit, error) {
	seen := make(map[Commit]int)
	queue := append(commits[:0:0], commits...)

	for len(queue) > 0 {
		commit, queue := queue[0], queue[1:]

		i := seen[commit]
		i += 1
		if i == len(commits) {
			return commit, nil
		}
		seen[commit] = i

		queue = append(queue, commit.FirstParent())
	}

	return nil, fmt.Errorf("cannot find common first parent")
}

type Signature struct {
	Name  string    `json:"name"  bson:"name"`
	Email string    `json:"email" bson:"email"`
	When  time.Time `json:"date"  bson:"date"`
}

type (
	FileDiffCB func(DiffDelta) error
	HunkDiffCB func(DiffDelta) (HunkAnalysis, error)
	LineDiffCB func(DiffDelta) (LineAnalysis, error)
)

type HunkAnalysis interface {
	AnalyzeHunk(DiffHunk) error
}

type LineAnalysis interface {
	AnalyzeHunk(DiffHunk) error
	AnalyzeLine(DiffLine) error
}
