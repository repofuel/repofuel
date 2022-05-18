package identifier

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	commitIDPrefix = "Commit:"
	commitIDLen    = len(commitIDPrefix) + len(CommitID{}.RepoID) + len(CommitID{}.CommitHash)
)

type CommitID struct {
	RepoID     RepositoryID `json:"repo_id"   bson:"r"`
	CommitHash Hash         `json:"hash"      bson:"h"`
}

//deprecated
type marshalCommitID struct {
	RepoID     RepositoryID `bson:"r"`
	CommitHash Hash         `bson:"h"`
}

// todo: it should be deprecated
func CommitIDFromStr(id string, sep byte) (*CommitID, error) {
	splitIndex := strings.IndexByte(id, sep)

	repoID := []byte(id[:splitIndex])
	commitHash := []byte(id[splitIndex+1:])

	var c CommitID

	_, err := hex.Decode(c.CommitHash[:], commitHash)
	if err != nil {
		return nil, fmt.Errorf("decode commit hash: %w", err)
	}

	_, err = hex.Decode(c.RepoID[:], repoID)
	if err != nil {
		return nil, fmt.Errorf("decode repo id: %w", err)
	}

	return &c, nil
}

func (c *CommitID) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("commit ID must be a string")
	}

	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}

	if len(b) != commitIDLen {
		return fmt.Errorf("incorrect CommitID bytes length, it shouhd be %d, got: %d", commitIDLen, len(b))
	}

	copy(c.RepoID[:], b[len(commitIDPrefix):len(commitIDPrefix)+len(c.RepoID)])
	copy(c.CommitHash[:], b[commitIDLen-len(c.CommitHash):])

	return nil
}

func (c CommitID) MarshalGQL(w io.Writer) {
	b := make([]byte, 0, commitIDLen)
	b = append(b, commitIDPrefix...)
	b = append(b, c.RepoID[:]...)
	b = append(b, c.CommitHash[:]...)

	res := make([]byte, base64.StdEncoding.EncodedLen(commitIDLen)+2)
	res[0] = '"'
	base64.StdEncoding.Encode(res[1:], b)
	res[len(res)-1] = '"'

	w.Write(res)
}

func (c *CommitID) String() string {
	return toStringCommitID(c.RepoID, c.CommitHash)
}

func NewCommitID(repoID RepositoryID, commitHash Hash) *CommitID {
	return &CommitID{
		RepoID:     repoID,
		CommitHash: commitHash,
	}
}

func CommitIDFromBytes(b []byte) *CommitID {
	var c CommitID

	copy(c.RepoID[:], b)
	copy(c.CommitHash[:], b[len(c.RepoID):])

	return &c
}

func toStringCommitID(repoID RepositoryID, commitHash Hash) string {
	var b strings.Builder
	b.WriteString(repoID.Hex())
	b.WriteString("_")
	b.WriteString(commitHash.Hex())
	return b.String()
}

// deprecated
func CommitIDFormStrHash(repoID RepositoryID, strHash string) string {
	var b strings.Builder
	b.WriteString(repoID.Hex())
	b.WriteString("_")
	b.WriteString(strHash)
	return b.String()
}
