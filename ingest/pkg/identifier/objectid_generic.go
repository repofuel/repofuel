package identifier

import (
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/cheekybits/genny/generic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate genny -in=objectid_generic.go -out=objectid_gnerated.go gen "Item=Repository,PullRequest,Job,User,Feedback,Organization"

type Item generic.Type

const (
	_ItemIDPrefix = "Item:"
	_ItemIDLen    = len(_ItemIDPrefix) + len(ItemID{})
)

type ItemID primitive.ObjectID

func (id ItemID) String() string {
	return id.Hex()
}

func (id *ItemID) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("repo ID must be a string")
	}

	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}

	if len(b) != _ItemIDLen {
		return fmt.Errorf("incorrect CommitID bytes length, it shouhd be %d, got: %d", _ItemIDLen, len(b))
	}

	copy(id[:], b[len(_ItemIDPrefix):])

	return nil
}

func (id ItemID) MarshalGQL(w io.Writer) {
	b := make([]byte, 0, _ItemIDLen)
	b = append(b, _ItemIDPrefix...)
	b = append(b, id[:]...)

	res := make([]byte, base64.StdEncoding.EncodedLen(_ItemIDLen)+2)
	res[0] = '"'
	base64.StdEncoding.Encode(res[1:], b)
	res[len(res)-1] = '"'

	w.Write(res)
}

func (id ItemID) NodeID() string {
	b := make([]byte, 0, _ItemIDLen)
	b = append(b, _ItemIDPrefix...)
	b = append(b, id[:]...)

	res := make([]byte, base64.StdEncoding.EncodedLen(_ItemIDLen))
	base64.StdEncoding.Encode(res, b)

	return string(res)
}

func (id *ItemID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t != bsontype.ObjectID {
		return fmt.Errorf("invalid RepositoryID")
	}

	copy(id[:], data)

	return nil
}

func (id ItemID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(primitive.ObjectID(id))
}

func (id ItemID) MarshalJSON() ([]byte, error) {
	return primitive.ObjectID(id).MarshalJSON()
}

func (id *ItemID) UnmarshalJSON(b []byte) error {
	return (*primitive.ObjectID)(id).UnmarshalJSON(b)
}

func (id ItemID) Hex() string {
	return primitive.ObjectID(id).Hex()
}

func (id ItemID) Base64() string {
	res := make([]byte, base64.StdEncoding.EncodedLen(len(id)))
	base64.StdEncoding.Encode(res, id[:])

	return string(res)
}

func (id ItemID) IsZero() bool {
	return primitive.ObjectID(id).IsZero()
}

func (id ItemID) Timestamp() time.Time {
	return primitive.ObjectID(id).Timestamp()
}

func ItemIDFromHex(s string) (ItemID, error) {
	oid, err := primitive.ObjectIDFromHex(s)
	return ItemID(oid), err
}

func ItemIDFromBytes(b []byte) ItemID {
	var id ItemID
	copy(id[:], b)
	return id
}

func ItemIDFromNodeID(s string) (ItemID, error) {
	var id ItemID
	err := id.UnmarshalGQL(s)
	return id, err
}
