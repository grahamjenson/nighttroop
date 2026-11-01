
package jobsdb

import (
	"nighttroop.t1d.nz/utils/errors"
)


// GetID returns the ID value.
func (m Job) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Job.
func (m *Job) SetID(v int64) {
	m.ID = v
}

// GetQueue returns the Queue value.
func (m Job) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated Job.
func (m *Job) SetQueue(v string) {
	m.Queue = v
}

// GetJob returns the Job value.
func (m Job) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated Job.
func (m *Job) SetJob(v string) {
	m.Job = v
}

// GetJobStatus returns the JobStatus value.
func (m Job) GetJobStatus() string {
	return m.JobStatus
}

// SetJobStatus sets the JobStatus value and returns the updated Job.
func (m *Job) SetJobStatus(v string) {
	m.JobStatus = v
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m Job) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated Job.
func (m *Job) SetExecuteAfter(v int64) {
	m.ExecuteAfter = v
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m Job) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// SetRemainingAttempts sets the RemainingAttempts value and returns the updated Job.
func (m *Job) SetRemainingAttempts(v int64) {
	m.RemainingAttempts = v
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m Job) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated Job.
func (m *Job) SetConsumerFetchedAt(v int64) {
	m.ConsumerFetchedAt = v
}

// GetFinishedAt returns the FinishedAt value.
func (m Job) GetFinishedAt() int64 {
	return m.FinishedAt
}

// SetFinishedAt sets the FinishedAt value and returns the updated Job.
func (m *Job) SetFinishedAt(v int64) {
	m.FinishedAt = v
}

// GetDedupingKey returns the DedupingKey value.
func (m Job) GetDedupingKey() string {
	return m.DedupingKey
}

// SetDedupingKey sets the DedupingKey value and returns the updated Job.
func (m *Job) SetDedupingKey(v string) {
	m.DedupingKey = v
}

// GetErrors returns the Errors value.
func (m Job) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated Job.
func (m *Job) SetErrors(v errors.ErrorList) {
	m.Errors = v
}

// GetRunID returns the RunID value.
func (m Job) GetRunID() *string {
	return m.RunID
}

// SetRunID sets the RunID value and returns the updated Job.
func (m *Job) SetRunID(v *string) {
	m.RunID = v
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

// GetDayQueueJob returns the DayQueueJob value.
func (m JobStatsDaily) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetDayQueueJob(v string) {
	m.DayQueueJob = v
}

// GetTotalJobs returns the TotalJobs value.
func (m JobStatsDaily) GetTotalJobs() int64 {
	return m.TotalJobs
}

// SetTotalJobs sets the TotalJobs value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetTotalJobs(v int64) {
	m.TotalJobs = v
}

// GetTotalSuccess returns the TotalSuccess value.
func (m JobStatsDaily) GetTotalSuccess() int64 {
	return m.TotalSuccess
}

// SetTotalSuccess sets the TotalSuccess value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetTotalSuccess(v int64) {
	m.TotalSuccess = v
}

// GetTotalFailed returns the TotalFailed value.
func (m JobStatsDaily) GetTotalFailed() int64 {
	return m.TotalFailed
}

// SetTotalFailed sets the TotalFailed value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetTotalFailed(v int64) {
	m.TotalFailed = v
}

// GetDurationSum returns the DurationSum value.
func (m JobStatsDaily) GetDurationSum() float64 {
	return m.DurationSum
}

// SetDurationSum sets the DurationSum value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetDurationSum(v float64) {
	m.DurationSum = v
}

// GetDurationSquaredSum returns the DurationSquaredSum value.
func (m JobStatsDaily) GetDurationSquaredSum() float64 {
	return m.DurationSquaredSum
}

// SetDurationSquaredSum sets the DurationSquaredSum value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetDurationSquaredSum(v float64) {
	m.DurationSquaredSum = v
}

// GetErrors returns the Errors value.
func (m JobStatsDaily) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetErrors(v errors.ErrorList) {
	m.Errors = v
}

// GetFailedRunIdsJson returns the FailedRunIdsJson value.
func (m JobStatsDaily) GetFailedRunIdsJson() errors.ErrorList {
	return m.FailedRunIdsJson
}

// SetFailedRunIdsJson sets the FailedRunIdsJson value and returns the updated JobStatsDaily.
func (m *JobStatsDaily) SetFailedRunIdsJson(v errors.ErrorList) {
	m.FailedRunIdsJson = v
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

// GetDayQueueJob returns the DayQueueJob value.
func (m EnsureJobStatsDailyForGroupParams) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated EnsureJobStatsDailyForGroupParams.
func (m *EnsureJobStatsDailyForGroupParams) SetDayQueueJob(v string) {
	m.DayQueueJob = v
}

// GetDayQueueJob returns the DayQueueJob value.
func (m GetJobStatsDailyParams) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated GetJobStatsDailyParams.
func (m *GetJobStatsDailyParams) SetDayQueueJob(v string) {
	m.DayQueueJob = v
}

// GetSkip returns the Skip value.
func (m SearchJobStatsDailyParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchJobStatsDailyParams.
func (m *SearchJobStatsDailyParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchJobStatsDailyParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchJobStatsDailyParams.
func (m *SearchJobStatsDailyParams) SetLimit(v int64) {
	m.Limit = v
}

// GetTotalJobs returns the TotalJobs value.
func (m UpdateJobStatsDailyParams) GetTotalJobs() int64 {
	return m.TotalJobs
}

// SetTotalJobs sets the TotalJobs value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetTotalJobs(v int64) {
	m.TotalJobs = v
}

// GetTotalSuccess returns the TotalSuccess value.
func (m UpdateJobStatsDailyParams) GetTotalSuccess() int64 {
	return m.TotalSuccess
}

// SetTotalSuccess sets the TotalSuccess value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetTotalSuccess(v int64) {
	m.TotalSuccess = v
}

// GetTotalFailed returns the TotalFailed value.
func (m UpdateJobStatsDailyParams) GetTotalFailed() int64 {
	return m.TotalFailed
}

// SetTotalFailed sets the TotalFailed value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetTotalFailed(v int64) {
	m.TotalFailed = v
}

// GetDurationSum returns the DurationSum value.
func (m UpdateJobStatsDailyParams) GetDurationSum() float64 {
	return m.DurationSum
}

// SetDurationSum sets the DurationSum value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetDurationSum(v float64) {
	m.DurationSum = v
}

// GetDurationSquaredSum returns the DurationSquaredSum value.
func (m UpdateJobStatsDailyParams) GetDurationSquaredSum() float64 {
	return m.DurationSquaredSum
}

// SetDurationSquaredSum sets the DurationSquaredSum value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetDurationSquaredSum(v float64) {
	m.DurationSquaredSum = v
}

// GetErrors returns the Errors value.
func (m UpdateJobStatsDailyParams) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetErrors(v errors.ErrorList) {
	m.Errors = v
}

// GetFailedRunIdsJson returns the FailedRunIdsJson value.
func (m UpdateJobStatsDailyParams) GetFailedRunIdsJson() errors.ErrorList {
	return m.FailedRunIdsJson
}

// SetFailedRunIdsJson sets the FailedRunIdsJson value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetFailedRunIdsJson(v errors.ErrorList) {
	m.FailedRunIdsJson = v
}

// GetDayQueueJob returns the DayQueueJob value.
func (m UpdateJobStatsDailyParams) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated UpdateJobStatsDailyParams.
func (m *UpdateJobStatsDailyParams) SetDayQueueJob(v string) {
	m.DayQueueJob = v
}

// GetID returns the ID value.
func (m CompleteJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated CompleteJobParams.
func (m *CompleteJobParams) SetID(v int64) {
	m.ID = v
}

// GetIds returns the Ids value.
func (m DeleteJobsByIdsParams) GetIds() []int64 {
	return m.Ids
}

// SetIds sets the Ids value and returns the updated DeleteJobsByIdsParams.
func (m *DeleteJobsByIdsParams) SetIds(v []int64) {
	m.Ids = v
}

// GetErrors returns the Errors value.
func (m FailJobParams) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated FailJobParams.
func (m *FailJobParams) SetErrors(v errors.ErrorList) {
	m.Errors = v
}

// GetRunID returns the RunID value.
func (m FailJobParams) GetRunID() *string {
	return m.RunID
}

// SetRunID sets the RunID value and returns the updated FailJobParams.
func (m *FailJobParams) SetRunID(v *string) {
	m.RunID = v
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

// GetSecondsAgo returns the SecondsAgo value.
func (m GetJobsToBeSummarisedParams) GetSecondsAgo() int64 {
	return m.SecondsAgo
}

// SetSecondsAgo sets the SecondsAgo value and returns the updated GetJobsToBeSummarisedParams.
func (m *GetJobsToBeSummarisedParams) SetSecondsAgo(v int64) {
	m.SecondsAgo = v
}

// GetJob returns the Job value.
func (m InsertFailedCronJobParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated InsertFailedCronJobParams.
func (m *InsertFailedCronJobParams) SetJob(v string) {
	m.Job = v
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m InsertFailedCronJobParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated InsertFailedCronJobParams.
func (m *InsertFailedCronJobParams) SetConsumerFetchedAt(v int64) {
	m.ConsumerFetchedAt = v
}

// GetErrors returns the Errors value.
func (m InsertFailedCronJobParams) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated InsertFailedCronJobParams.
func (m *InsertFailedCronJobParams) SetErrors(v errors.ErrorList) {
	m.Errors = v
}

// GetRunID returns the RunID value.
func (m InsertFailedCronJobParams) GetRunID() *string {
	return m.RunID
}

// SetRunID sets the RunID value and returns the updated InsertFailedCronJobParams.
func (m *InsertFailedCronJobParams) SetRunID(v *string) {
	m.RunID = v
}

// GetJob returns the Job value.
func (m InsertSuccessfulCronJobParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated InsertSuccessfulCronJobParams.
func (m *InsertSuccessfulCronJobParams) SetJob(v string) {
	m.Job = v
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m InsertSuccessfulCronJobParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated InsertSuccessfulCronJobParams.
func (m *InsertSuccessfulCronJobParams) SetConsumerFetchedAt(v int64) {
	m.ConsumerFetchedAt = v
}

// GetQueue returns the Queue value.
func (m MarkJobsForConsumerParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated MarkJobsForConsumerParams.
func (m *MarkJobsForConsumerParams) SetQueue(v string) {
	m.Queue = v
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m MarkJobsForConsumerParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated MarkJobsForConsumerParams.
func (m *MarkJobsForConsumerParams) SetExecuteAfter(v int64) {
	m.ExecuteAfter = v
}

// GetLimit returns the Limit value.
func (m MarkJobsForConsumerParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated MarkJobsForConsumerParams.
func (m *MarkJobsForConsumerParams) SetLimit(v int64) {
	m.Limit = v
}

// GetQueue returns the Queue value.
func (m QueueJobIgnoreDupeParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated QueueJobIgnoreDupeParams.
func (m *QueueJobIgnoreDupeParams) SetQueue(v string) {
	m.Queue = v
}

// GetJob returns the Job value.
func (m QueueJobIgnoreDupeParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated QueueJobIgnoreDupeParams.
func (m *QueueJobIgnoreDupeParams) SetJob(v string) {
	m.Job = v
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m QueueJobIgnoreDupeParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated QueueJobIgnoreDupeParams.
func (m *QueueJobIgnoreDupeParams) SetExecuteAfter(v int64) {
	m.ExecuteAfter = v
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m QueueJobIgnoreDupeParams) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// SetRemainingAttempts sets the RemainingAttempts value and returns the updated QueueJobIgnoreDupeParams.
func (m *QueueJobIgnoreDupeParams) SetRemainingAttempts(v int64) {
	m.RemainingAttempts = v
}

// GetDedupingKey returns the DedupingKey value.
func (m QueueJobIgnoreDupeParams) GetDedupingKey() string {
	return m.DedupingKey
}

// SetDedupingKey sets the DedupingKey value and returns the updated QueueJobIgnoreDupeParams.
func (m *QueueJobIgnoreDupeParams) SetDedupingKey(v string) {
	m.DedupingKey = v
}

// GetQueue returns the Queue value.
func (m QueueJobReplaceDupeParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated QueueJobReplaceDupeParams.
func (m *QueueJobReplaceDupeParams) SetQueue(v string) {
	m.Queue = v
}

// GetJob returns the Job value.
func (m QueueJobReplaceDupeParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated QueueJobReplaceDupeParams.
func (m *QueueJobReplaceDupeParams) SetJob(v string) {
	m.Job = v
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m QueueJobReplaceDupeParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated QueueJobReplaceDupeParams.
func (m *QueueJobReplaceDupeParams) SetExecuteAfter(v int64) {
	m.ExecuteAfter = v
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m QueueJobReplaceDupeParams) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// SetRemainingAttempts sets the RemainingAttempts value and returns the updated QueueJobReplaceDupeParams.
func (m *QueueJobReplaceDupeParams) SetRemainingAttempts(v int64) {
	m.RemainingAttempts = v
}

// GetDedupingKey returns the DedupingKey value.
func (m QueueJobReplaceDupeParams) GetDedupingKey() string {
	return m.DedupingKey
}

// SetDedupingKey sets the DedupingKey value and returns the updated QueueJobReplaceDupeParams.
func (m *QueueJobReplaceDupeParams) SetDedupingKey(v string) {
	m.DedupingKey = v
}

// GetQueue returns the Queue value.
func (m ResetJobsParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated ResetJobsParams.
func (m *ResetJobsParams) SetQueue(v string) {
	m.Queue = v
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m ResetJobsParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated ResetJobsParams.
func (m *ResetJobsParams) SetConsumerFetchedAt(v int64) {
	m.ConsumerFetchedAt = v
}

// GetSkip returns the Skip value.
func (m SearchJobsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchJobsParams.
func (m *SearchJobsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchJobsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchJobsParams.
func (m *SearchJobsParams) SetLimit(v int64) {
	m.Limit = v
}

