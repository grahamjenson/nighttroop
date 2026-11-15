package site_data_templ

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/addmodal"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestFieldLimits checks every data form's fields name a column that is there, and only narrow its limits
func TestFieldLimits(t *testing.T) {
	ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{}, "site", "", 0, 0, "")
	ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{})
	forms := map[string][]addmodal.Field{
		"glucose": glucoseFields(ctx), "insulin": insulinDoseFields(ctx), "food": foodIntakeFields(ctx),
		"temp basal": tempBasalFields(ctx), "temp target": tempTargetFields(ctx), "profile switch": profileSwitchFields(ctx),
		"activity": activityFields(ctx), "notification": notificationEventFields(ctx),
	}
	for _, groupType := range sitemodels.EventGroupTypes {
		forms["new "+groupType] = eventGroupForm(ctx, nil, groupType).Fields
		forms["edit "+groupType] = eventGroupForm(ctx, &sitemodels.SiteEventGroup{}, groupType).Fields
	}
	for name, fields := range forms {
		require.NoError(t, addmodal.Check(fields), name)
	}
}
