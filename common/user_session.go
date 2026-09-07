package common

// Defaults from upstream; session cleanup is not scheduled during migration.
const (
	DefaultUserSessionIssuanceWindowSeconds = 24 * 60 * 60
	DefaultUserSessionRevokedRetentionDays = 7
)

var (
	UserSessionIssuanceWindowSeconds = int64(DefaultUserSessionIssuanceWindowSeconds)
	UserSessionRevokedRetentionDays = DefaultUserSessionRevokedRetentionDays
)
