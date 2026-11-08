package flow

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/pquerna/otp/totp"
)

// Render fills in a step's {{ }} values from the context, with helpers like site, token, dig and server_url
func Render(input string, ctx map[string]any) (string, error) {
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				dug := Dig(ctx, path)
				switch dug.(type) {
				case nil:
					return "no_value"
				case string:
					return dug.(string)
				case float64:
					return fmt.Sprintf("%.0f", dug.(float64))
				case int:
					return dug.(int)
				case bool:
					return dug.(bool)
				default:
					return dug
				}
			},
			"site": func() string {
				if v, ok := ctx["site"]; ok && v != nil {
					return ctx["site"].(string)
				}
				return ctx["default_site"].(string)
			},
			"token": func() string {
				if v, ok := ctx["token"]; ok && v != nil {
					return ctx["token"].(string)
				}
				return ctx["default_token"].(string)
			},
			"read_token": func() string {
				return ctx["read_token"].(string)
			},
			"jwt": func() string {
				if v, ok := ctx["jwt"]; ok && v != nil {
					return ctx["jwt"].(string)
				}
				return ctx["default_jwt"].(string)
			},
			"read_jwt": func() string {
				return ctx["read_jwt"].(string)
			},
			"server_url": func() string {
				return ctx["server_url"].(string)
			},
			// site_url is a site's address on the server, from the context's scheme and host, e.g. site_url "abcdefgh"
			// for http://abcdefgh.localhost:1234 in a test, or https://abcdefgh.staging.nighttroop.nz on staging
			"site_url": func(subdomain string) string {
				return fmt.Sprintf("%s://%s.%s", ctx["scheme"], subdomain, ctx["host"])
			},
			// totp is the code an authenticator app shows now for a TOTP secret, as a new site's totp_secret
			"totp": func(secret string) (string, error) {
				return totp.GenerateCode(secret, time.Now())
			},
			// env is an environment variable, e.g. a password given to the script, and an error when it is not set
			"env": func(key string) (string, error) {
				value, ok := os.LookupEnv(key)
				if !ok {
					return "", fmt.Errorf("%s is not set", key)
				}
				return value, nil
			},
			// now is the time in a timezone, e.g. {{ now "Pacific/Auckland" "2006-01-02 15:" }}
			"now": func(zone string, layout string) (string, error) {
				location, err := time.LoadLocation(zone)
				if err != nil {
					return "", err
				}
				return time.Now().In(location).Format(layout), nil
			},
			// replace is s with each old replaced by new, e.g. {{ replace ".localhost" ".ns.localhost" .site_url }}
			// repeat is s count times, e.g. for a value longer than a field keeps
			"repeat": func(s string, count int) string { return strings.Repeat(s, count) },
			"replace": func(old, new, s string) string {
				return strings.ReplaceAll(s, old, new)
			},
			"time":  templateTime,
			"itime": templateITime,
			// utctime is offset minutes ago in UTC with no zone, as LoopFollow sends, e.g. utctime 60 for an hour ago
			"utctime": func(offset int) string {
				return time.Now().UTC().Add(-time.Duration(offset) * time.Minute).Format("2006-01-02T15:04:05")
			},
			// javatime is minutes ago as Java's yyyy-MM-dd'T'HH:mm:ss.SSSZ at +1300, as xDrip+ sends dateString
			"javatime": func(offset int) string {
				at := time.Now().Add(-time.Duration(offset) * time.Minute)
				return at.In(time.FixedZone("", 13*60*60)).Format("2006-01-02T15:04:05.000-0700")
			},
			// sha1 is the SHA1 in hex, as a Nightscout client sends its API secret, e.g. sha1 token
			"sha1": func(s string) string {
				sum := sha1.Sum([]byte(s))
				return hex.EncodeToString(sum[:])
			},
			"date":    func(days int) string { return templateDayTime(days, 0).Format(time.DateOnly) },
			"daytime": func(days, minutes int) string { return templateDayTime(days, minutes).Format(time.RFC3339) },
			"dayms":   func(days, minutes int) int64 { return templateDayTime(days, minutes).UnixMilli() },
			// addms is a dug time in milliseconds plus ms, e.g. addms (dig "last.result.collections.entries") 1
			"addms": func(dug any, ms int64) (int64, error) {
				n, err := strconv.ParseInt(fmt.Sprint(dug), 10, 64)
				return n + ms, err
			},
			"tokenByName": func(list []any, name string) string {
				for _, item := range list {
					m := item.(map[string]any)
					if m["name"] == name {
						return m["accessToken"].(string)
					}
				}
				return ""
			},
		}).
		Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func templateTime(offset int) string {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.Format(time.RFC3339)
}

// templateDayTime is minutes after midnight, in UTC, the day days from today, e.g. daytime -1 720 for noon yesterday
func templateDayTime(days, minutes int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, days).Add(time.Duration(minutes) * time.Minute)
}

func templateITime(offset int) int64 {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.UnixMilli()
}

// Dig is the value at a dotted path in root, through maps and list indexes, e.g. "entries.0.sgv", or nil
func Dig(root any, path string) any {

	parts := strings.Split(path, ".")
	var cur = root

	for _, p := range parts {
		switch v := cur.(type) {

		case map[string]any:
			next, ok := v[p]
			if !ok {
				return nil
			}
			cur = next

		case []any:
			// allow numeric indexing into slices
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(v) {
				return nil
			}
			cur = v[i]

		default:
			return nil
		}
	}

	return cur
}
