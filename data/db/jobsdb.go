package db

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

type JobsDB struct {
	*jobs.DB
}

func NewJobsDB(ctx context.Context, path string, cfg config.ServerConfig) (*JobsDB, error) {

	rawDB, err := jobs.NewDB(ctx, utils.JOBS_DB_PATH, 100, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot open jobs db: %s %w", path, err)
	}

	// Create a new jobs db
	jobsDB := &JobsDB{
		DB: rawDB,
	}

	return jobsDB, nil
}
