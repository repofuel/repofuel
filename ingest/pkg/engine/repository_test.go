package engine

import (
	"bytes"
	"testing"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

//todo: to be moved to the identifier package to test the code generation

const expectedBase64RepoID = "\"UmVwb3NpdG9yeTpex0Prbb6R9xqLBD4=\""

var expectedRepoID, _ = identifier.RepositoryIDFromHex("5ec743eb6dbe91f71a8b043e")

func TestRepoID_MarshalGQL(t *testing.T) {
	var b bytes.Buffer
	expectedRepoID.MarshalGQL(&b)

	if b.String() != expectedBase64RepoID {
		t.Errorf("expected base64 string: %s, got %s", expectedBase64RepoID, b.String())
	}
}

func TestRepoID_UnmarshalGQL(t *testing.T) {
	var id identifier.RepositoryID

	err := id.UnmarshalGQL(expectedBase64RepoID[1 : len(expectedBase64RepoID)-1])
	if err != nil {
		t.Error(err)
	}

	if id != expectedRepoID {
		t.Errorf("expected RepositoryID: %s, got: %s", expectedRepoID, id)
	}
}
