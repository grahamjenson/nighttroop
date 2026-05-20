package user

import (
	"fmt"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

func (syncSource *SyncSource) GetRawData(security *security.SecurityConfig) (unencryptedURLOrUsername string, unencryptedPassword string, uniqHash string, err error) {
	unencryptedPassword, err = security.Decrypt(syncSource.Password)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting sync source password: %w", err)
	}
	unencryptedURLOrUsername, err = security.Decrypt(syncSource.UrlOrUsername)
	if err != nil {
		return "", "", "", fmt.Errorf("error decrypting sync source URL or username: %w", err)
	}

	// Any two users with this constalations will be making the exact same request to the same URL.
	// We are deduping these requests to prevent multiple calls
	uniqString := fmt.Sprintf("%s:%s:%s", syncSource.Type, unencryptedURLOrUsername, unencryptedPassword)
	if syncSource.Type == consts.SyncSourceTypeDexcom {
		uniqString = fmt.Sprintf("%s:%s:%s:%s", syncSource.Type, unencryptedURLOrUsername, unencryptedPassword, syncSource.DexcomRegion)
	}
	uniqHash, err = security.GenerateHash(uniqString)
	if err != nil {
		return "", "", "", fmt.Errorf("error generating hash for sync job: %w", err)
	}

	return unencryptedURLOrUsername, unencryptedPassword, uniqHash, nil
}
