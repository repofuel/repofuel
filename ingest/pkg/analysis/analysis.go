// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package analysis

import (
	"bytes"
	"context"
	"errors"
	"math"
	"path"

	"github.com/go-enry/go-enry/v2"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/classify"
	"github.com/repofuel/repofuel/ingest/pkg/commitmsg"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	ErrFixWithNoParents      = errors.New("fix commit does not have any parent")
	ErrNoFilesForCalculation = errors.New("commit doesn't have modified source code for metrics calculation")
)

const (
	szzMaxHD = 100
	szzMaxLD = 5000
	szzMaxNF = 50
)

type RepositoryAnalysis struct {
	repo    *engine.Repository
	job     identifier.JobID
	tracker ProgressTracker
	logger  *zerolog.Logger

	commitsDB entity.CommitDataSource
}

func (a *RepositoryAnalysis) Finish(context.Context) error {
	return nil
}

type ProgressTracker interface {
	IncreaseProgress(int)
}

func NewRepositoryAnalysis(repo *engine.Repository, job identifier.JobID, commitsDB entity.CommitDataSource, tracker ProgressTracker) *RepositoryAnalysis {
	return &RepositoryAnalysis{
		repo:      repo,
		job:       job,
		commitsDB: commitsDB,
		tracker:   tracker,
	}
}

func (a *RepositoryAnalysis) Run(ctx context.Context, roots engine.CommitSet) error {
	a.logger = log.Ctx(ctx)

	return engine.RunForwardAnalysis(ctx, a, roots.Slice())
}

func (a *RepositoryAnalysis) AnalyzeCommit(ctx context.Context, c engine.Commit) error {
	a.logger.Debug().Hex("commit", c.Hash().Bytes()).Msg("analyze commit")

	a.tracker.IncreaseProgress(1)

	obj, err := c.Object()
	if err != nil {
		return err
	}
	defer obj.Free()

	message := obj.Message()

	var ec entity.Commit
	ec.ID = identifier.NewCommitID(a.repo.ID, obj.Hash())
	ec.Author = obj.Author()
	ec.Message = entity.LimitedMessage(message)
	ec.Job = a.job
	ec.Branches = c.Branches().Slice()

	analyzedFiles, err := analyzeFiles(ctx, obj)
	if err != nil {
		a.logger.Err(err).
			Hex("commit", ec.ID.CommitHash[:]).
			Msg("analyze commit files")
		return a.storeCommit(ctx, &ec, analyzedFiles)
	}

	ec.Tags = classify.FindCategories(message).Slice()

	c.SetFiles(codeFilesList(analyzedFiles))

	if c.IsMerge() {
		// we do not calculate metrics for merges
		ec.Merge = true
		return a.storeCommit(ctx, &ec, analyzedFiles)
	}

	m, err := calculateMetrics(ctx, c, analyzedFiles)
	if err != nil {
		if err != ErrNoFilesForCalculation {
			a.logger.Err(err).
				Hex("commit", ec.ID.CommitHash[:]).
				Msg("calculate commit metrics")
		}
		return a.storeCommit(ctx, &ec, analyzedFiles)
	}
	ec.Metrics = m

	issues, includeBug, err := a.repo.IssuesFromText(ctx, message)
	if err != nil {
		a.logger.Err(err).
			Hex("commit", ec.ID.CommitHash[:]).
			Msg("fetch issues from commit message")
		return a.storeCommit(ctx, &ec, analyzedFiles)
	}
	ec.Issues = issues

	// ignore big commits
	if (includeBug || commitmsg.IsCorrective(message)) &&
		m.LD < szzMaxLD &&
		m.HD < szzMaxHD &&
		m.NF < szzMaxNF {

		ec.Fix = true
		bugs, err := a.traceDeletedChunks(ctx, c, analyzedFiles)
		if err != nil {
			a.logger.Err(err).
				Hex("commit", ec.ID.CommitHash[:]).
				Msg("trace bug inducing")
			return a.storeCommit(ctx, &ec, analyzedFiles)
		}

		err = a.commitsDB.MarkBuggy(ctx, a.repo.ID, c.Hash(), bugs)
		if err != nil {
			return err
		}
	}

	return a.storeCommit(ctx, &ec, analyzedFiles)
}

func codeFilesList(analyzedFiles []*fileAnalysis) map[string]*engine.FileInfo {
	if len(analyzedFiles) == 0 {
		return nil
	}

	var files = make(map[string]*engine.FileInfo, len(analyzedFiles))
	for _, f := range analyzedFiles {
		if f.Type == classify.FileCode {
			files[f.Path] = f.FileInfo
		}
	}

	return files
}

//todo: refactor, we stor the files in the commit earlier and then we do not need this function. We can call `InsertOrReplace` directly.
func (a *RepositoryAnalysis) storeCommit(ctx context.Context, ec *entity.Commit, analyzedFiles []*fileAnalysis) error {
	files := make([]*entity.File, len(analyzedFiles))
	for i, fa := range analyzedFiles {
		files[i] = &entity.File{
			FileInfo:      fa.FileInfo,
			Type:          fa.Type,
			Language:      fa.Language,
			Fixing:        nil,
			Metrics:       &fa.FileMeasures,
			SameDeveloper: fa.SameDeveloper,
		}
	}

	ec.Files = files

	return a.commitsDB.InsertOrReplace(ctx, ec)
}

func (a *RepositoryAnalysis) traceDeletedChunks(ctx context.Context, c engine.Commit, analyzedFiles []*fileAnalysis) (identifier.HashSet, error) {
	if c.NumParents() == 0 {
		// Example of a fix commit that does not have parents:
		// https://github.com/androguard/androguard/commit/d4770c4b4b1cc42fbe118d0e543ae33abdf27fba
		return nil, ErrFixWithNoParents
	}

	buggies := identifier.NewHashSet()
	for _, f := range analyzedFiles {
		if len(f.DeletedChunks) == 0 {
			continue
		}

		hashes, err := a.repo.InducingCommits(ctx, c.FirstParent().Hash(), f.OldOrNewPath(), f.DeletedChunks...)
		if err != nil {
			return nil, err
		}

		buggies.Update(hashes)
		f.Fixing = hashes.Slice()
		f.Fix = len(hashes) > 0
	}
	return buggies, nil
}

type fileAnalysis struct {
	*engine.FileInfo
	metrics.FileMeasures

	SameDeveloper bool
	Language      string
	Type          classify.FileType
	DeletedChunks []engine.ChunkAddr
	Fixing        []identifier.Hash
	Developers    engine.DeveloperSet
}

func (f *fileAnalysis) OldOrNewPath() string {
	if f.OldPath == "" {
		return f.Path
	}
	return f.OldPath
}

func analyzeFiles(ctx context.Context, obj engine.CommitObject) ([]*fileAnalysis, error) {
	var analyzedFiles []*fileAnalysis //todo: we can know the size from the diff before iterate over

	err := obj.DiffHunks(ctx, func(delta engine.DiffDelta) (engine.HunkAnalysis, error) {
		f, err := analyzeFile(delta)
		if err != nil {
			return nil, err
		}

		analyzedFiles = append(analyzedFiles, f)

		return f, err
	})
	if err != nil {
		return nil, err
	}

	return analyzedFiles, nil
}

func analyzeFile(f engine.DiffDelta) (*fileAnalysis, error) {
	fa := &fileAnalysis{
		FileInfo:   new(engine.FileInfo),
		Developers: engine.NewDeveloperSet(),
	}

	fa.Action = f.Action()
	switch fa.Action {
	case engine.DeltaRenamed:
		fa.Path = f.ToPath()
		fa.OldPath = f.FromPath()
	case engine.DeltaDeleted:
		fa.Path = f.FromPath()
	default:
		fa.Path = f.ToPath()
	}

	switch {
	case f.IsSymlink():
		fa.Type = classify.FileSymlink
	case f.IsBinary():
		fa.Type = classify.FileBinary
	case enry.IsVendor(fa.Path):
		fa.Type = classify.FileDependency
	case enry.IsTest(fa.Path):
		fa.Type = classify.FileTests
	case enry.IsDocumentation(fa.Path):
		fa.Type = classify.FileDocumentation
	case enry.IsConfiguration(fa.Path):
		fa.Type = classify.FileConfiguration
	}
	if fa.Type != 0 {
		return fa, nil
	}

	content, err := f.NewContent()
	if err != nil {
		return nil, err
	}

	if content == nil {
		content, err = f.OldContent()
		if err != nil {
			return nil, err
		}
		fa.LT = float64(numLines(content))
	} else {
		fa.LT = float64(numLines(content)) + fa.LD - fa.LA
	}

	if enry.IsGenerated(fa.Path, content) {
		fa.Type = classify.FileGenerated
		return fa, nil
	}

	// fixme: using newContent is more precise and old content could be used only for deleted files.
	lang := enry.GetLanguage(fa.Path, content)
	if enry.GetLanguageType(lang) == enry.Programming {
		fa.Type = classify.FileCode
		fa.Subsystem = engine.SubsystemFromPath(fa.Path)
		fa.Language = lang
	}

	return fa, nil
}

var lineSep = []byte{'\n'}

func numLines(content []byte) int {
	num := bytes.Count(content, lineSep)
	if !bytes.HasSuffix(content, lineSep) {
		num += 1
	}

	return num
}

func (f *fileAnalysis) Directory() string {
	dir, _ := path.Split(f.Path)
	return dir
}

func (f *fileAnalysis) AnalyzeHunk(hunk engine.DiffHunk) error {
	if f.Type != classify.FileCode {
		// we ignore none source code files
		return nil
	}

	if hunk.LinesAdded() > 0 {
		f.HA += 1
		f.LA += float64(hunk.LinesAdded())
	}

	if hunk.LinesDeleted() > 0 {
		f.HD += 1
		f.LD += float64(hunk.LinesDeleted())
		f.DeletedChunks = append(f.DeletedChunks, hunk.AddressDeleted())
	}

	return nil
}

func calculatePositiveAge(c engine.Commit, ancestor engine.Commit) float64 {
	age := c.AuthorDate().Sub(ancestor.AuthorDate()).Hours() / 24
	if age < 1 {
		return 1
	}

	return age
}

//TODO: bots should be marked or filtered out
//TODO: consider the co-author in the experience
//TODO: consider the committer (if different developer) in the experience
func calculateMetrics(ctx context.Context, c engine.Commit, analyzedFiles []*fileAnalysis) (*metrics.ChangeMeasures, error) {
	var subsystems = engine.NewStringSet()
	var directories = engine.NewStringSet()
	var trackedFiles = make(map[string]*fileAnalysis, len(analyzedFiles))
	var nf, nfWithAge, la, ld, ha, hd, lt, sexp, exp, age, rexp float64

	for _, fa := range analyzedFiles {
		if isIgnoredFile(fa) {
			continue
		}

		subsystems.Add(fa.Subsystem)
		directories.Add(fa.Directory())

		if fa.Action != engine.DeltaAdded {
			lastChange, err := engine.LastFileChange(ctx, c, fa.OldOrNewPath())
			if err != nil {
				lastChange = c
				log.Ctx(ctx).Err(err).
					Hex("commit", c.Hash().Bytes()).
					Str("file", fa.OldOrNewPath()).
					Msg("find last change for a file")
			} else {
				fa.SameDeveloper = c.Developer() == lastChange.Developer()
			}

			fa.AGE = c.AuthorDate().Sub(lastChange.AuthorDate()).Hours() / 24
			if fa.AGE > 0 {
				nfWithAge += 1
				age += fa.AGE
			}
		}

		trackedFiles[fa.Path] = fa

		// Sum the file metrics
		nf += 1
		la += fa.LA
		ld += fa.LD
		ha += fa.HA
		hd += fa.HD
		lt += fa.LT
	}

	if nf == 0 {
		return nil, ErrNoFilesForCalculation
	}

	if nfWithAge > 0 {
		age = age / nfWithAge
	}

	changes := engine.NewCommitSet()
	developers := engine.NewDeveloperSet()

	err := VisitAncestors(ctx, c, trackedFiles, func(p engine.Commit, trackedFiles map[string]*fileAnalysis) {
		var commonDevSubsystem bool
		var commonDevFile bool
		var commonFile bool
		var dev = p.Developer()
		var commonDev = c.Developer() == dev

		for filepath, f := range p.Files() {
			// common files
			if fa, ok := trackedFiles[filepath]; ok {

				if !commonFile {
					commonFile = true
					if commonDev {
						commonDevFile = true
					}

					changes.Add(p)
					developers.Add(dev)
				}

				if !fa.Developers.Has(dev) {
					fa.Developers.Add(dev)
					fa.NDEV += 1
				}

				if f.Fix {
					fa.NFC += 1
				}

				if commonDev {
					fa.EXP += 1
					fa.REXP += 1 / calculatePositiveAge(c, p)
				}

				fa.NUC += 1

				// adjust the file list to follow the renames
				switch f.Action {
				case engine.DeltaAdded:
					delete(trackedFiles, filepath)
				case engine.DeltaRenamed:
					delete(trackedFiles, filepath)
					trackedFiles[f.OldPath] = fa
				}
			}

			// common developer subsystems
			if commonDev && !commonDevSubsystem && subsystems.Has(f.Subsystem) {
				commonDevSubsystem = true
			}
		}

		if nf == 0 {
			// Do not count experience if no code files
			return
		}

		if commonDev {
			exp += 1
		}

		if commonDevSubsystem {
			sexp += 1
		}

		if commonDevFile {
			rexp += 1 / calculatePositiveAge(c, p)
		}
	})
	if err != nil {
		return nil, err
	}

	return &metrics.ChangeMeasures{
		NS:      float64(subsystems.Count()),
		ND:      float64(directories.Count()),
		NF:      nf,
		Entropy: calculateEntropy(la, ld, analyzedFiles),
		LA:      la,
		LD:      ld,
		HA:      ha,
		HD:      hd,
		LT:      lt / nf, // fixme: should we keep it without normalizin?
		NDEV:    float64(developers.Count()),
		AGE:     age,
		NUC:     float64(changes.Count()),
		EXP:     exp,
		REXP:    rexp,
		SEXP:    sexp,
	}, nil
}

func calculateEntropy(la, ld float64, files []*fileAnalysis) float64 {
	// Number of modified lines in all files
	modLines := la + ld
	if modLines == 0 {
		return 0
	}

	entropy := 0.0
	for _, f := range files {
		if isIgnoredFile(f) {
			continue
		}

		p := (f.LA + f.LD) / modLines
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func isIgnoredFile(f *fileAnalysis) bool {
	// we do  not consider files that don't change code
	return f.Type != classify.FileCode || f.LA+f.LD == 0
}

func VisitAncestors(ctx context.Context, c engine.Commit, trackedFiles map[string]*fileAnalysis, fn func(engine.Commit, map[string]*fileAnalysis)) error {
	if !c.HasParent() {
		return nil
	}

	seen := engine.NewCommitSet()
	commitsStack := engine.CommitsStack{c.FirstParent()}
	filesStack := fileAnalysisStack{trackedFiles}

	// skip the  current commit and add its parents

	for !commitsStack.IsEmpty() {
		c := commitsStack.Pop()
		f := filesStack.Pop()
		fn(c, f)

		parents := c.Parents()
		for i := len(parents) - 1; i >= 0; i -= 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			p := parents[i]

			if seen.Has(p) {
				continue
			}

			if i > 0 {
				f = cloneFileAnalysis(f)
			}

			seen.Add(p)
			commitsStack.Push(p)
			filesStack.Push(f)
		}
	}
	return nil
}

func cloneFileAnalysis(m map[string]*fileAnalysis) map[string]*fileAnalysis {
	r := make(map[string]*fileAnalysis, len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

type fileAnalysisStack []map[string]*fileAnalysis

func (s *fileAnalysisStack) Pop() map[string]*fileAnalysis {
	n := len(*s) - 1
	// ge the last item
	item := (*s)[n]
	// avoid memory leak
	(*s)[n] = nil
	// delete the item from the stack
	*s = (*s)[:n]
	return item
}

func (s *fileAnalysisStack) Push(item map[string]*fileAnalysis) {
	*s = append(*s, item)
}

func (s fileAnalysisStack) IsEmpty() bool {
	return len(s) == 0
}
