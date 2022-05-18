package classify

import (
	"testing"
)

const sampleCommitMessage = `feat(launcher): Set default Firefox prefs (puppeteer#5149) (puppeteer#5195)

Add recommended automation preferences to profile
setup when launching Firefox. This profile can be overridden
by using the 'userDataDir' launch option, or individual prefs
can be overwritten with the 'extraPrefsFirefox' option.

The preferences have been reviewed by peers at Mozilla
over at https://bugzilla.mozilla.org/show_bug.cgi?id=1596888

Co-Authored-By: Mathias Bynens <mathias@qiwi.be>`

func TestConventionalCommits(t *testing.T) {
	tag := ConventionalCommits(sampleCommitMessage)
	if tag != Feature {
		t.Errorf("expected tag: %s, got: %s", Feature, tag)
	}
}
