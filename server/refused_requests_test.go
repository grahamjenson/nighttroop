package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/testutils"
)

// Requests for a site that does not exist, as from a sweep of subdomains, and to a site without a token, are still
// refused, but are not each kept in the logs DB, which keeps only its newest rows: each minute's are summed in one log
// for each reason, with the subdomains or routes refused most
func TestRefusedRequestsAreSummed(t *testing.T) {
	config, controlDB, site := testutils.NewControlDBWithSite(t)
	handler := GetMainHandler(controlDB, config)
	buffer, before := &aslog.Buffer{Full: make(chan struct{}, 1)}, slog.Default()
	slog.SetDefault(slog.New(aslog.ContextHandler{Handler: aslog.BufferHandler{Buffer: buffer}}))
	t.Cleanup(func() { slog.SetDefault(before) })
	aslog.RefusedRequests.Flush(t.Context())
	buffer.Take()

	send := func(url string) int {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}
	for range 3 {
		require.Equal(t, http.StatusGone, send("http://zzzzzzzz.ns.localhost/api/v1/entries.json"))
	}
	require.Equal(t, http.StatusNotFound, send("http://zzzzzzzz.localhost/"))
	require.Equal(t, http.StatusUnauthorized, send("http://"+site.SiteDB.Subdomain+".ns.localhost/api/v1/entries.json"))

	records, _ := buffer.Take()
	for _, record := range records {
		require.NotEqual(t, "Main router site does not exist", record.Message)
		require.NotEqual(t, "JWT and Legacy Token Auth Failed for Nightscout API", record.Message)
	}

	aslog.RefusedRequests.Flush(t.Context())
	records, _ = buffer.Take()
	summaries := map[string]string{}
	for _, record := range records {
		require.Equal(t, "Unauthenticated requests refused", record.Message)
		require.Equal(t, "info", record.Level)
		var attributes struct {
			Reason string `json:"reason"`
			Status int    `json:"status"`
		}
		require.NoError(t, json.Unmarshal([]byte(record.Attributes), &attributes))
		summaries[attributes.Reason+" "+http.StatusText(attributes.Status)] = record.Attributes
	}
	require.Len(t, summaries, 3)
	require.JSONEq(t, `{"reason":"Main router site does not exist","status":410,"total":3,"keys":1,"other":0,
		"top":[{"key":"zzzzzzzz","count":3}]}`, summaries["Main router site does not exist Gone"])
	require.JSONEq(t, `{"reason":"Main router site does not exist","status":404,"total":1,"keys":1,"other":0,
		"top":[{"key":"zzzzzzzz","count":1}]}`, summaries["Main router site does not exist Not Found"])
	require.JSONEq(t, `{"reason":"JWT and Legacy Token Auth Failed for Nightscout API","status":401,"total":1,"keys":1,
		"other":0,"top":[{"key":"/api/v1/*","count":1}]}`, summaries["JWT and Legacy Token Auth Failed for Nightscout API Unauthorized"])
}
