// Package geoip estimates where an IP address is from DB-IP's free Lite databases, licensed CC BY 4.0, which asks
// that pages showing their results link to https://db-ip.com (see footer.DBIPCredit).
//
// dbip-country-lite.mmdb is DB-IP's IP to Country Lite, as packaged by @ip-location-db/dbip-country-mmdb on npm
// (DB-IP's own download is https://db-ip.com/db/download/ip-to-country-lite). It is built in, so every server knows
// countries; the Docker image adds the latest IP to City Lite, for cities too, at GEOIP_DB_PATH.
package geoip

import (
	_ "embed"
	"log/slog"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

//go:embed dbip-country-lite.mmdb
var countryLite []byte

// Location is where an IP address is estimated to be, each part empty when unknown, as for a private address
type Location struct {
	// Country is the ISO 3166-1 alpha-2 code, e.g. NZ
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
}

// Locator finds IP addresses' locations
type Locator struct {
	db *maxminddb.Reader
}

// record reads both layouts: the built in country database's flat country_code, and DB-IP's own City Lite database's
// country and city, laid out as MaxMind's are
type record struct {
	CountryCode string `maxminddb:"country_code"`
	Country     struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
}

// New opens the database at path, or the built in country database when there is no path or it cannot be opened
func New(path string) *Locator {
	if path != "" {
		db, err := maxminddb.Open(path)
		if err == nil {
			return &Locator{db: db}
		}
		slog.Warn("Cannot open the GeoIP database, so only countries are known", "path", path, "error", err)
	}
	db, err := maxminddb.OpenBytes(countryLite)
	if err != nil {
		slog.Error("Cannot open the built in GeoIP database, so no locations are known", "error", err)
		return &Locator{}
	}
	return &Locator{db: db}
}

// Locate is where the IP address is estimated to be
func (l *Locator) Locate(ip string) Location {
	addr, err := netip.ParseAddr(ip)
	if l == nil || l.db == nil || err != nil {
		return Location{}
	}
	var r record
	if err := l.db.Lookup(addr).Decode(&r); err != nil {
		return Location{}
	}
	country := r.Country.ISOCode
	if country == "" {
		country = r.CountryCode
	}
	return Location{Country: country, City: r.City.Names["en"]}
}
