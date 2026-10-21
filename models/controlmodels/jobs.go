package controlmodels

// Job is a background job on a queue, until it is summarised into the day's job stats
type Job struct {
	ID                int64    `json:"id"`
	Queue             string   `json:"queue"`
	Job               string   `json:"job"`
	JobStatus         string   `json:"job_status"`
	ExecuteAfter      int64    `json:"execute_after"`
	RemainingAttempts int64    `json:"remaining_attempts"`
	ConsumerFetchedAt int64    `json:"consumer_fetched_at"`
	FinishedAt        int64    `json:"finished_at"`
	DedupingKey       string   `json:"deduping_key"`
	Errors            []string `json:"errors"`
	CreatedAt         int64    `json:"created_at"`
	UpdatedAt         int64    `json:"updated_at"`
}
