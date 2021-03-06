// Code generated by jsonenums -type=Stage; DO NOT EDIT.

package status

import (
	"encoding/json"
	"fmt"
)

var (
	_StageNameToValue = map[string]Stage{
		"Queued":      Queued,
		"Added":       Added,
		"Cloning":     Cloning,
		"Cloned":      Cloned,
		"Fetching":    Fetching,
		"Fetched":     Fetched,
		"Ingesting":   Ingesting,
		"Ingested":    Ingested,
		"Analyzing":   Analyzing,
		"Analyzed":    Analyzed,
		"Predicting":  Predicting,
		"Predicted":   Predicted,
		"Ready":       Ready,
		"Failed":      Failed,
		"Canceled":    Canceled,
		"Watched":     Watched,
		"Recovered":   Recovered,
		"Ignored":     Ignored,
		"Progressing": Progressing,
	}

	_StageValueToName = map[Stage]string{
		Queued:      "Queued",
		Added:       "Added",
		Cloning:     "Cloning",
		Cloned:      "Cloned",
		Fetching:    "Fetching",
		Fetched:     "Fetched",
		Ingesting:   "Ingesting",
		Ingested:    "Ingested",
		Analyzing:   "Analyzing",
		Analyzed:    "Analyzed",
		Predicting:  "Predicting",
		Predicted:   "Predicted",
		Ready:       "Ready",
		Failed:      "Failed",
		Canceled:    "Canceled",
		Watched:     "Watched",
		Recovered:   "Recovered",
		Ignored:     "Ignored",
		Progressing: "Progressing",
	}
)

func init() {
	var v Stage
	if _, ok := interface{}(v).(fmt.Stringer); ok {
		_StageNameToValue = map[string]Stage{
			interface{}(Queued).(fmt.Stringer).String():      Queued,
			interface{}(Added).(fmt.Stringer).String():       Added,
			interface{}(Cloning).(fmt.Stringer).String():     Cloning,
			interface{}(Cloned).(fmt.Stringer).String():      Cloned,
			interface{}(Fetching).(fmt.Stringer).String():    Fetching,
			interface{}(Fetched).(fmt.Stringer).String():     Fetched,
			interface{}(Ingesting).(fmt.Stringer).String():   Ingesting,
			interface{}(Ingested).(fmt.Stringer).String():    Ingested,
			interface{}(Analyzing).(fmt.Stringer).String():   Analyzing,
			interface{}(Analyzed).(fmt.Stringer).String():    Analyzed,
			interface{}(Predicting).(fmt.Stringer).String():  Predicting,
			interface{}(Predicted).(fmt.Stringer).String():   Predicted,
			interface{}(Ready).(fmt.Stringer).String():       Ready,
			interface{}(Failed).(fmt.Stringer).String():      Failed,
			interface{}(Canceled).(fmt.Stringer).String():    Canceled,
			interface{}(Watched).(fmt.Stringer).String():     Watched,
			interface{}(Recovered).(fmt.Stringer).String():   Recovered,
			interface{}(Ignored).(fmt.Stringer).String():     Ignored,
			interface{}(Progressing).(fmt.Stringer).String(): Progressing,
		}
	}
}

// MarshalJSON is generated so Stage satisfies json.Marshaler.
func (r Stage) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(r).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	s, ok := _StageValueToName[r]
	if !ok {
		return nil, fmt.Errorf("invalid Stage: %d", r)
	}
	return json.Marshal(s)
}

// UnmarshalJSON is generated so Stage satisfies json.Unmarshaler.
func (r *Stage) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Stage should be a string, got %s", data)
	}
	v, ok := _StageNameToValue[s]
	if !ok {
		return fmt.Errorf("invalid Stage %q", s)
	}
	*r = v
	return nil
}
