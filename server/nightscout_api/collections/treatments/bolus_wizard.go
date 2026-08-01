package treatments

import (
	"context"

	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

type BolusWizard struct {
}

func NewBolusWizard() BolusWizard {
	return BolusWizard{}
}

func (u BolusWizard) Create(ctx context.Context, siteContext *httputils.SiteContext, treatment nightscout.Treatment) (*nsutils.Record, *nsutils.HTTPError) {

	return nil, nsutils.ErrNotImplemented
}

func (u BolusWizard) Update(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any) (*nsutils.Record, *nsutils.HTTPError) {
	return nil, nsutils.ErrNotImplemented
}
