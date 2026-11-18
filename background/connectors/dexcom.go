package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/sitedb"
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
	refused := false
	refuse := func(err error, record any) {
		if !refused {
			refused = true
			recordRefused(ctx, jobInfo, err, record)
		}
	}
	client.Unreadable = func(raw json.RawMessage, err error) {
		refuse(fmt.Errorf("Dexcom: cannot read a reading: %w", err), raw)
	}
	egvs, err := client.FetchEGVs(ctx, minutes)
	if err != nil {
		return 0, fmt.Errorf("Dexcom Error: %w", err)
	}
	added, err := keepReadings(ctx, jobInfo.Site, egvs, func(e dexcommodels.EGV) sitedb.CreateGlucoseReadingParams {
		return dexcommodels.EGVToCreate(e, "dexcom_connector")
	}, func(err error, e dexcommodels.EGV) { refuse(fmt.Errorf("Dexcom: cannot keep a reading: %w", err), e) })
	if err != nil {
		return 0, fmt.Errorf("Dexcom Error: %w", err)
	}
	if added > 0 {
		slog.InfoContext(ctx, "Added new Dexcom entries for site", "added", added)
	}
	return added, nil
}
