// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a written permission.

package classify

import (
	"regexp"
	"strings"

	"github.com/surgebase/porter2"
)

type Stemmer func(s string) string

var defaultStemmer Stemmer = porter2.Stem

func StemMatchKeywords(keywords map[string]Tag, stem Stemmer, s string) CategoriesSet {
	categories := NewCategoriesSet()
	for _, s := range strings.Fields(s) {
		s = stem(strings.Trim(s, ":.`\"'"))
		if c, ok := keywords[s]; ok {
			categories.Add(c)
		}
	}
	return categories
}

func StemMatch(stemmer Stemmer, s string) CategoriesSet {
	return StemMatchKeywords(keywords, stemmer, s)
}

func FindCategories(s string) CategoriesSet {
	categories := StemMatch(defaultStemmer, s)
	tag := ConventionalCommits(s)
	if tag != 0 {
		categories.Add(tag)
	}
	return categories
}

var conventionalCommitsPattern = regexp.MustCompile("^(\\w*)(?:\\((.*)\\))?!?: (.*)$")
var conventionalCommitsTags = map[string]Tag{
	"feat":     Feature,
	"feature":  Feature,
	"fix":      Fix,
	"perf":     PerformanceImprovements,
	"revert":   Reverts,
	"docs":     Documentations,
	"style":    Style,
	"chore":    MiscellaneousChores,
	"refactor": Refactor,
	"test":     Tests,
	"build":    Build,
	"ci":       CI,
}

func ConventionalCommits(s string) Tag {
	firstLine := strings.SplitN(s, "\n", 2)[0]
	res := conventionalCommitsPattern.FindStringSubmatch(firstLine)
	if len(res) < 2 {
		return 0
	}

	return conventionalCommitsTags[res[1]]
}
