package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/utils/consts"
)

func (b *SyncJobManager) syncDexcom(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("No user sync sources found for unique hash: %s", syncInfo.UniqueHash)
	}

	slog.Info("Syncing Dexcom", "unique_hash", syncInfo.UniqueHash)

	client := dexcom_client.NewClient(syncInfo.urlOrUsername, syncInfo.password, syncInfo.region == consts.DexcomRegionOther)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	egvs, err := client.FetchLastHourEGVs(ctx)
	if err != nil {
		return fmt.Errorf("Dexcom Error: %w", err)
	}

	for _, e := range egvs {
		fmt.Printf(
			"Dexcom %s %d mg/dL %s (%s)\n",
			e.DisplayTime.Format(time.RFC3339),
			e.Value,
			e.TrendArrow,
			e.Trend,
		)
	}
	return nil
}
