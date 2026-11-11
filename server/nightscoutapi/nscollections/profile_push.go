package nscollections

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/push"
)

// Stored registers the app that uploaded the profile for pushes, when it says where they go: Loop in its loopSettings,
// for its remote commands, and Trio at the top of its profile, where LoopFollow sends its remote commands. Each is an
// APNs token for the identity that uploaded it, from a production build unless isAPNSProduction is false. A
// registration that cannot be kept is logged, as the profile is still kept.
func (c ProfileCollection) Stored(ctx context.Context, siteContext *httputils.SiteContext, nsModel nightscoutmodels.ProfileStore) {
	registration := push.Registration{Provider: consts.PushProviderAPNs, Production: true}
	remote := nsModel.RemoteSettings
	switch loop := nsModel.LoopSettings; {
	case loop != nil && loop.DeviceToken != nil && loop.BundleIdentifier != nil:
		registration.App, registration.BundleID, registration.Token = string(consts.PushAppLoop), string(*loop.BundleIdentifier), string(*loop.DeviceToken)
	case remote != nil && remote.DeviceToken != nil && remote.BundleIdentifier != nil:
		registration.App, registration.BundleID, registration.Token = string(consts.PushAppTrio), string(*remote.BundleIdentifier), string(*remote.DeviceToken)
	default:
		return
	}
	if remote != nil && remote.IsAPNSProduction != nil {
		registration.Production = bool(*remote.IsAPNSProduction)
	}
	if _, err := push.RegisterAlone(ctx, siteContext.ControlDB, siteContext.Config.Security, siteContext.SiteID,
		siteContext.IdentityModel.ID, registration); err != nil {
		slog.WarnContext(ctx, "Cannot register a profile's app for pushes", "app", registration.App, "error", err)
	}
}
