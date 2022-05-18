package entity

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const commitCursorLen = 8 + len(CommitCursor{}.CommitHash)

type CommitCursor struct {
	RepoID     identifier.RepositoryID `bson:"_id.r"`
	AuthorDate time.Time               `bson:"author.date"`
	CommitHash identifier.Hash         `bson:"_id.h"`
}

//deprecated
func nodeToCursor(input interface{}) *string {
	switch e := input.(type) {
	case *Commit:
		return nodeToCommitCursor(e)

	default:
		//todo: add more info
		panic("unimplemented courser")
	}
}

// nodeToItemCursor is a dummy function for code generation.
func nodeToItemCursor(n *Item) *string {
	return nil
}

func nodeToCommitCursor(n *Commit) *string {
	return (&CommitCursor{
		AuthorDate: n.Author.When,
		CommitHash: n.ID.CommitHash,
	}).Base64()
}

func nodeToRepositoryCursor(n *Repository) *string {
	//todo: should be based on the sort index
	s := n.ID.Base64()
	return &s
}

func nodeToPullRequestCursor(n *PullRequest) *string {
	//todo: should be based on the sort index
	s := n.ID.Base64()
	return &s
}

func nodeToFeedbackCursor(n *Feedback) *string {
	//todo: should be based on the sort index
	s := n.ID.Base64()
	return &s
}

func nodeToOrganizationCursor(n *Organization) *string {
	//todo: should be based on the sort index
	s := n.ID.Base64()
	return &s
}

func nodeToJobCursor(n *Job) *string {
	s := n.ID.Base64()
	return &s
}

func objectIDToBase64(dst []byte, src *primitive.ObjectID) {
	base64.StdEncoding.Encode(dst, src[:])
}

func (c *CommitCursor) Base64() *string {
	if c == nil {
		return nil
	}

	buf := make([]byte, base64.StdEncoding.EncodedLen(commitCursorLen))
	commitCursorToBase64(buf, c)
	s := string(buf)

	return &s
}

func commitCursorToBase64(dst []byte, src *CommitCursor) {
	buf := dst[len(dst)-commitCursorLen:]
	binary.LittleEndian.PutUint64(buf, uint64(src.AuthorDate.Unix()))
	copy(buf[8:], src.CommitHash[:])

	base64.StdEncoding.Encode(dst, buf)
}

func base64ToCommitCursor(dst *CommitCursor, src []byte) error {
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(buf, src)
	if err != nil {
		return err
	}

	if n != commitCursorLen {
		return errors.New("invalid length")
	}

	sec := binary.LittleEndian.Uint64(buf)
	dst.AuthorDate = time.Unix(int64(sec), 0)

	copy(dst.CommitHash[:], buf[8:])

	return nil
}

func (c *CommitCursor) MarshalBase64() (data []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(commitCursorLen))

	commitCursorToBase64(buf, c)

	return buf, nil
}

func (c *CommitCursor) UnmarshalBase64(data []byte) error {
	return base64ToCommitCursor(c, data)
}

func (c CommitCursor) MarshalGQL(w io.Writer) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(commitCursorLen)+2)

	buf[0] = '"'
	commitCursorToBase64(buf[1:], &c)
	buf[len(buf)-1] = '"'

	w.Write(buf)
}

type PageInfo struct {
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	StartEdge       Edge
	EndEdge         Edge
}

func (info *PageInfo) PageInfo() *PageInfo {
	return info
}

func (info *PageInfo) StartCursor() *string {
	if info.StartEdge == nil {
		return nil
	}

	return info.StartEdge.Cursor()
}

func (info *PageInfo) EndCursor() *string {
	if info.EndEdge == nil {
		return nil
	}

	return info.EndEdge.Cursor()
}

type PaginationInput struct {
	Before *string
	After  *string
	First  *int
	Last   *int
}

func (opts *PaginationInput) Validate(connection string, limit int64) error {
	if opts.First != nil && opts.Last != nil {
		return fmt.Errorf("passing both `first` and `last` to paginate the `%s` connection is not supported", connection)
	}

	//if opts.After != nil && opts.Before != nil {
	//	return fmt.Errorf("passing both `after` and `before` to paginate the `%s` connection is not supported", connection)
	//}

	if opts.First == nil && opts.Last == nil {
		return fmt.Errorf("you must provide a `first` or `last` value to properly paginate the `%s` connection", connection)
	}

	if opts.First != nil {
		input := (int64)(*opts.First)
		if input > limit {
			return &ErrExceedPaginationLimit{
				Limit:      limit,
				Connection: connection,
				Input:      input,
				Direction:  "first",
			}
		}

		if input < 0 {
			return fmt.Errorf("`first` on the `%s` connection cannot be less than zero", connection)
		}
	}

	if opts.Last != nil {
		input := (int64)(*opts.Last)
		if input > limit {
			return &ErrExceedPaginationLimit{
				Limit:      limit,
				Connection: connection,
				Input:      input,
				Direction:  "last",
			}
		}

		if input < 0 {
			return fmt.Errorf("`last` on the `%s` connection cannot be less than zero", connection)
		}
	}

	return nil
}

type Edge interface {
	Cursor() *string
}

type ErrExceedPaginationLimit struct {
	Limit      int64
	Connection string
	Input      int64
	Direction  string
}

func (e *ErrExceedPaginationLimit) Error() string {
	return fmt.Sprintf("requesting %d records on the `%s` connection exceeds the `%s` limit of %d records",
		e.Input, e.Connection, e.Direction, e.Limit)
}
