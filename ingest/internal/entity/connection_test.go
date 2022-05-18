package entity

import (
	"testing"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

const expectedBase64CommitCursor = "GOHnXgAAAADmpb7Pd7DUkPc6i09/EoODDy2Rzg=="

var expectedCommitCursor = CommitCursor{
	AuthorDate: time.Unix(1592254744, 0),
	CommitHash: identifier.NewHash("e6a5becf77b0d490f73a8b4f7f1283830f2d91ce"),
}

func TestCommitCursor_MarshalBase64(t *testing.T) {
	date, err := expectedCommitCursor.MarshalBase64()
	if err != nil {
		t.Fatal(err)
	}

	if string(date) != expectedBase64CommitCursor {
		t.Errorf("expected base64 string: \"%s\", got \"%s\"", expectedBase64CommitCursor, string(date))
	}
}

func TestCommitCursor_UnmarshalBase64(t *testing.T) {
	var commitCursor CommitCursor

	err := commitCursor.UnmarshalBase64([]byte(expectedBase64CommitCursor))
	if err != nil {
		t.Fatal(err)
	}

	if commitCursor != expectedCommitCursor {
		t.Fatalf("expected commit cursor: %s,\ngot: %s", expectedCommitCursor, commitCursor)
	}
}
