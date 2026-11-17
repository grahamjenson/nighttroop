
package jobsdb


// GetID returns the ID value.
func (m Job) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Job.
func (m *Job) SetID(v int64) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Job) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Job.
func (m *Job) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m Job) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated Job.
func (m *Job) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetCreatedAt returns the CreatedAt value.
func (m JobStatsDaily) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m JobStatsDaily) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// SetLimit sets the Limit value and returns the updated SearchJobStatsDailyParams.
func (m *SearchJobStatsDailyParams) SetLimit(v int64) {
	m.Limit = v
}

// GetID returns the ID value.
func (m CompleteJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated CompleteJobParams.
func (m *CompleteJobParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m FailJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated FailJobParams.
func (m *FailJobParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m FindJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated FindJobParams.
func (m *FindJobParams) SetID(v int64) {
	m.ID = v
}

// SetLimit sets the Limit value and returns the updated MarkJobsForConsumerParams.
func (m *MarkJobsForConsumerParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated SearchJobsParams.
func (m *SearchJobsParams) SetLimit(v int64) {
	m.Limit = v
}

