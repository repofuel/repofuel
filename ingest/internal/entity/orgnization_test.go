package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBSONMarshalEmptyOrganizationStruct(t *testing.T) {
	var doc Organization

	b, err := bson.Marshal(doc)
	assert.NoError(t, err)

	assert.Equal(t, b, []byte{5, 0, 0, 0, 0}, "empty struct should be marshaled to empty bson")
}
