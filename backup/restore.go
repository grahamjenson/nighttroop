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

	err := TryRestoreControlDBFile(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	err := RestoreAndOpenAllUsers(s3Client)
	if err != nil {
		return fmt.Errorf("cannot try download all files: %w", err)
	}

	return nil
}

func TryRestoreControlDBFile(s3Client *utils.S3Client) error {
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

func RestoreAndOpenAllUsers(s3Client *utils.S3Client) error {
	defer utils.TrackTime(time.Now(), "RestoreAndOpenAllUsers")
	g, _ := errgroup.WithContext(context.Background())
	limit := int(math.Max(1, float64(runtime.NumCPU()-1)))
	g.SetLimit(limit)
	fmt.Println("RestoreAndOpenAllUsers running", len(db.userPools), limit)
	for _, pool := range db.allPools() {

		g.Go(func() error {
			userSubdomains, err := db.GetUserPoolUsersFromDB(pool.char)
			fmt.Println("UserSubdomains", string(pool.char), userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot get user pool users: %w", err)
			}
			err = pool.RestoreAndOpenAllUsers(db.S3Client, userSubdomains)
			if err != nil {
				return fmt.Errorf("cannot restore user pool: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("One of the RestoreAndOpenAllUsers  tasks failed: %v\n", err)
		return fmt.Errorf("cannot restore and open all users: %w", err)
	} else {
		fmt.Println("All RestoreAndOpenAllUsers tasks succeeded!")
	}

	return nil
}
