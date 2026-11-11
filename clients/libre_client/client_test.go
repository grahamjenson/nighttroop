package libre_client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// fakeLibre is a LibreLinkUp server whose regions are path prefixes, e.g. /eu/llu/auth/login
type fakeLibre struct {
	*httptest.Server
	// login is the login response body for each region
	login    map[string]string
	requests []*http.Request
}

func newFakeLibre(t *testing.T, login map[string]string, graph string) *fakeLibre {
	f := &fakeLibre{login: login}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests = append(f.requests, r)
		region, path, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
		switch path {
		case "llu/auth/login":
			_, _ = w.Write([]byte(f.login[region]))
		case "llu/connections":
			_, _ = w.Write([]byte(`{"status":0,"data":[{"id":"c1","patientId":"patient-1"}]}`))
		case "llu/connections/patient-1/graph":
			_, _ = w.Write([]byte(graph))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

// client is an online client whose regions are the fake server's path prefixes
func (f *fakeLibre) client(t *testing.T, region string) *Client {
	c, err := NewClient("follower@example.com", "password", region)
	require.NoError(t, err)
	c.offline = false
	c.urlFor = func(region string) (string, error) {
		if _, err := regionURL(region); err != nil {
			return "", err
		}
		return f.URL + "/" + strings.ToLower(region), nil
	}
	c.baseURL, _ = c.urlFor(region)
	return c
}

func loginOK(userID string) string {
	return fmt.Sprintf(`{"status":0,"data":{"user":{"id":%q},"authTicket":{"token":"tok-1","expires":%d,"duration":15552000000}}}`, userID, time.Now().Add(time.Hour).Unix())
}

const graphJSON = `{"status":0,"data":{
	"connection":{"patientId":"patient-1","firstName":"John","glucoseMeasurement":{"FactoryTimestamp":"12/30/2022 4:47:40 PM","Timestamp":"12/30/2022 5:47:40 PM","type":1,"ValueInMgPerDl":115,"TrendArrow":4,"Value":115}},
	"graphData":[
		{"FactoryTimestamp":"12/30/2022 4:17:40 PM","Timestamp":"12/30/2022 5:17:40 PM","type":0,"ValueInMgPerDl":173,"Value":173,"TrendArrow":5},
		{"FactoryTimestamp":"12/30/2022 4:32:40 PM","Timestamp":"12/30/2022 5:32:40 PM","type":0,"ValueInMgPerDl":1500,"Value":1500},
		{"FactoryTimestamp":"not a time","Timestamp":"12/30/2022 5:32:40 PM","type":0,"ValueInMgPerDl":120,"Value":120}
	]}}`

func TestRegionURLOnlyAllowsKnownRegions(t *testing.T) {
	tests := []struct {
		region string
		url    string
	}{
		{consts.LibreRegionGlobal, "https://api.libreview.io"},
		{"EU", "https://api-eu.libreview.io"},
		{"EU2", "https://api-eu2.libreview.io"},
		{"US", "https://api-us.libreview.io"},
		{"evil.example.com", ""},
		{"eu", ""},
		{"", ""},
		{"OUS", ""},
	}
	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			got, err := regionURL(tt.region)
			if tt.url == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.url, got)
		})
	}
	_, err := NewClient("a@example.com", "p", "evil.example.com")
	require.Error(t, err)
}

func TestFetchReadingsLogsInAndParsesTheGraph(t *testing.T) {
	f := newFakeLibre(t, map[string]string{"eu": loginOK("user-1")}, graphJSON)
	c := f.client(t, "EU")

	readings, err := c.FetchReadings(context.Background())
	require.NoError(t, err)

	// The out of range and badly timed graph readings are left out; the latest is last and keeps its trend
	require.Len(t, readings, 2)
	require.Equal(t, 173.0, readings[0].Mgdl)
	require.Nil(t, readings[0].TrendArrow, "graph readings have no trend")
	require.Equal(t, time.Date(2022, 12, 30, 16, 17, 40, 0, time.UTC), readings[0].FactoryTime)
	require.Equal(t, int64(60), readings[0].UTCOffsetMinutes)
	require.Equal(t, 115.0, readings[1].Mgdl)
	require.Equal(t, 4, *readings[1].TrendArrow)

	require.Len(t, f.requests, 3)
	login := f.requests[0]
	require.Equal(t, http.MethodPost, login.Method)
	require.Equal(t, "llu.android", login.Header.Get("product"))
	require.NotEmpty(t, login.Header.Get("version"))
	require.Empty(t, login.Header.Get("authorization"))
	sum := sha256.Sum256([]byte("user-1"))
	for _, r := range f.requests[1:] {
		require.Equal(t, "Bearer tok-1", r.Header.Get("authorization"))
		require.Equal(t, hex.EncodeToString(sum[:]), r.Header.Get("account-id"))
	}

	// The login is kept for the next sync
	_, err = c.FetchReadings(context.Background())
	require.NoError(t, err)
	require.Len(t, f.requests, 5)
	require.Equal(t, "/eu/llu/connections", f.requests[3].URL.Path)
}

func TestLoginFollowsRedirectToAKnownRegionOnly(t *testing.T) {
	f := newFakeLibre(t, map[string]string{
		"global": `{"status":0,"data":{"redirect":true,"region":"eu"}}`,
		"eu":     loginOK("user-1"),
	}, graphJSON)
	c := f.client(t, consts.LibreRegionGlobal)
	require.NoError(t, c.Login(context.Background()))
	require.Equal(t, f.URL+"/eu", c.baseURL)
	require.Equal(t, "tok-1", c.token)

	// A region that is not in consts.LibreRegions is refused, so a response cannot pick the host
	f.login["global"] = `{"status":0,"data":{"redirect":true,"region":"evil.example.com/"}}`
	c = f.client(t, consts.LibreRegionGlobal)
	err := c.Login(context.Background())
	require.ErrorContains(t, err, "unknown LibreLinkUp region")
	require.Empty(t, c.token)

	// Redirects do not bounce between regions
	f.login["eu"] = `{"status":0,"data":{"redirect":true,"region":"us"}}`
	f.login["us"] = `{"status":0,"data":{"redirect":true,"region":"eu"}}`
	c = f.client(t, "EU")
	require.ErrorContains(t, c.Login(context.Background()), "redirected more than once")
}

func TestLoginErrors(t *testing.T) {
	tests := []struct {
		name  string
		login string
		want  string
	}{
		{"terms of use", `{"status":4,"data":{"step":{"type":"tou","componentName":"AcceptDocument"}}}`, "accept the terms of use in the LibreLinkUp app"},
		{"privacy policy", `{"status":4,"data":{"step":{"type":"pp"}}}`, "accept the privacy policy in the LibreLinkUp app"},
		{"status 4 without a step", `{"status":4,"data":{}}`, "accept the terms of use in the LibreLinkUp app"},
		{"bad password", `{"status":2,"error":{"message":"notAuthenticated"}}`, "invalid email or password"},
		{"no token", `{"status":0,"data":{"user":{"id":"u"}}}`, "invalid email or password"},
		{"not json", `<html>`, "error decoding libre response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeLibre(t, map[string]string{"eu": tt.login}, graphJSON)
			_, err := f.client(t, "EU").FetchReadings(context.Background())
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestFetchReadingsRefusesBadResponses(t *testing.T) {
	reading := `{"FactoryTimestamp":"12/30/2022 4:17:40 PM","Timestamp":"12/30/2022 5:17:40 PM","ValueInMgPerDl":100}`
	tooMany := `{"status":0,"data":{"graphData":[` + strings.Repeat(reading+",", maxReadings) + reading + `]}}`
	f := newFakeLibre(t, map[string]string{"eu": loginOK("u")}, tooMany)
	_, err := f.client(t, "EU").FetchReadings(context.Background())
	require.ErrorContains(t, err, "more than 1000")

	tooBig, err := json.Marshal(map[string]string{"padding": strings.Repeat("a", maxResponseBytes)})
	require.NoError(t, err)
	f = newFakeLibre(t, map[string]string{"eu": loginOK("u")}, string(tooBig))
	_, err = f.client(t, "EU").FetchReadings(context.Background())
	require.ErrorContains(t, err, "larger than")
}

func TestUnauthorizedLogsInAgainNextTime(t *testing.T) {
	logins := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/llu/auth/login" {
			logins++
			_, _ = w.Write([]byte(loginOK("u")))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("secret body"))
	}))
	defer srv.Close()
	c, err := NewClient("a@example.com", "p", "EU")
	require.NoError(t, err)
	c.offline, c.baseURL = false, srv.URL

	_, err = c.FetchReadings(context.Background())
	require.ErrorContains(t, err, "status=401")
	require.NotContains(t, err.Error(), "secret body")
	_, err = c.FetchReadings(context.Background())
	require.Error(t, err)
	require.Equal(t, 2, logins)
}

// A failed request is a sync error, stored, logged and sent to Sentry, so it never has the followed patient's ID in its
// path
func TestFetchReadingsErrorsHideThePatient(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, 0} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/llu/auth/login":
				_, _ = w.Write([]byte(loginOK("u")))
			case "/llu/connections":
				_, _ = w.Write([]byte(`{"status":0,"data":[{"patientId":"patient-1"}]}`))
			default:
				if status == 0 {
					<-r.Context().Done() // times out
					return
				}
				w.WriteHeader(status)
			}
		}))
		c, err := NewClient("a@example.com", "p", "EU")
		require.NoError(t, err)
		c.offline, c.baseURL = false, srv.URL
		c.httpClient.Timeout = 100 * time.Millisecond

		_, err = c.FetchReadings(context.Background())
		srv.Close()
		require.Error(t, err)
		require.NotContains(t, err.Error(), "patient-1")
		require.NotContains(t, err.Error(), srv.URL)
	}
}

func TestFetchReadingsRefusesRedirect(t *testing.T) {
	secretHits := 0
	secret := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { secretHits++ }))
	defer secret.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, secret.URL+"/latest/meta-data", http.StatusFound)
	}))
	defer redirector.Close()
	c, err := NewClient("a@example.com", "p", "EU")
	require.NoError(t, err)
	c.offline, c.baseURL = false, redirector.URL

	_, err = c.FetchReadings(context.Background())
	require.ErrorContains(t, err, "302")
	require.NotContains(t, err.Error(), secret.URL)
	require.Equal(t, 0, secretHits)
}

func TestOfflineLoginUnderGoTest(t *testing.T) {
	c, err := NewClient("someone@example.com", "p", "EU")
	require.NoError(t, err)
	_, err = c.FetchReadings(context.Background())
	require.ErrorContains(t, err, "libre authentication failed")

	c, err = NewClient(TestValidEmail, "p", "EU")
	require.NoError(t, err)
	readings, err := c.FetchReadings(context.Background())
	require.NoError(t, err)
	require.Empty(t, readings)
}
