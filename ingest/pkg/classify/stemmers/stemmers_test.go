//+build !no_dep_test

// Package stemmers provides test and benchmarks for go stemmer packages.
package stemmers

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	agonopol "github.com/agonopol/go-stem"
	caneroj1 "github.com/caneroj1/stemmer"
	dchest "github.com/dchest/stemmer/porter2"
	kljensen "github.com/kljensen/snowball/english"
	reiver "github.com/reiver/go-porterstemmer"
	"github.com/repofuel/repofuel/ingest/pkg/classify"
	surgebase "github.com/surgebase/porter2"
)

const commitMsg = `test(language-service): Remove unused code in test project (#37122)

This commit removes the bootstrap() function in the test project since
its presence has no effect on the behavior of language service.

Also removes the explicit cast when instantiating CounterDirectiveContext,
and let type inference takes care of that.

PR Close #37122`

type stemAdapter func(string) string

func (s2 stemAdapter) Stem(s string) string {
	return s2(s)
}

type stemAdapterBool func(string, bool) string

func (s2 stemAdapterBool) Stem(s string) string {
	return s2(s, true)
}

type stemAdapterToLower func(string) string

func (s2 stemAdapterToLower) Stem(s string) string {
	return strings.ToLower(s2(s))
}

type byteStemAdapter func([]byte) []byte

func (s2 byteStemAdapter) Stem(s string) string {
	return string(s2([]byte(s)))
}

type runeStemAdapter func([]rune) []rune

func (s2 runeStemAdapter) Stem(s string) string {
	return string(s2([]rune(s)))
}

func noStem(s string) string {
	return s
}

func BenchmarkStemmerMatch(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(stemAdapter(noStem).Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchDchest(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(dchest.Stemmer.Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchKljensen(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(stemAdapterBool(kljensen.Stem).Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchSurgebase(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(stemAdapter(surgebase.Stem).Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchCaneroj1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(stemAdapter(caneroj1.Stem).Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchAgonopol(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(byteStemAdapter(agonopol.Stem).Stem, commitMsg)
	}
}

func BenchmarkStemmerMatchReiver(b *testing.B) {
	for n := 0; n < b.N; n++ {
		classify.StemMatch(runeStemAdapter(reiver.Stem).Stem, commitMsg)
	}
}

func BenchmarkStemmerOneWordDchest(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dchest.Stemmer.Stem("Classification")
	}
}

func BenchmarkStemmerOneWordKljensen(b *testing.B) {
	for n := 0; n < b.N; n++ {
		kljensen.Stem("Classification", true)
	}
}

func BenchmarkStemmerOneWordSurgebase(b *testing.B) {
	for n := 0; n < b.N; n++ {
		surgebase.Stem("Classification")
	}
}

func BenchmarkStemmerOneWordCaneroj1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		caneroj1.Stem("Classification")
	}
}

func BenchmarkStemmerOneWordAgonopol(b *testing.B) {
	for n := 0; n < b.N; n++ {
		agonopol.Stem([]byte("Classification"))
	}
}

func BenchmarkStemmerOneWordReiver(b *testing.B) {
	for n := 0; n < b.N; n++ {
		reiver.Stem([]rune("Classification"))
	}
}

func TestStemmerDchest(t *testing.T) {
	testStemmer(t, dchest.Stemmer.Stem, "porter2")
}

func TestStemmerSurgebase(t *testing.T) {
	testStemmer(t, stemAdapter(surgebase.Stem).Stem, "porter2")
}

func TestStemmerKljensen(t *testing.T) {
	t.Skip("we know that github.com/kljensen/snowball/english fails on  \"'''\" expected \"'\" got \"\"")
	testStemmer(t, stemAdapterBool(kljensen.Stem).Stem, "porter2")
}

func TestStemmerCaneroj1(t *testing.T) {
	t.Skip("we know that github.com/caneroj1/stemmer fails on the porter dataset")
	testStemmer(t, stemAdapterToLower(caneroj1.Stem).Stem, "porter")
}

func TestStemmerAgonopol(t *testing.T) {
	testStemmer(t, byteStemAdapter(agonopol.Stem).Stem, "porter")
}

func TestStemmerReiver(t *testing.T) {
	testStemmer(t, runeStemAdapter(reiver.Stem).Stem, "porter")
}

func testStemmer(t *testing.T, stem classify.Stemmer, testDataPrefix string) {
	vocFile, err := os.Open(testDataPrefix + "_voc.txt")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	defer vocFile.Close()

	outputFile, err := os.Open(testDataPrefix + "_output.txt")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	defer outputFile.Close()

	bvoc := bufio.NewReader(vocFile)
	bout := bufio.NewReader(outputFile)

	for {
		voc, err := bvoc.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			return
		}

		output, err := bout.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			return
		}

		voc = voc[:len(voc)-1]
		output = output[:len(output)-1]
		str := stem(voc)
		if str != output {
			t.Errorf(`"%s" expected %q got %q`, voc, output, str)
		}
	}
}
