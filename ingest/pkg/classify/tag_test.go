package classify

import (
	"testing"
)

func TestKeywords(t *testing.T) {
	for k := range keywords {
		if sk := defaultStemmer(k); sk != k {
			if k == "licens" {
				// stemming the stemmd word of license will be licen (not applicable)
				continue
			}
			t.Errorf("keywords shoudl be stemmed, expecting '%s', got '%s'", sk, k)
		}
	}
}
