package classify

import (
	"fmt"
	"io"
	"strconv"
)

type Tag uint8

//go:generate stringer -type=Tag -linecomment
//go:generate jsonenums -type=Tag
const (
	Code     Tag = iota + 1
	NoneCode     // None Code
	Fix
	Bug
	Add
	Update
	Feature
	Tests
	Documentations
	Refactor // Code Refactoring
	License
	Build         // Build System
	CI            // Continuous Integration
	TechnicalDebt // Technical Debt
	Style
	Release
	Dependencies
	GeneratedCode           // Generated Code
	PerformanceImprovements // Performance Improvements
	Reverts
	MiscellaneousChores // Miscellaneous Chores
)

// keywords is not defined for `Tests`, `Documentations`, `Build`, `CI`.
var keywords = map[string]Tag{
	"fix":     Fix,
	"bug":     Fix,
	"wrong":   Fix,
	"fail":    Fix,
	"problem": Fix,
	"correct": Fix,
	"corrig":  Fix, // corriger
	"resolv":  Fix, // resolve
	"faux":    Fix, // French
	"faut":    Fix, // French: faute
	"echou":   Fix, // French
	"échou":   Fix, // French
	"résol":   Fix, // French
	"résou":   Fix, // French
	"résolu":  Fix, // French
	"résoudr": Fix, // French: résoudre
	"problèm": Fix, // French

	"test":      Tests,
	"junit":     Tests,
	"coverag":   Tests, // coverage
	"assert":    Tests,
	"couvertur": Tests, // French: couverture
	"assur":     Tests, // French

	"add":       Add,
	"updat":     Update,   // update
	"featur":    Feature,  // feature
	"improv":    Update,   // improve
	"chang":     Update,   // change
	"renam":     Refactor, // rename
	"move":      Refactor,
	"refactor":  Refactor,
	"licens":    License, // license
	"hack":      TechnicalDebt,
	"todo":      TechnicalDebt,
	"style":     Style,
	"format":    Style,
	"indent":    Style,
	"changelog": Release,
	"depend":    Dependencies, // dependencies
	"librari":   Dependencies, // libraries
}

func (t *Tag) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	tag, ok := _TagNameToValue[s]
	if !ok {
		return fmt.Errorf("invalid Tag %q", s)
	}
	*t = tag
	return nil
}

func (t Tag) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(t.String()))
}

var nothing void

type void struct{}

type CategoriesSet map[Tag]void

func NewCategoriesSet() CategoriesSet {
	return make(CategoriesSet)
}

func (s CategoriesSet) Add(h Tag) {
	s[h] = nothing
}

func (s CategoriesSet) Slice() []Tag {
	slice := make([]Tag, len(s))
	var i int
	for k := range s {
		slice[i] = k
		i += 1
	}
	return slice
}

func TagFromString(s string) (Tag, error) {
	t, ok := _TagNameToValue[s]
	if !ok {
		return 0, fmt.Errorf("invalid Tag %q", s)
	}

	return t, nil
}

type FileType uint8

//go:generate stringer -type=FileType -linecomment -trimprefix=File
//go:generate jsonenums -type=FileType

const (
	FileBinary FileType = iota + 1
	FileConfiguration
	FileDocumentation
	FileCode
	FileGenerated
	FileDependency
	FileTests
	FileSymlink
)

func (t FileType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(t.String()))
}
