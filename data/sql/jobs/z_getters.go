
package jobs

import (
	"nighttroop.t1d.nz/utils/errors"
)


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
func (m EnsureJobStatsDailyForGroupParams) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated EnsureJobStatsDailyForGroupParams.
func (m EnsureJobStatsDailyForGroupParams) SetDayQueueJob(v string) EnsureJobStatsDailyForGroupParams {
	m.DayQueueJob = v
	return m
}

// GetTotalJobs returns the TotalJobs value.
func (m UpdateJobStatsDailyParams) GetTotalJobs() int64 {
	return m.TotalJobs
}

// SetTotalJobs sets the TotalJobs value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetTotalJobs(v int64) UpdateJobStatsDailyParams {
	m.TotalJobs = v
	return m
}

// GetTotalSuccess returns the TotalSuccess value.
func (m UpdateJobStatsDailyParams) GetTotalSuccess() int64 {
	return m.TotalSuccess
}

// SetTotalSuccess sets the TotalSuccess value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetTotalSuccess(v int64) UpdateJobStatsDailyParams {
	m.TotalSuccess = v
	return m
}

// GetTotalFailed returns the TotalFailed value.
func (m UpdateJobStatsDailyParams) GetTotalFailed() int64 {
	return m.TotalFailed
}

// SetTotalFailed sets the TotalFailed value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetTotalFailed(v int64) UpdateJobStatsDailyParams {
	m.TotalFailed = v
	return m
}

// GetDurationSum returns the DurationSum value.
func (m UpdateJobStatsDailyParams) GetDurationSum() float64 {
	return m.DurationSum
}

// SetDurationSum sets the DurationSum value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetDurationSum(v float64) UpdateJobStatsDailyParams {
	m.DurationSum = v
	return m
}

// GetDurationSquaredSum returns the DurationSquaredSum value.
func (m UpdateJobStatsDailyParams) GetDurationSquaredSum() float64 {
	return m.DurationSquaredSum
}

// SetDurationSquaredSum sets the DurationSquaredSum value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetDurationSquaredSum(v float64) UpdateJobStatsDailyParams {
	m.DurationSquaredSum = v
	return m
}

// GetErrors returns the Errors value.
func (m UpdateJobStatsDailyParams) GetErrors() interface{} {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetErrors(v interface{}) UpdateJobStatsDailyParams {
	m.Errors = v
	return m
}

// GetDayQueueJob returns the DayQueueJob value.
func (m UpdateJobStatsDailyParams) GetDayQueueJob() string {
	return m.DayQueueJob
}

// SetDayQueueJob sets the DayQueueJob value and returns the updated UpdateJobStatsDailyParams.
func (m UpdateJobStatsDailyParams) SetDayQueueJob(v string) UpdateJobStatsDailyParams {
	m.DayQueueJob = v
	return m
}

// GetQueue returns the Queue value.
func (m QueueJobIgnoreDupeParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated QueueJobIgnoreDupeParams.
func (m QueueJobIgnoreDupeParams) SetQueue(v string) QueueJobIgnoreDupeParams {
	m.Queue = v
	return m
}

// GetJob returns the Job value.
func (m QueueJobIgnoreDupeParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated QueueJobIgnoreDupeParams.
func (m QueueJobIgnoreDupeParams) SetJob(v string) QueueJobIgnoreDupeParams {
	m.Job = v
	return m
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m QueueJobIgnoreDupeParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated QueueJobIgnoreDupeParams.
func (m QueueJobIgnoreDupeParams) SetExecuteAfter(v int64) QueueJobIgnoreDupeParams {
	m.ExecuteAfter = v
	return m
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m QueueJobIgnoreDupeParams) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// SetRemainingAttempts sets the RemainingAttempts value and returns the updated QueueJobIgnoreDupeParams.
func (m QueueJobIgnoreDupeParams) SetRemainingAttempts(v int64) QueueJobIgnoreDupeParams {
	m.RemainingAttempts = v
	return m
}

// GetDedupingKey returns the DedupingKey value.
func (m QueueJobIgnoreDupeParams) GetDedupingKey() string {
	return m.DedupingKey
}

// SetDedupingKey sets the DedupingKey value and returns the updated QueueJobIgnoreDupeParams.
func (m QueueJobIgnoreDupeParams) SetDedupingKey(v string) QueueJobIgnoreDupeParams {
	m.DedupingKey = v
	return m
}

// GetQueue returns the Queue value.
func (m QueueJobReplaceDupeParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated QueueJobReplaceDupeParams.
func (m QueueJobReplaceDupeParams) SetQueue(v string) QueueJobReplaceDupeParams {
	m.Queue = v
	return m
}

// GetJob returns the Job value.
func (m QueueJobReplaceDupeParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated QueueJobReplaceDupeParams.
func (m QueueJobReplaceDupeParams) SetJob(v string) QueueJobReplaceDupeParams {
	m.Job = v
	return m
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m QueueJobReplaceDupeParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated QueueJobReplaceDupeParams.
func (m QueueJobReplaceDupeParams) SetExecuteAfter(v int64) QueueJobReplaceDupeParams {
	m.ExecuteAfter = v
	return m
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m QueueJobReplaceDupeParams) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// SetRemainingAttempts sets the RemainingAttempts value and returns the updated QueueJobReplaceDupeParams.
func (m QueueJobReplaceDupeParams) SetRemainingAttempts(v int64) QueueJobReplaceDupeParams {
	m.RemainingAttempts = v
	return m
}

// GetDedupingKey returns the DedupingKey value.
func (m QueueJobReplaceDupeParams) GetDedupingKey() string {
	return m.DedupingKey
}

// SetDedupingKey sets the DedupingKey value and returns the updated QueueJobReplaceDupeParams.
func (m QueueJobReplaceDupeParams) SetDedupingKey(v string) QueueJobReplaceDupeParams {
	m.DedupingKey = v
	return m
}

// GetID returns the ID value.
func (m CompleteJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated CompleteJobParams.
func (m CompleteJobParams) SetID(v int64) CompleteJobParams {
	m.ID = v
	return m
}

// GetErrors returns the Errors value.
func (m FailJobParams) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated FailJobParams.
func (m FailJobParams) SetErrors(v errors.ErrorList) FailJobParams {
	m.Errors = v
	return m
}

// GetID returns the ID value.
func (m FailJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated FailJobParams.
func (m FailJobParams) SetID(v int64) FailJobParams {
	m.ID = v
	return m
}

// GetQueue returns the Queue value.
func (m MarkJobsForConsumerParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated MarkJobsForConsumerParams.
func (m MarkJobsForConsumerParams) SetQueue(v string) MarkJobsForConsumerParams {
	m.Queue = v
	return m
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m MarkJobsForConsumerParams) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// SetExecuteAfter sets the ExecuteAfter value and returns the updated MarkJobsForConsumerParams.
func (m MarkJobsForConsumerParams) SetExecuteAfter(v int64) MarkJobsForConsumerParams {
	m.ExecuteAfter = v
	return m
}

// GetLimit returns the Limit value.
func (m MarkJobsForConsumerParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated MarkJobsForConsumerParams.
func (m MarkJobsForConsumerParams) SetLimit(v int64) MarkJobsForConsumerParams {
	m.Limit = v
	return m
}

// GetQueue returns the Queue value.
func (m ResetJobsParams) GetQueue() string {
	return m.Queue
}

// SetQueue sets the Queue value and returns the updated ResetJobsParams.
func (m ResetJobsParams) SetQueue(v string) ResetJobsParams {
	m.Queue = v
	return m
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m ResetJobsParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated ResetJobsParams.
func (m ResetJobsParams) SetConsumerFetchedAt(v int64) ResetJobsParams {
	m.ConsumerFetchedAt = v
	return m
}

// GetID returns the ID value.
func (m FindJobParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated FindJobParams.
func (m FindJobParams) SetID(v int64) FindJobParams {
	m.ID = v
	return m
}

// GetSecondsAgo returns the SecondsAgo value.
func (m GetJobsToBeSummarisedParams) GetSecondsAgo() int64 {
	return m.SecondsAgo
}

// SetSecondsAgo sets the SecondsAgo value and returns the updated GetJobsToBeSummarisedParams.
func (m GetJobsToBeSummarisedParams) SetSecondsAgo(v int64) GetJobsToBeSummarisedParams {
	m.SecondsAgo = v
	return m
}

// GetIds returns the Ids value.
func (m DeleteJobsByIdsParams) GetIds() []int64 {
	return m.Ids
}

// SetIds sets the Ids value and returns the updated DeleteJobsByIdsParams.
func (m DeleteJobsByIdsParams) SetIds(v []int64) DeleteJobsByIdsParams {
	m.Ids = v
	return m
}

// GetJob returns the Job value.
func (m InsertSuccessfulCronJobParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated InsertSuccessfulCronJobParams.
func (m InsertSuccessfulCronJobParams) SetJob(v string) InsertSuccessfulCronJobParams {
	m.Job = v
	return m
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m InsertSuccessfulCronJobParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated InsertSuccessfulCronJobParams.
func (m InsertSuccessfulCronJobParams) SetConsumerFetchedAt(v int64) InsertSuccessfulCronJobParams {
	m.ConsumerFetchedAt = v
	return m
}

// GetJob returns the Job value.
func (m InsertFailedCronJobParams) GetJob() string {
	return m.Job
}

// SetJob sets the Job value and returns the updated InsertFailedCronJobParams.
func (m InsertFailedCronJobParams) SetJob(v string) InsertFailedCronJobParams {
	m.Job = v
	return m
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m InsertFailedCronJobParams) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// SetConsumerFetchedAt sets the ConsumerFetchedAt value and returns the updated InsertFailedCronJobParams.
func (m InsertFailedCronJobParams) SetConsumerFetchedAt(v int64) InsertFailedCronJobParams {
	m.ConsumerFetchedAt = v
	return m
}

// GetErrors returns the Errors value.
func (m InsertFailedCronJobParams) GetErrors() errors.ErrorList {
	return m.Errors
}

// SetErrors sets the Errors value and returns the updated InsertFailedCronJobParams.
func (m InsertFailedCronJobParams) SetErrors(v errors.ErrorList) InsertFailedCronJobParams {
	m.Errors = v
	return m
}

