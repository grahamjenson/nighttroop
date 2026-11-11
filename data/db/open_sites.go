package db

import "sync"

// openSites are the site DBs the server has open, by subdomain, for any number of goroutines to use at once. A site is
// served only while its DB is here.
type openSites struct {
	mu          sync.RWMutex
	bySubdomain map[string]*SiteDB
}

// get is the site's open DB, or nil when it is not open
func (o *openSites) get(subdomain string) *SiteDB {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.bySubdomain[subdomain]
}

func (o *openSites) add(site *SiteDB) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.bySubdomain[site.Subdomain] = site
}

// remove takes the site's DB out, for its caller to close, or is nil when it is not open
func (o *openSites) remove(subdomain string) *SiteDB {
	o.mu.Lock()
	defer o.mu.Unlock()
	site := o.bySubdomain[subdomain]
	delete(o.bySubdomain, subdomain)
	return site
}

func (o *openSites) count() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.bySubdomain)
}

func (o *openSites) all() []*SiteDB {
	o.mu.RLock()
	defer o.mu.RUnlock()
	sites := make([]*SiteDB, 0, len(o.bySubdomain))
	for _, site := range o.bySubdomain {
		sites = append(sites, site)
	}
	return sites
}
