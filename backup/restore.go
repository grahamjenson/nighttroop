package backup

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/utils"
)

func TryRestoreAllFiles(s3Client *utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "TryRestoreAllFiles")
	fmt.Println("TryRestoreAllFiles")

	err := TryDownloadExtractControlDbs(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	err := TryDownloadExrtractUserDbs(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	return nil
}

func TryDownloadExtractControlDbs(s3Client *utils.S3Client) error {
	// If it already exists dont download
	if db.DBFileExists() {
		return nil
	}
	os.MkdirAll("data/backup", os.ModePerm)
	os.MkdirAll("data/db", os.ModePerm)
	found, err := db.S3Client.DownloadFile(db.BackupPath, db.Path)
	if !found && err == nil {
		fmt.Println("No control db backup found")
		return nil
	}
	// TODO: check to see the error is not a 404
	if err != nil {
		return fmt.Errorf("cannot download control db: %w", err)
	}

	return nil
}

func TryDownloadExrtractUserDbs(s3Client *utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "TryDownloadExrtractUserDbs")
	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("TryDownloadExrtractUserDbs running", len(db.userPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			userSubdomains, err := db.GetUserPoolUsersFromDB(pool.char)
			fmt.Println("UserSubdomains", string(pool.char), userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot get user pool users: %w", err)
			}
			err = pool.TryDownloadExrtractUserDbs(db.S3Client, userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot restore user pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the TryDownloadExrtractUserDbs  tasks failed: %v\n", err)
		return fmt.Errorf("cannot restore and open all users: %w", err)
	} else {
		fmt.Println("All TryDownloadExrtractUserDbs tasks succeeded!")
	}

	return nil
}

func (up *UserPool) RestoreAndOpenAllUsers(s3Client *S3Client, subdomains []string) error {

	if len(subdomains) == 0 {
		return nil
	}

	var err error
	// Check to see if user are already on disk
	for _, subdomain := range subdomains {
		_, err := up.OpenUser(subdomain)
		if err != nil {
			fmt.Printf("Error opening user db %v %v %v\n", string(up.char), subdomain, err)
			return fmt.Errorf("cannot open user db: %w", err)
		}
		// We ignore if the DB was not found to be created
	}

	restored := up.GetNUsersDBs()
	if restored > 0 {
		if restored != len(subdomains) {
			fmt.Printf("Not all user were restored for Tennant Pool %v %v %v\n", string(up.char), restored, len(subdomains))
			return fmt.Errorf("not all user were restored")
		}
		// If they should all be on disk if any are
		return nil
	}

	// Now we need to see if we want to restore from S3
	fmt.Println("Trying to restore user from backup for pool", string(up.char))

	found, err := s3Client.DownloadAndExtractLZ4(up.S3Path, up.Path)
	fmt.Println("Finished s3 download backup for pool", string(up.char))
	if !found && err == nil {
		fmt.Println("No user backup found")
		return nil
	}
	if err != nil {
		// TODO dont fail here if just not found
		return fmt.Errorf("cannot download user db: %w", err)
	}

	for _, subdomain := range subdomains {

		// At this point in time we are ignoring errors
		user, err := up.OpenUser(subdomain)
		if err != nil {
			fmt.Printf("Error opening user db %v %v %v\n", string(up.char), subdomain, err)
			return fmt.Errorf("cannot open user db: %w", err)
		}
		if user == nil {
			fmt.Printf("User db %v %v not found\n", string(up.char), subdomain)
			return fmt.Errorf("cannot open user db: %w", err)
		}
	}

	return nil

}
