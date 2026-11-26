package connectors

import (
	"context"
	"log/slog"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/models/dexcommodels"
	"nighttroop.t1d.nz/utils/consts"
)

// syncDexcom copies the last 15 minutes of Dexcom readings into the site, returning how many were new
func (b *ConnectorJobManager) syncDexcom(jobInfo *ConnectorJobInfo) (int, error) {
	slog.InfoContext(jobInfo.ctx, "Syncing Dexcom")

	ctx, cancel := context.WithTimeout(jobInfo.ctx, syncTimeout)
	defer cancel()

	return copyDexcomEGVs(ctx, jobInfo, 15)
}

// copyDexcomEGVs copies the last minutes of Dexcom readings into the site, returning how many were new
func copyDexcomEGVs(ctx context.Context, jobInfo *ConnectorJobInfo, minutes int) (int, error) {
	client := dexcom_client.NewClient(jobInfo.urlOrUsername, jobInfo.password, consts.NewDexcomRegionFromString(jobInfo.region) == consts.DexcomRegionOther)
	fetch := func(ctx context.Context) ([]dexcommodels.EGV, error) { return client.FetchEGVs(ctx, minutes) }
	return copyReadings(ctx, jobInfo, "Dexcom", &client.Unreadable, fetch, dexcommodels.EGVToCreate)
}
