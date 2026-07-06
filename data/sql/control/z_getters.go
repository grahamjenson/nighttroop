
package control

import (
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
)


// GetID returns the ID value.
func (m *Admin) GetID() int64 {
	return m.ID
}

// GetName returns the Name value.
func (m *Admin) GetName() string {
	return m.Name
}

// GetPassword returns the Password value.
func (m *Admin) GetPassword() string {
	return m.Password
}

// GetTotpSecret returns the TotpSecret value.
func (m *Admin) GetTotpSecret() string {
	return m.TotpSecret
}

// GetRole returns the Role value.
func (m *Admin) GetRole() consts.AdminRole {
	return m.Role
}

// GetCreatedAt returns the CreatedAt value.
func (m *Admin) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetLastSeenAt returns the LastSeenAt value.
func (m *Admin) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// GetID returns the ID value.
func (m *DataSourceBlockList) GetID() int64 {
	return m.ID
}

// GetCreatedAt returns the CreatedAt value.
func (m *DataSourceBlockList) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUsername returns the Username value.
func (m *DataSourceBlockList) GetUsername() *string {
	return m.Username
}

// GetUrl returns the Url value.
func (m *DataSourceBlockList) GetUrl() *string {
	return m.Url
}

// GetReason returns the Reason value.
func (m *DataSourceBlockList) GetReason() *string {
	return m.Reason
}

// GetID returns the ID value.
func (m *Event) GetID() int64 {
	return m.ID
}

// GetType returns the Type value.
func (m *Event) GetType() *string {
	return m.Type
}

// GetDetails returns the Details value.
func (m *Event) GetDetails() *string {
	return m.Details
}

// GetStart returns the Start value.
func (m *Event) GetStart() int64 {
	return m.Start
}

// GetEnd returns the End value.
func (m *Event) GetEnd() *int64 {
	return m.End
}

// GetDayQueueJob returns the DayQueueJob value.
func (m *JobStatsDaily) GetDayQueueJob() string {
	return m.DayQueueJob
}

// GetTotalJobs returns the TotalJobs value.
func (m *JobStatsDaily) GetTotalJobs() int64 {
	return m.TotalJobs
}

// GetTotalSuccess returns the TotalSuccess value.
func (m *JobStatsDaily) GetTotalSuccess() int64 {
	return m.TotalSuccess
}

// GetTotalFailed returns the TotalFailed value.
func (m *JobStatsDaily) GetTotalFailed() int64 {
	return m.TotalFailed
}

// GetDurationSum returns the DurationSum value.
func (m *JobStatsDaily) GetDurationSum() float64 {
	return m.DurationSum
}

// GetDurationSquaredSum returns the DurationSquaredSum value.
func (m *JobStatsDaily) GetDurationSquaredSum() float64 {
	return m.DurationSquaredSum
}

// GetErrors returns the Errors value.
func (m *JobStatsDaily) GetErrors() errors.ErrorList {
	return m.Errors
}

// GetCreatedAt returns the CreatedAt value.
func (m *JobStatsDaily) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *JobStatsDaily) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetID returns the ID value.
func (m *Job) GetID() int64 {
	return m.ID
}

// GetQueue returns the Queue value.
func (m *Job) GetQueue() string {
	return m.Queue
}

// GetJob returns the Job value.
func (m *Job) GetJob() string {
	return m.Job
}

// GetJobStatus returns the JobStatus value.
func (m *Job) GetJobStatus() string {
	return m.JobStatus
}

// GetExecuteAfter returns the ExecuteAfter value.
func (m *Job) GetExecuteAfter() int64 {
	return m.ExecuteAfter
}

// GetRemainingAttempts returns the RemainingAttempts value.
func (m *Job) GetRemainingAttempts() int64 {
	return m.RemainingAttempts
}

// GetConsumerFetchedAt returns the ConsumerFetchedAt value.
func (m *Job) GetConsumerFetchedAt() int64 {
	return m.ConsumerFetchedAt
}

// GetFinishedAt returns the FinishedAt value.
func (m *Job) GetFinishedAt() int64 {
	return m.FinishedAt
}

// GetDedupingKey returns the DedupingKey value.
func (m *Job) GetDedupingKey() string {
	return m.DedupingKey
}

// GetErrors returns the Errors value.
func (m *Job) GetErrors() errors.ErrorList {
	return m.Errors
}

// GetCreatedAt returns the CreatedAt value.
func (m *Job) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *Job) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetID returns the ID value.
func (m *SiteIdentity) GetID() int64 {
	return m.ID
}

// GetSubdomain returns the Subdomain value.
func (m *SiteIdentity) GetSubdomain() string {
	return m.Subdomain
}

// GetName returns the Name value.
func (m *SiteIdentity) GetName() string {
	return m.Name
}

// GetPassword returns the Password value.
func (m *SiteIdentity) GetPassword() string {
	return m.Password
}

// GetTotpSecret returns the TotpSecret value.
func (m *SiteIdentity) GetTotpSecret() string {
	return m.TotpSecret
}

// GetRole returns the Role value.
func (m *SiteIdentity) GetRole() consts.IdentityRole {
	return m.Role
}

// GetCreatedAt returns the CreatedAt value.
func (m *SiteIdentity) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetLastSeenAt returns the LastSeenAt value.
func (m *SiteIdentity) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// GetLastWriteAt returns the LastWriteAt value.
func (m *SiteIdentity) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// GetUpdatedByIdentityID returns the UpdatedByIdentityID value.
func (m *SiteIdentity) GetUpdatedByIdentityID() *int64 {
	return m.UpdatedByIdentityID
}

// GetUpdatedByIdentityName returns the UpdatedByIdentityName value.
func (m *SiteIdentity) GetUpdatedByIdentityName() *string {
	return m.UpdatedByIdentityName
}

// GetCreatedByIdentityID returns the CreatedByIdentityID value.
func (m *SiteIdentity) GetCreatedByIdentityID() int64 {
	return m.CreatedByIdentityID
}

// GetCreatedByIdentityName returns the CreatedByIdentityName value.
func (m *SiteIdentity) GetCreatedByIdentityName() string {
	return m.CreatedByIdentityName
}

// GetDeletedAt returns the DeletedAt value.
func (m *SiteIdentity) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetEnabled returns the Enabled value.
func (m *SiteIdentity) GetEnabled() bool {
	return m.Enabled
}

// GetSubdomain returns the Subdomain value.
func (m *SiteSetting) GetSubdomain() string {
	return m.Subdomain
}

// GetKey returns the Key value.
func (m *SiteSetting) GetKey() string {
	return m.Key
}

// GetValue returns the Value value.
func (m *SiteSetting) GetValue() string {
	return m.Value
}

// GetDeletedAt returns the DeletedAt value.
func (m *SiteSetting) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetCreatedAt returns the CreatedAt value.
func (m *SiteSetting) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *SiteSetting) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetID returns the ID value.
func (m *SiteSyncSource) GetID() int64 {
	return m.ID
}

// GetSubdomain returns the Subdomain value.
func (m *SiteSyncSource) GetSubdomain() string {
	return m.Subdomain
}

// GetType returns the Type value.
func (m *SiteSyncSource) GetType() consts.SyncSourceType {
	return m.Type
}

// GetUrlOrUsername returns the UrlOrUsername value.
func (m *SiteSyncSource) GetUrlOrUsername() string {
	return m.UrlOrUsername
}

// GetPassword returns the Password value.
func (m *SiteSyncSource) GetPassword() string {
	return m.Password
}

// GetDexcomRegion returns the DexcomRegion value.
func (m *SiteSyncSource) GetDexcomRegion() consts.DexcomRegion {
	return m.DexcomRegion
}

// GetCreatedAt returns the CreatedAt value.
func (m *SiteSyncSource) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *SiteSyncSource) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetEnabled returns the Enabled value.
func (m *SiteSyncSource) GetEnabled() bool {
	return m.Enabled
}

// GetUniqueHash returns the UniqueHash value.
func (m *SiteSyncSource) GetUniqueHash() string {
	return m.UniqueHash
}

// GetID returns the ID value.
func (m *Site) GetID() int64 {
	return m.ID
}

// GetSubdomain returns the Subdomain value.
func (m *Site) GetSubdomain() string {
	return m.Subdomain
}

// GetBatch returns the Batch value.
func (m *Site) GetBatch() string {
	return m.Batch
}

// GetBackupKeyHash returns the BackupKeyHash value.
func (m *Site) GetBackupKeyHash() string {
	return m.BackupKeyHash
}

// GetCreatedAt returns the CreatedAt value.
func (m *Site) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetMaxDbSizeKbLimit returns the MaxDbSizeKbLimit value.
func (m *Site) GetMaxDbSizeKbLimit() *int64 {
	return m.MaxDbSizeKbLimit
}

// GetLastSeenAt returns the LastSeenAt value.
func (m *Site) GetLastSeenAt() *int64 {
	return m.LastSeenAt
}

// GetLastWriteAt returns the LastWriteAt value.
func (m *Site) GetLastWriteAt() *int64 {
	return m.LastWriteAt
}

// GetDeletedAt returns the DeletedAt value.
func (m *Site) GetDeletedAt() *int64 {
	return m.DeletedAt
}

// GetName returns the Name value.
func (m *SqliteSequence) GetName() interface{} {
	return m.Name
}

// GetSeq returns the Seq value.
func (m *SqliteSequence) GetSeq() interface{} {
	return m.Seq
}

// GetID returns the ID value.
func (m *SyncSourceReport) GetID() int64 {
	return m.ID
}

// GetCreatedAt returns the CreatedAt value.
func (m *SyncSourceReport) GetCreatedAt() int64 {
	return m.CreatedAt
}

// GetUpdatedAt returns the UpdatedAt value.
func (m *SyncSourceReport) GetUpdatedAt() int64 {
	return m.UpdatedAt
}

// GetUniqueHash returns the UniqueHash value.
func (m *SyncSourceReport) GetUniqueHash() string {
	return m.UniqueHash
}

// GetSyncIntervalSecondsReason returns the SyncIntervalSecondsReason value.
func (m *SyncSourceReport) GetSyncIntervalSecondsReason() string {
	return m.SyncIntervalSecondsReason
}

// GetSyncIntervalSeconds returns the SyncIntervalSeconds value.
func (m *SyncSourceReport) GetSyncIntervalSeconds() int64 {
	return m.SyncIntervalSeconds
}

// GetLastAttemptedSync returns the LastAttemptedSync value.
func (m *SyncSourceReport) GetLastAttemptedSync() *int64 {
	return m.LastAttemptedSync
}

// GetLastSuccessfulSync returns the LastSuccessfulSync value.
func (m *SyncSourceReport) GetLastSuccessfulSync() *int64 {
	return m.LastSuccessfulSync
}

// GetLastFailedSync returns the LastFailedSync value.
func (m *SyncSourceReport) GetLastFailedSync() *int64 {
	return m.LastFailedSync
}

// GetTotalSyncs returns the TotalSyncs value.
func (m *SyncSourceReport) GetTotalSyncs() *int64 {
	return m.TotalSyncs
}

// GetTotalSuccessfulSyncs returns the TotalSuccessfulSyncs value.
func (m *SyncSourceReport) GetTotalSuccessfulSyncs() *int64 {
	return m.TotalSuccessfulSyncs
}

// GetTotalFailedSyncs returns the TotalFailedSyncs value.
func (m *SyncSourceReport) GetTotalFailedSyncs() *int64 {
	return m.TotalFailedSyncs
}

// GetConsecutiveSuccessfulSyncs returns the ConsecutiveSuccessfulSyncs value.
func (m *SyncSourceReport) GetConsecutiveSuccessfulSyncs() *int64 {
	return m.ConsecutiveSuccessfulSyncs
}

// GetConsecutiveFailedSyncs returns the ConsecutiveFailedSyncs value.
func (m *SyncSourceReport) GetConsecutiveFailedSyncs() *int64 {
	return m.ConsecutiveFailedSyncs
}

// GetPreviousSyncErrors returns the PreviousSyncErrors value.
func (m *SyncSourceReport) GetPreviousSyncErrors() errors.ErrorList {
	return m.PreviousSyncErrors
}

// GetTotalEma returns the TotalEma value.
func (m *SyncSourceReport) GetTotalEma() *float64 {
	return m.TotalEma
}

// GetSuccessEma returns the SuccessEma value.
func (m *SyncSourceReport) GetSuccessEma() *float64 {
	return m.SuccessEma
}

// GetFailureEma returns the FailureEma value.
func (m *SyncSourceReport) GetFailureEma() *float64 {
	return m.FailureEma
}

// GetLastSyncDuration returns the LastSyncDuration value.
func (m *SyncSourceReport) GetLastSyncDuration() *int64 {
	return m.LastSyncDuration
}

