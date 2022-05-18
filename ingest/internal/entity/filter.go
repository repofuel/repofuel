package entity

import (
	"go.mongodb.org/mongo-driver/bson"
)

type FieldsProjection []string

func (f FieldsProjection) MarshalBSON() ([]byte, error) {
	r := make(bson.M, len(f))
	for i := range f {
		r[f[i]] = 1
	}

	return bson.Marshal(r)
}
