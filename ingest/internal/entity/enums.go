package entity

import (
	"fmt"
	"io"
	"strconv"
)

type Frequency string

const (
	FrequencyDaily   Frequency = "DAILY"
	FrequencyMonthly Frequency = "MONTHLY"
	FrequencyYearly  Frequency = "YEARLY"
)

var AllFrequency = []Frequency{
	FrequencyDaily,
	FrequencyMonthly,
	FrequencyYearly,
}

func (e Frequency) IsValid() bool {
	switch e {
	case FrequencyDaily, FrequencyMonthly, FrequencyYearly:
		return true
	}
	return false
}

func (e Frequency) String() string {
	return string(e)
}

func (e *Frequency) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Frequency(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Frequency", str)
	}
	return nil
}

func (e Frequency) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type RepositoryAffiliation string

const (
	RepositoryAffiliationOwner        RepositoryAffiliation = "OWNER"
	RepositoryAffiliationCollaborator RepositoryAffiliation = "COLLABORATOR"
	RepositoryAffiliationMonitor      RepositoryAffiliation = "MONITOR"
	RepositoryAffiliationAccess       RepositoryAffiliation = "ACCESS"
)

var AllRepositoryAffiliation = []RepositoryAffiliation{
	RepositoryAffiliationOwner,
	RepositoryAffiliationCollaborator,
	RepositoryAffiliationMonitor,
	RepositoryAffiliationAccess,
}

func (e RepositoryAffiliation) IsValid() bool {
	switch e {
	case RepositoryAffiliationOwner, RepositoryAffiliationCollaborator, RepositoryAffiliationMonitor, RepositoryAffiliationAccess:
		return true
	}
	return false
}

func (e RepositoryAffiliation) String() string {
	return string(e)
}

func (e *RepositoryAffiliation) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RepositoryAffiliation(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RepositoryAffiliation", str)
	}
	return nil
}

func (e RepositoryAffiliation) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
