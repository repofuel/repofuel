package marshals

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalFloat32(f float32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf("%g", f))
	})
}

func UnmarshalFloat32(v interface{}) (float32, error) {
	switch v := v.(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return float32(f), err
	case int:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case float64:
		return float32(v), nil
	case float32:
		return v, nil
	case json.Number:
		f, err := strconv.ParseFloat(string(v), 64)
		return float32(f), err
	default:
		return 0, fmt.Errorf("%T is not an float", v)
	}
}

func MarshalDuration(d time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf("%d", d.Milliseconds()))
	})
}

func UnmarshalDuration(v interface{}) (time.Duration, error) {
	i, err := graphql.UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	return time.Duration(i) * time.Millisecond, nil
}
