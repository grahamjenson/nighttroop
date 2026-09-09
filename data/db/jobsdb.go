package db

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/utils/config"
)

type JobsDB struct {
	*jobs.DB
}

func NewJobsDB(ctx context.Context, cfg *config.ServerConfig) (*JobsDB, error) {

	rawDB, err := jobs.NewDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot open jobs db: %w", err)
	}

	// Create a new jobs db
	jobsDB := &JobsDB{
		DB: rawDB,
	}

	return jobsDB, nil
}
