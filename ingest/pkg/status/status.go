package status

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

//go:generate stringer -type=Stage -linecomment
//go:generate jsonenums -type=Stage
type Stage uint8

const (
	Queued Stage = iota
	Added
	Cloning
	Cloned
	Fetching
	Fetched
	Ingesting
	Ingested
	Analyzing
	Analyzed
	Predicting
	Predicted
	Ready
	Failed
	Canceled
	Watched
	Recovered
	Ignored
	Progressing
)

var _StageEnumToStageValue = make(map[string]Stage, len(_StageNameToValue))
var _StageValueToStageQuotedEnum = make(map[Stage]string, len(_StageNameToValue))

func init() {
	for k := range _StageValueToName {
		name := strings.ReplaceAll(strings.ToUpper(k.String()), " ", "_")
		_StageEnumToStageValue[name] = k
		_StageValueToStageQuotedEnum[k] = strconv.Quote(name)
	}
}

func (i *Stage) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	enumsValue, ok := _StageEnumToStageValue[s]
	if !ok {
		return fmt.Errorf("invalid Stage %q", s)
	}
	*i = enumsValue
	return nil
}

func (i Stage) MarshalGQL(w io.Writer) {
	v, ok := _StageValueToStageQuotedEnum[i]
	if !ok {
		fmt.Fprint(w, strconv.Quote(i.String()))
		return
	}

	fmt.Fprint(w, v)
}
