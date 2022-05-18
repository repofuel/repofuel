package identifier

import (
	"bytes"
	"testing"
)

const expectedBase64CommitID = "\"Q29tbWl0Ol7XFRiOzwLLM9BMUualvs93sNSQ9zqLT38Sg4MPLZHO\""

var expectedCommitID = CommitID{
	RepoID:     objectIDFromHex("5ed715188ecf02cb33d04c52"),
	CommitHash: NewHash("e6a5becf77b0d490f73a8b4f7f1283830f2d91ce"),
}

func objectIDFromHex(s string) RepositoryID {
	id, _ := RepositoryIDFromHex(s)
	return id
}

func TestCommitID_MarshalGQL(t *testing.T) {
	var b bytes.Buffer
	expectedCommitID.MarshalGQL(&b)

	if b.String() != expectedBase64CommitID {
		t.Errorf("expected base64 string: \"%s\", got %s", expectedBase64CommitID, b.String())
	}
}

func TestCommitID_UnmarshalGQL(t *testing.T) {
	var commitID CommitID

	err := commitID.UnmarshalGQL(expectedBase64CommitID[1 : len(expectedBase64CommitID)-1])
	if err != nil {
		t.Fatal(err)
	}

	if commitID != expectedCommitID {
		t.Errorf("expected commit id: %s,\ngot: %s", expectedCommitID, commitID)
	}
}

func TestCommitIDFromStr(t *testing.T) {
	commitID, err := CommitIDFromStr(expectedCommitID.RepoID.Hex()+"_"+expectedCommitID.CommitHash.Hex(), '_')
	if err != nil {
		t.Fatal(err)
	}

	if *commitID != expectedCommitID {
		t.Errorf("expected commit id: %s,\ngot: %s", &expectedCommitID, commitID)
	}
}
