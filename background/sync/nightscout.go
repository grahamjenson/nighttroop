package sync

import (
	"context"
	"fmt"
	"time"

	"nighttroop.t1d.nz/clients/nightscout_client"
)

func (b *SyncJobManager) syncNightscout(syncInfo *SyncJobInfo) error {
	if len(syncInfo.userSyncSources) == 0 {
		return fmt.Errorf("No user sync sources found for unique hash: %s", syncInfo.UniqueHash)
	}

	nsc := nightscout_client.NewClient(syncInfo.urlOrUsername, syncInfo.password)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	egvs, err := nsc.GetEntries(ctx, 10)
	if err != nil {
		return fmt.Errorf("Nightscout Error: %w", err)
	}

	for _, e := range egvs {
		fmt.Printf(
			"NS %s %f mg/dL %s\n",
			e.DateString.Format(time.RFC3339),
			e.SGV,
			e.Direction,
		)
	}
	return nil

}
