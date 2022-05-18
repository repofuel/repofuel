// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a written permission.

package commitmsg

import (
	"strings"
)

var (
	correctiveKeywords = []string{
		"fix",
		"bug",
		"wrong",
		"fail",
		"problem",
		"correct",
		"corriger",
		"resolve",
		"faux",
		"faute",
		"echou",
		"échou",
		"résol",
		"résou",
		"résolu",
		"résoudre",
		"problèm",
	}

	featureAdditionKeywords = []string{
		"new",
		"add",
		"requirement",
		"initial",
		"create",
		"introduce",
		"implement",
		"mettre",
		"ajout",
		"feature",
		"neuf",
		"nouveau",
		"introdu",
	}

	nonFunctionalKeywords = []string{
		"doc",
		"merge",
	}

	perfectiveKeywords = []string{
		"clean",
		"better",
		"suppression",
		"refactor",
		"ajustement",
		"renam",
		"renom",
		"menage",
		"ménage",
	}

	testsKeywords = []string{
		"test",
		"junit",
		"coverage",
		"assert",
		"couverture",
		"assur",
	}

	licenseKeywords = []string{
		"license",
	}
)

func GetCategoryByKeywords(message string) Category {
	message = strings.ToLower(message)
	switch {
	case IsCorrective(message):
		return Corrective
	case IsFeatureAddition(message):
		return FeatureAddition
	case IsNonFunctional(message):
		return NonFunctional
	case IsPerfective(message):
		return Perfective
	case IsTests(message):
		return Tests
	case IsLicense(message):
		return License
	default:
		return Others
	}
}

// deprecated
func IsCorrective(message string) bool {
	return HasKeyword(message, correctiveKeywords)
}

func IsFeatureAddition(message string) bool {
	return HasKeyword(message, featureAdditionKeywords)
}

func IsNonFunctional(message string) bool {
	return HasKeyword(message, nonFunctionalKeywords)
}

func IsPerfective(message string) bool {
	return HasKeyword(message, perfectiveKeywords)
}

func IsTests(message string) bool {
	return HasKeyword(message, testsKeywords)
}

func IsLicense(message string) bool {
	return HasKeyword(message, licenseKeywords)
}

// HasKeyword check if a string have any of the provided keywords.
// The string and the keywords should be in lower-case.
func HasKeyword(s string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
