package commitmsg

import (
	"regexp"
	"strings"
	"testing"
)

var keywords = correctiveKeywords

const commitMsg = `test(language-service): Remove unused code in test project (#37122)

This commit removes the bootstrap() function in the test project since
its presence has no effect on the behavior of language service.

Also removes the explicit cast when instantiating CounterDirectiveContext,
and let type inference takes care of that.

PR Close #37122`

var regExp = regexp.MustCompile(strings.Join(keywords, "|"))

var regExpSlice []*regexp.Regexp

func init() {
	regExpSlice = make([]*regexp.Regexp, 0, len(keywords))
	for _, s := range keywords {
		regExpSlice = append(regExpSlice, regexp.MustCompile(s))
	}
}

func BenchmarkRegexpMatchBulk(b *testing.B) {
	s := strings.ToLower(commitMsg)
	for n := 0; n < b.N; n++ {
		regExp.MatchString(s)
	}
}

func BenchmarkRegexpMatchStringSlice(b *testing.B) {
	s := strings.ToLower(commitMsg)
	for n := 0; n < b.N; n++ {
		for _, reg := range regExpSlice {
			reg.MatchString(s)
		}
	}
}

func BenchmarkStringsContains(b *testing.B) {
	s := strings.ToLower(commitMsg)
	for n := 0; n < b.N; n++ {
		for _, k := range keywords {
			strings.Contains(s, k)
		}
	}
}
