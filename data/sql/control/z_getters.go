
package control

import (
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
)


// GetID returns the ID value.
func (m DataSourceBlockList) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetID(v int64) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m DataSourceBlockList) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUsername returns the Username value.
func (m DataSourceBlockList) GetUsername() *string {
	return m.Username
}

// SetUsername sets the Username value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetUsername(v *string) {
	m.Username = v
}

// GetUrl returns the Url value.
func (m DataSourceBlockList) GetUrl() *string {
	return m.Url
}

// SetUrl sets the Url value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetUrl(v *string) {
	m.Url = v
}

// GetReason returns the Reason value.
func (m DataSourceBlockList) GetReason() *string {
	return m.Reason
}

// SetReason sets the Reason value and returns the updated DataSourceBlockList.
func (m *DataSourceBlockList) SetReason(v *string) {
	m.Reason = v
}

// GetID returns the ID value.
func (m Event) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Event.
func (m *Event) SetID(v int64) {
	m.ID = v
}

// GetType returns the Type value.
func (m Event) GetType() *string {
	return m.Type
}

// SetType sets the Type value and returns the updated Event.
func (m *Event) SetType(v *string) {
	m.Type = v
}

// GetDetails returns the Details value.
func (m Event) GetDetails() *string {
	return m.Details
}

// SetDetails sets the Details value and returns the updated Event.
func (m *Event) SetDetails(v *string) {
	m.Details = v
}

// GetStart returns the Start value.
func (m Event) GetStart() int64 {
	return m.Start
}

// SetStart sets the Start value and returns the updated Event.
func (m *Event) SetStart(v int64) {
	m.Start = v
}

// GetEnd returns the End value.
func (m Event) GetEnd() *int64 {
	return m.End
}

// SetEnd sets the End value and returns the updated Event.
func (m *Event) SetEnd(v *int64) {
	m.End = v
}

// GetID returns the ID value.
func (m Identity) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Identity.
func (m *Identity) SetID(v int64) {
	m.ID = v
}

// GetSubdomain returns the Subdomain value.
func (m Identity) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated Identity.
func (m *Identity) SetSubdomain(v string) {
	m.Subdomain = v
}

// GetName returns the Name value.
func (m Identity) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated Identity.
func (m *Identity) SetName(v string) {
	m.Name = v
}

// GetPassword returns the Password value.
func (m Identity) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated Identity.
func (m *Identity) SetPassword(v string) {
	m.Password = v
}

// GetTotpSecret returns the TotpSecret value.
func (m Identity) GetTotpSecret() string {
	return m.TotpSecret
}

// SetTotpSecret sets the TotpSecret value and returns the updated Identity.
func (m *Identity) SetTotpSecret(v string) {
	m.TotpSecret = v
}

// GetRole returns the Role value.
func (m Identity) GetRole() consts.IdentityRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated Identity.
func (m *Identity) SetRole(v consts.IdentityRole) {
	m.Role = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Identity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Identity.
func (m *Identity) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m Identity) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated Identity.
func (m *Identity) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m Identity) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated Identity.
func (m *Identity) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m Identity) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// SetUpdatedByIdentityID sets the UpdatedByIdentityID value and returns the updated Identity.
func (m *Identity) SetUpdatedByIdentityID(v *int64) {
	m.UpdatedByIdentityID = v
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m Identity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// SetUpdatedByIdentityName sets the UpdatedByIdentityName value and returns the updated Identity.
func (m *Identity) SetUpdatedByIdentityName(v *string) {
	m.UpdatedByIdentityName = v
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m Identity) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated Identity.
func (m *Identity) SetCreatedByIdentityID(v int64) {
	m.CreatedByIdentityID = v
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m Identity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated Identity.
func (m *Identity) SetCreatedByIdentityName(v string) {
	m.CreatedByIdentityName = v
}

// GetDeletedAt returns the DeletedAt value.
func (m Identity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated Identity.
func (m *Identity) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetEnabled returns the Enabled value.
func (m Identity) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated Identity.
func (m *Identity) SetEnabled(v bool) {
	m.Enabled = v
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

// GetID returns the ID value.
func (m SiteSyncSource) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetID(v int64) {
	m.ID = v
}

// GetSubdomain returns the Subdomain value.
func (m SiteSyncSource) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetSubdomain(v string) {
	m.Subdomain = v
}

// GetType returns the Type value.
func (m SiteSyncSource) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetType(v consts.SyncSourceType) {
	m.Type = v
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m SiteSyncSource) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetUrlOrUsername(v string) {
	m.UrlOrUsername = v
}

// GetPassword returns the Password value.
func (m SiteSyncSource) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetPassword(v string) {
	m.Password = v
}

// GetDexcomRegion returns the DexcomRegion value.
func (m SiteSyncSource) GetDexcomRegion() consts.DexcomRegion {
	return m.DexcomRegion
}

// SetDexcomRegion sets the DexcomRegion value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetDexcomRegion(v consts.DexcomRegion) {
	m.DexcomRegion = v
}

// GetCreatedAt returns the CreatedAt value.
func (m SiteSyncSource) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m SiteSyncSource) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetEnabled returns the Enabled value.
func (m SiteSyncSource) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetEnabled(v bool) {
	m.Enabled = v
}

// GetUniqueHash returns the UniqueHash value.
func (m SiteSyncSource) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated SiteSyncSource.
func (m *SiteSyncSource) SetUniqueHash(v string) {
	m.UniqueHash = v
}

// GetID returns the ID value.
func (m Site) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Site.
func (m *Site) SetID(v int64) {
	m.ID = v
}

// GetSubdomain returns the Subdomain value.
func (m Site) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated Site.
func (m *Site) SetSubdomain(v string) {
	m.Subdomain = v
}

// GetBatch returns the Batch value.
func (m Site) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated Site.
func (m *Site) SetBatch(v string) {
	m.Batch = v
}

// GetBackupKeyHash returns the BackupKeyHash value.
func (m Site) GetBackupKeyHash() string {
	return m.BackupKeyHash
}

// SetBackupKeyHash sets the BackupKeyHash value and returns the updated Site.
func (m *Site) SetBackupKeyHash(v string) {
	m.BackupKeyHash = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Site) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Site.
func (m *Site) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetMaxDbSizeKbLimit returns the MaxDbSizeKbLimit value.
func (m Site) GetMaxDbSizeKbLimit() *int64 {
	return m.MaxDbSizeKbLimit
}

// SetMaxDbSizeKbLimit sets the MaxDbSizeKbLimit value and returns the updated Site.
func (m *Site) SetMaxDbSizeKbLimit(v *int64) {
	m.MaxDbSizeKbLimit = v
}

// GetLastSeenAt returns the LastSeenAt value.
func (m Site) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated Site.
func (m *Site) SetLastSeenAt(v *int64) {
	m.LastSeenAt = v
}

// GetLastWriteAt returns the LastWriteAt value.
func (m Site) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated Site.
func (m *Site) SetLastWriteAt(v *int64) {
	m.LastWriteAt = v
}

// GetDeletedAt returns the DeletedAt value.
func (m Site) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// SetDeletedAt sets the DeletedAt value and returns the updated Site.
func (m *Site) SetDeletedAt(v *int64) {
	m.DeletedAt = v
}

// GetName returns the Name value.
func (m SqliteSequence) GetName() interface{} {
	return m.Name
}

// SetName sets the Name value and returns the updated SqliteSequence.
func (m *SqliteSequence) SetName(v interface{}) {
	m.Name = v
}

// GetSeq returns the Seq value.
func (m SqliteSequence) GetSeq() interface{} {
	return m.Seq
}

// SetSeq sets the Seq value and returns the updated SqliteSequence.
func (m *SqliteSequence) SetSeq(v interface{}) {
	m.Seq = v
}

// GetID returns the ID value.
func (m SyncSourceReport) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetID(v int64) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m SyncSourceReport) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt value.
func (m SyncSourceReport) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// SetUpdatedAt sets the UpdatedAt value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetUpdatedAt(v int64) {
	m.UpdatedAt = v
}

// GetUniqueHash returns the UniqueHash value.
func (m SyncSourceReport) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetUniqueHash(v string) {
	m.UniqueHash = v
}

// GetSyncIntervalSecondsReason returns the SyncIntervalSecondsReason value.
func (m SyncSourceReport) GetSyncIntervalSecondsReason() string {
	return m.SyncIntervalSecondsReason
}

// SetSyncIntervalSecondsReason sets the SyncIntervalSecondsReason value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetSyncIntervalSecondsReason(v string) {
	m.SyncIntervalSecondsReason = v
}

// GetSyncIntervalSeconds returns the SyncIntervalSeconds value.
func (m SyncSourceReport) GetSyncIntervalSeconds() int64 {
	return m.SyncIntervalSeconds
}

// SetSyncIntervalSeconds sets the SyncIntervalSeconds value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetSyncIntervalSeconds(v int64) {
	m.SyncIntervalSeconds = v
}

// GetLastAttemptedSync returns the LastAttemptedSync value.
func (m SyncSourceReport) GetLastAttemptedSync() *int64 {
	return m.LastAttemptedSync
}

// SetLastAttemptedSync sets the LastAttemptedSync value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetLastAttemptedSync(v *int64) {
	m.LastAttemptedSync = v
}

// GetLastSuccessfulSync returns the LastSuccessfulSync value.
func (m SyncSourceReport) GetLastSuccessfulSync() *int64 {
	return m.LastSuccessfulSync
}

// SetLastSuccessfulSync sets the LastSuccessfulSync value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetLastSuccessfulSync(v *int64) {
	m.LastSuccessfulSync = v
}

// GetLastFailedSync returns the LastFailedSync value.
func (m SyncSourceReport) GetLastFailedSync() *int64 {
	return m.LastFailedSync
}

// SetLastFailedSync sets the LastFailedSync value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetLastFailedSync(v *int64) {
	m.LastFailedSync = v
}

// GetTotalSyncs returns the TotalSyncs value.
func (m SyncSourceReport) GetTotalSyncs() *int64 {
	return m.TotalSyncs
}

// SetTotalSyncs sets the TotalSyncs value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetTotalSyncs(v *int64) {
	m.TotalSyncs = v
}

// GetTotalSuccessfulSyncs returns the TotalSuccessfulSyncs value.
func (m SyncSourceReport) GetTotalSuccessfulSyncs() *int64 {
	return m.TotalSuccessfulSyncs
}

// SetTotalSuccessfulSyncs sets the TotalSuccessfulSyncs value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetTotalSuccessfulSyncs(v *int64) {
	m.TotalSuccessfulSyncs = v
}

// GetTotalFailedSyncs returns the TotalFailedSyncs value.
func (m SyncSourceReport) GetTotalFailedSyncs() *int64 {
	return m.TotalFailedSyncs
}

// SetTotalFailedSyncs sets the TotalFailedSyncs value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetTotalFailedSyncs(v *int64) {
	m.TotalFailedSyncs = v
}

// GetConsecutiveSuccessfulSyncs returns the ConsecutiveSuccessfulSyncs value.
func (m SyncSourceReport) GetConsecutiveSuccessfulSyncs() *int64 {
	return m.ConsecutiveSuccessfulSyncs
}

// SetConsecutiveSuccessfulSyncs sets the ConsecutiveSuccessfulSyncs value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetConsecutiveSuccessfulSyncs(v *int64) {
	m.ConsecutiveSuccessfulSyncs = v
}

// GetConsecutiveFailedSyncs returns the ConsecutiveFailedSyncs value.
func (m SyncSourceReport) GetConsecutiveFailedSyncs() *int64 {
	return m.ConsecutiveFailedSyncs
}

// SetConsecutiveFailedSyncs sets the ConsecutiveFailedSyncs value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetConsecutiveFailedSyncs(v *int64) {
	m.ConsecutiveFailedSyncs = v
}

// GetPreviousSyncErrors returns the PreviousSyncErrors value.
func (m SyncSourceReport) GetPreviousSyncErrors() errors.ErrorList {
	return m.PreviousSyncErrors
}

// SetPreviousSyncErrors sets the PreviousSyncErrors value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetPreviousSyncErrors(v errors.ErrorList) {
	m.PreviousSyncErrors = v
}

// GetTotalEma returns the TotalEma value.
func (m SyncSourceReport) GetTotalEma() *float64 {
	return m.TotalEma
}

// SetTotalEma sets the TotalEma value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetTotalEma(v *float64) {
	m.TotalEma = v
}

// GetSuccessEma returns the SuccessEma value.
func (m SyncSourceReport) GetSuccessEma() *float64 {
	return m.SuccessEma
}

// SetSuccessEma sets the SuccessEma value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetSuccessEma(v *float64) {
	m.SuccessEma = v
}

// GetFailureEma returns the FailureEma value.
func (m SyncSourceReport) GetFailureEma() *float64 {
	return m.FailureEma
}

// SetFailureEma sets the FailureEma value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetFailureEma(v *float64) {
	m.FailureEma = v
}

// GetLastSyncDuration returns the LastSyncDuration value.
func (m SyncSourceReport) GetLastSyncDuration() *int64 {
	return m.LastSyncDuration
}

// SetLastSyncDuration sets the LastSyncDuration value and returns the updated SyncSourceReport.
func (m *SyncSourceReport) SetLastSyncDuration(v *int64) {
	m.LastSyncDuration = v
}

// GetSubdomain returns the Subdomain value.
func (m CreateIdentityParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetSubdomain(v string) CreateIdentityParams {
	m.Subdomain = v
	return m
}

// GetName returns the Name value.
func (m CreateIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetName(v string) CreateIdentityParams {
	m.Name = v
	return m
}

// GetPassword returns the Password value.
func (m CreateIdentityParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetPassword(v string) CreateIdentityParams {
	m.Password = v
	return m
}

// GetTotpSecret returns the TotpSecret value.
func (m CreateIdentityParams) GetTotpSecret() string {
	return m.TotpSecret
}

// SetTotpSecret sets the TotpSecret value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetTotpSecret(v string) CreateIdentityParams {
	m.TotpSecret = v
	return m
}

// GetRole returns the Role value.
func (m CreateIdentityParams) GetRole() consts.IdentityRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetRole(v consts.IdentityRole) CreateIdentityParams {
	m.Role = v
	return m
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateIdentityParams) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetCreatedByIdentityID(v int64) CreateIdentityParams {
	m.CreatedByIdentityID = v
	return m
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateIdentityParams.
func (m CreateIdentityParams) SetCreatedByIdentityName(v string) CreateIdentityParams {
	m.CreatedByIdentityName = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m CreateIfNotExistsIdentityParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetSubdomain(v string) CreateIfNotExistsIdentityParams {
	m.Subdomain = v
	return m
}

// GetName returns the Name value.
func (m CreateIfNotExistsIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetName(v string) CreateIfNotExistsIdentityParams {
	m.Name = v
	return m
}

// GetPassword returns the Password value.
func (m CreateIfNotExistsIdentityParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetPassword(v string) CreateIfNotExistsIdentityParams {
	m.Password = v
	return m
}

// GetTotpSecret returns the TotpSecret value.
func (m CreateIfNotExistsIdentityParams) GetTotpSecret() string {
	return m.TotpSecret
}

// SetTotpSecret sets the TotpSecret value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetTotpSecret(v string) CreateIfNotExistsIdentityParams {
	m.TotpSecret = v
	return m
}

// GetRole returns the Role value.
func (m CreateIfNotExistsIdentityParams) GetRole() consts.IdentityRole {
	return m.Role
}

// SetRole sets the Role value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetRole(v consts.IdentityRole) CreateIfNotExistsIdentityParams {
	m.Role = v
	return m
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m CreateIfNotExistsIdentityParams) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// SetCreatedByIdentityID sets the CreatedByIdentityID value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetCreatedByIdentityID(v int64) CreateIfNotExistsIdentityParams {
	m.CreatedByIdentityID = v
	return m
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m CreateIfNotExistsIdentityParams) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// SetCreatedByIdentityName sets the CreatedByIdentityName value and returns the updated CreateIfNotExistsIdentityParams.
func (m CreateIfNotExistsIdentityParams) SetCreatedByIdentityName(v string) CreateIfNotExistsIdentityParams {
	m.CreatedByIdentityName = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m AllIdentitiesParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated AllIdentitiesParams.
func (m AllIdentitiesParams) SetSubdomain(v string) AllIdentitiesParams {
	m.Subdomain = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m GetIdentityParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated GetIdentityParams.
func (m GetIdentityParams) SetSubdomain(v string) GetIdentityParams {
	m.Subdomain = v
	return m
}

// GetName returns the Name value.
func (m GetIdentityParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated GetIdentityParams.
func (m GetIdentityParams) SetName(v string) GetIdentityParams {
	m.Name = v
	return m
}

// GetLastSeenAt returns the LastSeenAt value.
func (m UpdateIdentityDateFlagsParams) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated UpdateIdentityDateFlagsParams.
func (m UpdateIdentityDateFlagsParams) SetLastSeenAt(v *int64) UpdateIdentityDateFlagsParams {
	m.LastSeenAt = v
	return m
}

// GetLastWriteAt returns the LastWriteAt value.
func (m UpdateIdentityDateFlagsParams) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated UpdateIdentityDateFlagsParams.
func (m UpdateIdentityDateFlagsParams) SetLastWriteAt(v *int64) UpdateIdentityDateFlagsParams {
	m.LastWriteAt = v
	return m
}

// GetID returns the ID value.
func (m UpdateIdentityDateFlagsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated UpdateIdentityDateFlagsParams.
func (m UpdateIdentityDateFlagsParams) SetID(v int64) UpdateIdentityDateFlagsParams {
	m.ID = v
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

// GetIds returns the Ids value.
func (m DeleteJobsByIdsParams) GetIds() []int64 {
	return m.Ids
}

// SetIds sets the Ids value and returns the updated DeleteJobsByIdsParams.
func (m DeleteJobsByIdsParams) SetIds(v []int64) DeleteJobsByIdsParams {
	m.Ids = v
	return m
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

// GetID returns the ID value.
func (m TruncateEventsParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated TruncateEventsParams.
func (m TruncateEventsParams) SetID(v int64) TruncateEventsParams {
	m.ID = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m GetAllSiteSyncSourcesForSiteParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated GetAllSiteSyncSourcesForSiteParams.
func (m GetAllSiteSyncSourcesForSiteParams) SetSubdomain(v string) GetAllSiteSyncSourcesForSiteParams {
	m.Subdomain = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m GetAllSiteSyncSourcesForUniqueHashParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated GetAllSiteSyncSourcesForUniqueHashParams.
func (m GetAllSiteSyncSourcesForUniqueHashParams) SetUniqueHash(v string) GetAllSiteSyncSourcesForUniqueHashParams {
	m.UniqueHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m CreateNightscoutSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated CreateNightscoutSyncSourceParams.
func (m CreateNightscoutSyncSourceParams) SetSubdomain(v string) CreateNightscoutSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetPassword returns the Password value.
func (m CreateNightscoutSyncSourceParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated CreateNightscoutSyncSourceParams.
func (m CreateNightscoutSyncSourceParams) SetPassword(v string) CreateNightscoutSyncSourceParams {
	m.Password = v
	return m
}

// GetEnabled returns the Enabled value.
func (m CreateNightscoutSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated CreateNightscoutSyncSourceParams.
func (m CreateNightscoutSyncSourceParams) SetEnabled(v bool) CreateNightscoutSyncSourceParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m CreateNightscoutSyncSourceParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated CreateNightscoutSyncSourceParams.
func (m CreateNightscoutSyncSourceParams) SetUrlOrUsername(v string) CreateNightscoutSyncSourceParams {
	m.UrlOrUsername = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m CreateNightscoutSyncSourceParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated CreateNightscoutSyncSourceParams.
func (m CreateNightscoutSyncSourceParams) SetUniqueHash(v string) CreateNightscoutSyncSourceParams {
	m.UniqueHash = v
	return m
}

// GetPassword returns the Password value.
func (m UpdateNightscoutSyncSourceParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated UpdateNightscoutSyncSourceParams.
func (m UpdateNightscoutSyncSourceParams) SetPassword(v string) UpdateNightscoutSyncSourceParams {
	m.Password = v
	return m
}

// GetEnabled returns the Enabled value.
func (m UpdateNightscoutSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateNightscoutSyncSourceParams.
func (m UpdateNightscoutSyncSourceParams) SetEnabled(v bool) UpdateNightscoutSyncSourceParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m UpdateNightscoutSyncSourceParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated UpdateNightscoutSyncSourceParams.
func (m UpdateNightscoutSyncSourceParams) SetUrlOrUsername(v string) UpdateNightscoutSyncSourceParams {
	m.UrlOrUsername = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateNightscoutSyncSourceParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateNightscoutSyncSourceParams.
func (m UpdateNightscoutSyncSourceParams) SetUniqueHash(v string) UpdateNightscoutSyncSourceParams {
	m.UniqueHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m UpdateNightscoutSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated UpdateNightscoutSyncSourceParams.
func (m UpdateNightscoutSyncSourceParams) SetSubdomain(v string) UpdateNightscoutSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetEnabled returns the Enabled value.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateNightscoutSyncSourceWithoutPasswordParams.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) SetEnabled(v bool) UpdateNightscoutSyncSourceWithoutPasswordParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated UpdateNightscoutSyncSourceWithoutPasswordParams.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) SetUrlOrUsername(v string) UpdateNightscoutSyncSourceWithoutPasswordParams {
	m.UrlOrUsername = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateNightscoutSyncSourceWithoutPasswordParams.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) SetUniqueHash(v string) UpdateNightscoutSyncSourceWithoutPasswordParams {
	m.UniqueHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated UpdateNightscoutSyncSourceWithoutPasswordParams.
func (m UpdateNightscoutSyncSourceWithoutPasswordParams) SetSubdomain(v string) UpdateNightscoutSyncSourceWithoutPasswordParams {
	m.Subdomain = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m DisableNightscoutSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated DisableNightscoutSyncSourceParams.
func (m DisableNightscoutSyncSourceParams) SetSubdomain(v string) DisableNightscoutSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m CreateDexcomSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetSubdomain(v string) CreateDexcomSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetPassword returns the Password value.
func (m CreateDexcomSyncSourceParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetPassword(v string) CreateDexcomSyncSourceParams {
	m.Password = v
	return m
}

// GetEnabled returns the Enabled value.
func (m CreateDexcomSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetEnabled(v bool) CreateDexcomSyncSourceParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m CreateDexcomSyncSourceParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetUrlOrUsername(v string) CreateDexcomSyncSourceParams {
	m.UrlOrUsername = v
	return m
}

// GetDexcomRegion returns the DexcomRegion value.
func (m CreateDexcomSyncSourceParams) GetDexcomRegion() consts.DexcomRegion {
	return m.DexcomRegion
}

// SetDexcomRegion sets the DexcomRegion value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetDexcomRegion(v consts.DexcomRegion) CreateDexcomSyncSourceParams {
	m.DexcomRegion = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m CreateDexcomSyncSourceParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated CreateDexcomSyncSourceParams.
func (m CreateDexcomSyncSourceParams) SetUniqueHash(v string) CreateDexcomSyncSourceParams {
	m.UniqueHash = v
	return m
}

// GetPassword returns the Password value.
func (m UpdateDexcomSyncSourceParams) GetPassword() string {
	return m.Password
}

// SetPassword sets the Password value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetPassword(v string) UpdateDexcomSyncSourceParams {
	m.Password = v
	return m
}

// GetEnabled returns the Enabled value.
func (m UpdateDexcomSyncSourceParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetEnabled(v bool) UpdateDexcomSyncSourceParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m UpdateDexcomSyncSourceParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetUrlOrUsername(v string) UpdateDexcomSyncSourceParams {
	m.UrlOrUsername = v
	return m
}

// GetDexcomRegion returns the DexcomRegion value.
func (m UpdateDexcomSyncSourceParams) GetDexcomRegion() consts.DexcomRegion {
	return m.DexcomRegion
}

// SetDexcomRegion sets the DexcomRegion value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetDexcomRegion(v consts.DexcomRegion) UpdateDexcomSyncSourceParams {
	m.DexcomRegion = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateDexcomSyncSourceParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetUniqueHash(v string) UpdateDexcomSyncSourceParams {
	m.UniqueHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m UpdateDexcomSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated UpdateDexcomSyncSourceParams.
func (m UpdateDexcomSyncSourceParams) SetSubdomain(v string) UpdateDexcomSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetEnabled returns the Enabled value.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) GetEnabled() bool {
	return m.Enabled
}

// SetEnabled sets the Enabled value and returns the updated UpdateDexcomSyncSourceWithoutPasswordParams.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) SetEnabled(v bool) UpdateDexcomSyncSourceWithoutPasswordParams {
	m.Enabled = v
	return m
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// SetUrlOrUsername sets the UrlOrUsername value and returns the updated UpdateDexcomSyncSourceWithoutPasswordParams.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) SetUrlOrUsername(v string) UpdateDexcomSyncSourceWithoutPasswordParams {
	m.UrlOrUsername = v
	return m
}

// GetDexcomRegion returns the DexcomRegion value.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) GetDexcomRegion() consts.DexcomRegion {
	return m.DexcomRegion
}

// SetDexcomRegion sets the DexcomRegion value and returns the updated UpdateDexcomSyncSourceWithoutPasswordParams.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) SetDexcomRegion(v consts.DexcomRegion) UpdateDexcomSyncSourceWithoutPasswordParams {
	m.DexcomRegion = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateDexcomSyncSourceWithoutPasswordParams.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) SetUniqueHash(v string) UpdateDexcomSyncSourceWithoutPasswordParams {
	m.UniqueHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated UpdateDexcomSyncSourceWithoutPasswordParams.
func (m UpdateDexcomSyncSourceWithoutPasswordParams) SetSubdomain(v string) UpdateDexcomSyncSourceWithoutPasswordParams {
	m.Subdomain = v
	return m
}

// GetType returns the Type value.
func (m DeleteSyncSourceParams) GetType() consts.SyncSourceType {
	return m.Type
}

// SetType sets the Type value and returns the updated DeleteSyncSourceParams.
func (m DeleteSyncSourceParams) SetType(v consts.SyncSourceType) DeleteSyncSourceParams {
	m.Type = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m DeleteSyncSourceParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated DeleteSyncSourceParams.
func (m DeleteSyncSourceParams) SetSubdomain(v string) DeleteSyncSourceParams {
	m.Subdomain = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m SubdomainExistsParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated SubdomainExistsParams.
func (m SubdomainExistsParams) SetSubdomain(v string) SubdomainExistsParams {
	m.Subdomain = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m CreateSiteParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated CreateSiteParams.
func (m CreateSiteParams) SetSubdomain(v string) CreateSiteParams {
	m.Subdomain = v
	return m
}

// GetBackupKeyHash returns the BackupKeyHash value.
func (m CreateSiteParams) GetBackupKeyHash() string {
	return m.BackupKeyHash
}

// SetBackupKeyHash sets the BackupKeyHash value and returns the updated CreateSiteParams.
func (m CreateSiteParams) SetBackupKeyHash(v string) CreateSiteParams {
	m.BackupKeyHash = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m GetSiteParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated GetSiteParams.
func (m GetSiteParams) SetSubdomain(v string) GetSiteParams {
	m.Subdomain = v
	return m
}

// GetJsonEach returns the JsonEach value.
func (m DeleteSitesParams) GetJsonEach() interface{} {
	return m.JsonEach
}

// SetJsonEach sets the JsonEach value and returns the updated DeleteSitesParams.
func (m DeleteSitesParams) SetJsonEach(v interface{}) DeleteSitesParams {
	m.JsonEach = v
	return m
}

// GetBatch returns the Batch value.
func (m GetBatchSitesParams) GetBatch() string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated GetBatchSitesParams.
func (m GetBatchSitesParams) SetBatch(v string) GetBatchSitesParams {
	m.Batch = v
	return m
}

// GetLastSeenAt returns the LastSeenAt value.
func (m GetInactiveSitesParams) GetLastSeenAt() float64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated GetInactiveSitesParams.
func (m GetInactiveSitesParams) SetLastSeenAt(v float64) GetInactiveSitesParams {
	m.LastSeenAt = v
	return m
}

// GetLastWriteAt returns the LastWriteAt value.
func (m GetInactiveSitesParams) GetLastWriteAt() float64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated GetInactiveSitesParams.
func (m GetInactiveSitesParams) SetLastWriteAt(v float64) GetInactiveSitesParams {
	m.LastWriteAt = v
	return m
}

// GetLastSeenAt returns the LastSeenAt value.
func (m GetNeverSeenSitesParams) GetLastSeenAt() float64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated GetNeverSeenSitesParams.
func (m GetNeverSeenSitesParams) SetLastSeenAt(v float64) GetNeverSeenSitesParams {
	m.LastSeenAt = v
	return m
}

// GetLastWriteAt returns the LastWriteAt value.
func (m GetNeverSeenSitesParams) GetLastWriteAt() float64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated GetNeverSeenSitesParams.
func (m GetNeverSeenSitesParams) SetLastWriteAt(v float64) GetNeverSeenSitesParams {
	m.LastWriteAt = v
	return m
}

// GetLastSeenAt returns the LastSeenAt value.
func (m UpdateSiteDateFlagsParams) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// SetLastSeenAt sets the LastSeenAt value and returns the updated UpdateSiteDateFlagsParams.
func (m UpdateSiteDateFlagsParams) SetLastSeenAt(v *int64) UpdateSiteDateFlagsParams {
	m.LastSeenAt = v
	return m
}

// GetLastWriteAt returns the LastWriteAt value.
func (m UpdateSiteDateFlagsParams) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// SetLastWriteAt sets the LastWriteAt value and returns the updated UpdateSiteDateFlagsParams.
func (m UpdateSiteDateFlagsParams) SetLastWriteAt(v *int64) UpdateSiteDateFlagsParams {
	m.LastWriteAt = v
	return m
}

// GetSubdomain returns the Subdomain value.
func (m UpdateSiteDateFlagsParams) GetSubdomain() string {
	return m.Subdomain
}

// SetSubdomain sets the Subdomain value and returns the updated UpdateSiteDateFlagsParams.
func (m UpdateSiteDateFlagsParams) SetSubdomain(v string) UpdateSiteDateFlagsParams {
	m.Subdomain = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateSyncReportAttemptParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateSyncReportAttemptParams.
func (m UpdateSyncReportAttemptParams) SetUniqueHash(v string) UpdateSyncReportAttemptParams {
	m.UniqueHash = v
	return m
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateSyncSourceReportsSuccessParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateSyncSourceReportsSuccessParams.
func (m UpdateSyncSourceReportsSuccessParams) SetElapsedMs(v *int64) UpdateSyncSourceReportsSuccessParams {
	m.ElapsedMs = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateSyncSourceReportsSuccessParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateSyncSourceReportsSuccessParams.
func (m UpdateSyncSourceReportsSuccessParams) SetUniqueHash(v string) UpdateSyncSourceReportsSuccessParams {
	m.UniqueHash = v
	return m
}

// GetErrorMessages returns the ErrorMessages value.
func (m UpdateSyncSourceReportsFailureParams) GetErrorMessages() errors.ErrorList {
	return m.ErrorMessages
}

// SetErrorMessages sets the ErrorMessages value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m UpdateSyncSourceReportsFailureParams) SetErrorMessages(v errors.ErrorList) UpdateSyncSourceReportsFailureParams {
	m.ErrorMessages = v
	return m
}

// GetElapsedMs returns the ElapsedMs value.
func (m UpdateSyncSourceReportsFailureParams) GetElapsedMs() *int64 {
	return m.ElapsedMs
}

// SetElapsedMs sets the ElapsedMs value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m UpdateSyncSourceReportsFailureParams) SetElapsedMs(v *int64) UpdateSyncSourceReportsFailureParams {
	m.ElapsedMs = v
	return m
}

// GetUniqueHash returns the UniqueHash value.
func (m UpdateSyncSourceReportsFailureParams) GetUniqueHash() string {
	return m.UniqueHash
}

// SetUniqueHash sets the UniqueHash value and returns the updated UpdateSyncSourceReportsFailureParams.
func (m UpdateSyncSourceReportsFailureParams) SetUniqueHash(v string) UpdateSyncSourceReportsFailureParams {
	m.UniqueHash = v
	return m
}

