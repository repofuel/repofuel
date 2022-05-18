package insights

//go:generate stringer -type=Reason -linecomment
//go:generate jsonenums -type=Reason
type Reason int

const (
	DevFirstCommit Reason = iota + 1
	ManyDevelopers
	FirstModification
	CommitLargeFiles
	DevNewToRepository
	DevNewToSubsystem
	DevIsLastModifier
	DevExpert
	FilesDevNew
	FilesDevExpert
	FilesFrequentlyChanged
	FilesFrequentlyFixed
	FilesFrequentlyBuggy
	FilesAbandoned
	FilesRecentlyModified
	CommitScattered
	CommitIsLargest
	CommitIsLargestAdd
	CommitIsLargestDelete
)

type description struct {
	Icon        *string `json:"icon"`
	Name        string
	Description string `json:"description"`
	Action      string
	Color       string
	Details     string
}


var catalog = map[Reason]*description{
	DevFirstCommit: {
		Description: "This is the first commit by this developer in this repository",
	},
	ManyDevelopers: {
		Description: "This code has been touched by many developers in the past",
	},
	FirstModification: {
		Description: "This is the first time this developers changes these files",
	},
	CommitLargeFiles: {
		Description: "The files modified in this commit are very large",
	},
	DevNewToRepository: {
		Description: "The commit author is new to this project",
	},
	DevNewToSubsystem: {
		Description: "The commit author is new to this component",
	},
	DevIsLastModifier: {
		Name:        "Last modifier",
		Description: "The commit author was the last developer to modify the file",
		Color:       "green.5",
	},
	DevExpert: {
		Description: "The commit author is an experienced developer on this project",
	},
	FilesDevNew: {
		Name:        "First modification",
		Description: "The commit author is new to the file",
		Color:       "red.5",
	},
	FilesDevExpert: {
		Name:        "Expert",
		Description: "The commit author is expert in the file",
		Color:       "green.5",
	},
	FilesFrequentlyChanged: {
		Name:        "Modified frequently",
		Description: "The file modified frequently",
		Color:       "green.5",
	},
	FilesFrequentlyFixed: {
		Name:        "Fixed frequently",
		Description: "The file get fixed frequently",
		Color:       "red.5",
	},
	FilesFrequentlyBuggy: {
		Name:        "Buggy",
		Description: "The file is buggy",
		Color:       "red.5",
	},
	FilesAbandoned: {
		Name:        "Abandoned",
		Description: "The file has not changed for a long time",
		Color:       "red.5",
	},
	FilesRecentlyModified: {
		Name:        "Recently changed",
		Description: "The file changed recently",
		Color:       "green.5",
	},
	CommitScattered: {
		Description: "This commit is highly scattered",
	},
	CommitIsLargest: {
		Description: "This commit is larger than usual for this project",
	},
	CommitIsLargestAdd: {
		Description: "This commit adds a large amount of code",
	},
	CommitIsLargestDelete: {
		Description: "This commit deletes a large amount of code",
	},
}

func (r Reason) Description() string {
	v, ok := catalog[r]
	if ok {
		return v.Description
	}

	return r.String()
}

func (r Reason) Icon() *string {
	v, ok := catalog[r]
	if ok {
		return v.Icon
	}
	return nil
}

func (r Reason) Name() string {
	v, ok := catalog[r]
	if ok && v.Name != "" {
		return v.Name
	}

	return r.String()
}

func (r Reason) Color() *string {
	v, ok := catalog[r]
	if ok && v.Color != "" {
		return &v.Color
	}

	return nil
}
