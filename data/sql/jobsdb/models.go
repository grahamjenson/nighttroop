package jobsdb

import "time"

func (j *Job) DayQueueJob() string {
	// calculates the "YYYY-MM-DD::queue::job" string for the job based on the CreatedAt timestamp
	t := time.UnixMilli(j.CreatedAt)
	return t.Format("2006-01-02") + "::" + j.Queue + "::" + j.Job
}

func (j *Job) TotalFailed() int64 {
	if j.JobStatus == "failed" {
		return 1
	}
	return 0
}

func (j *Job) TotalSuccess() int64 {
	if j.JobStatus == "completed" {
		return 1
	}
	return 0
}

// Duration is how long the job ran, from a consumer fetching it to it finishing, in milliseconds
func (j *Job) Duration() float64 {
	return float64(j.FinishedAt - j.ConsumerFetchedAt)
}

func (j *Job) DurationSquared() float64 {
	return j.Duration() * j.Duration()
}
