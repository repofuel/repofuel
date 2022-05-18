// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package git2go

import (
	"context"
	"fmt"
	"os"
	"strings"

	git "github.com/libgit2/git2go/v30"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/rs/zerolog/log"
)

var certificateCheckCallback git.CertificateCheckCallback

func init() {
	err := git.SetCacheMaxSize(1048576)
	if err != nil {
		log.Fatal().Err(err).Msg("set the max cache size for for git")
	}

	if strings.EqualFold(os.Getenv("CERTIFICATE_CHECK"), "off") {
		log.Warn().Msg("SSL certificate checking is disabled on git cloning and fetching")
		certificateCheckCallback = func(_ *git.Certificate, _ bool, _ string) git.ErrorCode {
			return git.ErrOk
		}
	}
}

type Adapter struct {
	git *git.Repository
}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (adp *Adapter) Open(path string) error {
	var err error
	adp.git, err = git.OpenRepository(path)
	if err != nil {
		if git.IsErrorCode(err, git.ErrNotFound) {
			return engine.ErrLocalRepoNotExist
		}
	}
	return err
}

func (adp *Adapter) Clone(ctx context.Context, url string, path string, getAuth engine.BasicAuthFunc) error {
	var err error
	adp.git, err = git.Clone(url, path, &git.CloneOptions{
		FetchOptions: defaultFetchOptions(ctx, getAuth),
		Bare:         true,
	})

	if gitErr, ok := err.(*git.GitError); ok && gitErr.Class == git.ErrClassInvalid && gitErr.Code == git.ErrExists {
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("clean up for fresh clone: %w", err)
		}

		// try again after delete the directory
		adp.git, err = git.Clone(url, path, &git.CloneOptions{
			FetchOptions: defaultFetchOptions(ctx, getAuth),
			Bare:         true,
		})
	}

	return err
}

const (
	defaultRemote       = "origin"
	defaultRemotePrefix = "origin/"
)

// only origin branches for now
func (adp *Adapter) Branches() (map[string]identifier.Hash, error) {
	itr, err := adp.git.NewBranchIterator(git.BranchRemote)
	if err != nil {
		return nil, err
	}

	var branches = make(map[string]identifier.Hash)
	err = itr.ForEach(func(b *git.Branch, _ git.BranchType) error {
		name, err := b.Name()
		if err != nil {
			return err
		}
		head := b.Target()
		if !strings.HasPrefix(name, defaultRemotePrefix) || head == nil {
			// ignore
			return nil
		}

		name = name[len(defaultRemotePrefix):] // remove the prefix
		branches[name] = identifier.Hash(*head)
		return nil
	})

	return branches, err
}

func (adp *Adapter) Commit(h identifier.Hash) (engine.Commit, error) {
	return newCommit(adp.git, git.Oid(h)), nil
}

func (adp *Adapter) Fetch(ctx context.Context, getAuth engine.BasicAuthFunc, remote, url string, branches ...string) error {
	err := adp.git.Remotes.SetUrl(remote, url)
	if err != nil {
		return err
	}

	re, err := adp.git.Remotes.Lookup(remote)
	if err != nil {
		return err
	}

	return re.Fetch(branches, defaultFetchOptions(ctx, getAuth), "")
}

func defaultFetchOptions(ctx context.Context, getAuth engine.BasicAuthFunc) *git.FetchOptions {
	return &git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			//SidebandProgressCallback: func(str string) git.ErrorCode {
			//	fmt.Println("SidebandProgressCallback:", str)
			//	return git.ErrOk
			//},
			//CompletionCallback: func(c git.RemoteCompletion) git.ErrorCode {
			//	fmt.Println("CompletionCallback", c)
			//	return git.ErrOk
			//},
			CredentialsCallback: func(_ string, _ string, _ git.CredType) (*git.Cred, error) {
				auth, err := getAuth(ctx)
				if err != nil {
					return nil, err
				}
				return git.NewCredUserpassPlaintext(auth.Username, auth.Password)
			},
			TransferProgressCallback: func(stats git.TransferProgress) git.ErrorCode {
				select {
				case <-ctx.Done():
					//todo: convert the error when it returned from git.Clone or git.Fetch
					// to context.Canceled in order to have the same expected behaviour
					return git.ErrUser
				default:
					return git.ErrOk
				}
			},
			//UpdateTipsCallback: func(refname string, a *git.Oid, b *git.Oid) git.ErrorCode {
			//	fmt.Println("UpdateTipsCallback:", refname, a, b)
			//	return git.ErrOk
			//},
			CertificateCheckCallback: certificateCheckCallback,
			//PackProgressCallback: func(stage int32, current, total uint32) git.ErrorCode {
			//	fmt.Println("PackProgressCallback:", "stage:", stage, "current:", current, "total:", total)
			//	return git.ErrOk
			//},
			//PushTransferProgressCallback: nil,
			//PushUpdateReferenceCallback:  nil,
		},
		Prune:           git.FetchPruneOn,
		UpdateFetchhead: true,
		DownloadTags:    git.DownloadTagsAuto,
	}
}

func (r *Adapter) InducingCommitsOneByOne(id identifier.Hash, path string, chunk engine.ChunkAddr) (identifier.HashSet, error) {
	oid := git.Oid(id)

	opts := &git.BlameOptions{
		Flags:              git.BlameNormal,
		MinMatchCharacters: 20,
		NewestCommit:       &oid,
		OldestCommit:       nil,
		MinLine:            uint32(chunk.Start),
		MaxLine:            uint32(chunk.End),
	}

	blame, err := r.git.BlameFile(path, opts)
	if err != nil {
		return nil, err
	}
	defer blame.Free()

	IDs := identifier.NewHashSet()
	for count, i := blame.HunkCount(), 0; i < count; i++ {
		hunk, err := blame.HunkByIndex(i)
		if err != nil {
			return nil, err
		}

		IDs.Add(identifier.Hash(*hunk.OrigCommitId))
	}
	return IDs, err
}

func (r *Adapter) InducingCommits(ctx context.Context, id identifier.Hash, path string, chunks ...engine.ChunkAddr) (identifier.HashSet, error) {
	oid := git.Oid(id)

	opts := &git.BlameOptions{
		Flags:              git.BlameNormal,
		MinMatchCharacters: 20,
		NewestCommit:       &oid,
		OldestCommit:       nil,
		MinLine:            uint32(chunks[0].Start),
		MaxLine:            uint32(chunks[len(chunks)-1].End),
	}

	blame, err := r.git.BlameFile(path, opts)
	if err != nil {
		return nil, err
	}
	defer blame.Free()

	IDs := identifier.NewHashSet()
	for i := range chunks {
		current := chunks[i]
		hunk, err := blame.HunkByLine(current.Start)
		if err != nil {
			return nil, err
		}
		IDs.Add(identifier.Hash(*hunk.OrigCommitId))

		endLine := int(hunk.FinalStartLineNumber + hunk.LinesInHunk - 1)
		for endLine < current.End {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			hunk, err = blame.HunkByLine(endLine + 1)
			if err != nil {
				//fixme: we could have git.ErrInvalid, we should be tolerant
				return nil, err
			}
			endLine = int(hunk.FinalStartLineNumber + hunk.LinesInHunk - 1)
			IDs.Add(identifier.Hash(*hunk.OrigCommitId))
		}
	}

	return IDs, err
}
