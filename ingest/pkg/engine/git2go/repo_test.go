package git2go

import (
	"testing"

	git "github.com/libgit2/git2go/v31"
)

func TestBlame(t *testing.T) {
	t.Skip("the test has a known error in libgit2")

	path := "../../../repos/tests/neutron"
	repo, err := git.OpenRepository(path)
	if err != nil {
		if !git.IsErrorCode(err, git.ErrNotFound) {
			t.Fatal(err)
		}
		repo, err = git.Clone("https://github.com/openstack/neutron", path, &git.CloneOptions{Bare: true})
		if err != nil {
			t.Fatal(err)
		}
	}

	id, err := git.NewOid("9f1c2488260868ce0c4a5ed60219c92679829281")
	if err != nil {
		t.Fatal(err)
	}

	opts := &git.BlameOptions{
		Flags:              git.BlameNormal,
		MinMatchCharacters: 20,
		NewestCommit:       id,
	}

	blame, err := repo.BlameFile("quantum/manager.py", opts)
	if err != nil {
		t.Fatal(err)
	}

	hunk, err := blame.HunkByLine(51)
	if err != nil {
		t.Fatal(err)
	}

	commit, err := repo.LookupCommit(hunk.FinalCommitId)
	if err != nil {
		t.Fatal(err)
	}

	if hunk.FinalSignature == nil {
		// fixme: it is a known error in libgit2
		t.Fatal("missed signature")
	}

	cs := commit.Author()
	hs := hunk.FinalSignature

	if !hs.When.Equal(cs.When) || hs.Name != cs.Name || hs.Email != cs.Email {
		t.Fatal("unexpected signature")
	}
}

func TestMessage(t *testing.T) {
	t.Skip("skip to avoid git cloning")

	path := "../../../repos/tests/neutron"
	repo, err := git.OpenRepository(path)
	if err != nil {
		if !git.IsErrorCode(err, git.ErrNotFound) {
			t.Fatal(err)
		}
		repo, err = git.Clone("https://github.com/openstack/neutron", path, &git.CloneOptions{Bare: true})
		if err != nil {
			t.Fatal(err)
		}
	}

	id, err := git.NewOid("d7c23431ad3959eb5fd74e42ea95d446e4e7566d")
	if err != nil {
		t.Fatal(err)
	}

	commit, err := repo.LookupCommit(id)
	if err != nil {
		t.Fatal(err)
	}

	commit.Message()
}
