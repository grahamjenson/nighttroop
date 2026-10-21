package controlmodels

// JobStatsDaily is one day's totals for a queue's job, built from the finished jobs it summarises
type JobStatsDaily struct {
	// DayQueueJob is YYYY-MM-DD::queue::job
	DayQueueJob        string   `json:"day_queue_job"`
	TotalJobs          int64    `json:"total_jobs"`
	TotalSuccess       int64    `json:"total_success"`
	TotalFailed        int64    `json:"total_failed"`
	DurationSum        float64  `json:"duration_sum"`
	DurationSquaredSum float64  `json:"duration_squared_sum"`
	Errors             []string `json:"errors"`
	CreatedAt          int64    `json:"created_at"`
	UpdatedAt          int64    `json:"updated_at"`
}
