// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a written permission.

package commitmsg

import (
	"encoding/json"
	"errors"
)

type Category uint8

const (
	Others Category = iota + 1
	Corrective
	FeatureAddition
	NonFunctional
	Perfective
	Tests
	License
	Merge
	NoSourceCode
)

func (c Category) String() string {
	switch c {
	case Corrective:
		return "Corrective"
	case FeatureAddition:
		return "Feature Addition"
	case NonFunctional:
		return "Non Functional"
	case Perfective:
		return "Perfective"
	case Merge:
		return "Merge"
	case Tests:
		return "Tests"
	case License:
		return "License"
	case NoSourceCode:
		return "No Source Code"
	case Others:
		return "General"
	default:
		return "Unknown"
	}
}

func (c Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

var categories map[string]Category

func init() {
	const start = Others
	const end = NoSourceCode
	categories = make(map[string]Category, int(end-start)+1)

	for i := start; i <= end; i++ {
		categories[i.String()] = i
	}
}

func (c *Category) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	var ok bool
	if *c, ok = categories[str]; !ok {
		return errors.New("unknown category")
	}

	return nil
}

func (c Category) IsFix() bool {
	if c == Corrective {
		return true
	}
	return false
}
