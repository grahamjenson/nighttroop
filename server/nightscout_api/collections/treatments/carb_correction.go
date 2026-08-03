package treatments

import (
	"context"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type CarbsCorrection struct{}

func (c CarbsCorrection) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment, createContext nightscout.CreateContext) (*site.SearchEventGroupsRow, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}

func (c CarbsCorrection) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
