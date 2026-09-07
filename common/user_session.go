package common

// Defaults from upstream; session cleanup is not scheduled during migration.
const (
	DefaultUserSessionActiveLimit = 50
	DefaultUserSessionIssuanceLimit = 100
	DefaultUserSessionIssuanceWindowSeconds = 24 * 60 * 60
	DefaultUserSessionRevokedRetentionDays = 7
)

var (
	UserSessionActiveLimit = DefaultUserSessionActiveLimit
	UserSessionIssuanceLimit = DefaultUserSessionIssuanceLimit
	UserSessionIssuanceWindowSeconds = int64(DefaultUserSessionIssuanceWindowSeconds)
	UserSessionRevokedRetentionDays = DefaultUserSessionRevokedRetentionDays
)
