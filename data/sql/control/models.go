package control

import "time"

func (u *User) LastSeenAtDate() *time.Time {
	if u.LastSeenAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastSeenAt)
	return &t
}

func (j *Job) DayQueueJob() string {
	// calculates the "YYYY-MM-DD::queue::job" string for the job based on the CreatedAt timestamp
	t := time.Unix(j.CreatedAt, 0)
	return t.Format("2006-01-02") + "::" + j.Queue + "::" + j.Job
}

func (j *Job) TotalFailed() int64 {
	if j.JobStatus == "failed" {
		return 1
	}
	return 0
}

func (j *Job) TotalSuccess() int64 {
	if j.JobStatus == "succeeded" {
		return 1
	}
	return 0
}

func (j *Job) Duration() float64 {
	// Annoying thing is this isnt how long it ran it is how long it waited
	// TODO: Make this return two values, How long it took from queuing to finish, and how long it took from fetched to complete
	return float64(j.UpdatedAt - j.CreatedAt)
}

func (j *Job) DurationSquared() float64 {
	// Annoying thing is this isnt how long it ran it is how long it waited
	// TODO: Make this return two values, How long it took from queuing to finish, and how long it took from fetched to complete
	return j.Duration() * j.Duration()
}
