package site_summary_templ

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// An identity that is not an admin cannot open the pages that say why readings stopped, so it is told to ask one, with
// no links. The admin's links are checked in tests/bdd/site/summary_stale.bdd.
func TestStaleHelpTellsAnyoneButAnAdminToAskOne(t *testing.T) {
	var html bytes.Buffer
	if err := staleHelp().Render(context.Background(), &html); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "Tell an admin of this site.") {
		t.Errorf("it does not say to tell an admin: %s", html.String())
	}
	if strings.Contains(html.String(), "<a ") {
		t.Errorf("it links to pages only an admin can open: %s", html.String())
	}
}
