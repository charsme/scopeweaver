package selfupdate

import (
	"strings"
	"testing"
)

// Self-update downloads and executes code, so it must only follow releases cut from
// the reviewed fork. Merging upstream can silently restore the upstream source; this
// fails that merge instead of letting unreviewed releases reach the pilot host.
func TestUpdateSourceIsReviewedFork(t *testing.T) {
	if Repo != "charsme/scopeweaver" {
		t.Fatalf("Repo=%q; self-update must point at the reviewed fork charsme/scopeweaver", Repo)
	}
	if !strings.Contains(latestURL, "/repos/charsme/scopeweaver/") {
		t.Fatalf("latestURL=%q does not query the reviewed fork", latestURL)
	}
}
