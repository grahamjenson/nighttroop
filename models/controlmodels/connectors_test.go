package controlmodels

import (
	"encoding/json"
	"strings"
	"testing"

	"nighttroop.t1d.nz/data/sql/controldb"
)

// The control API shows a connector's reports, never its encrypted credentials
func TestConnectorFromDBLeavesOutCredentials(t *testing.T) {
	url, password := "secret-url-enc", "secret-password-enc"
	source := ConnectorFromDB(&controldb.SiteConnector{SiteID: "site", Type: "nightscout", UrlOrUsernameEnc: &url, PasswordEnc: &password})

	body, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{url, password, "_enc"} {
		if strings.Contains(string(body), secret) {
			t.Errorf("connector JSON has %q: %s", secret, body)
		}
	}
	if source.ID() != "site::nightscout" {
		t.Errorf("ID is %q", source.ID())
	}
}
