package model

// TaskCallback is a private, durable notification outbox. It does not hold
// billing state; task polling remains the only authority for settlement.
type TaskCallback struct {
	ID              int64  `gorm:"primaryKey" json:"-"`
	TaskID          string `gorm:"type:varchar(191);uniqueIndex" json:"-"`
	TargetURL       string `gorm:"type:text" json:"-"`
	Token           string `gorm:"type:varchar(64)" json:"-"`
	CreatedAt       int64  `gorm:"index" json:"-"`
	DeliveredStatus string `gorm:"type:varchar(20)" json:"-"`
	Attempts        int    `json:"-"`
	NextAttemptAt   int64  `gorm:"index" json:"-"`
	LeaseUntil      int64  `json:"-"`
	RefreshAfter    int64  `json:"-"`
	ReadyAt         int64  `json:"-"`
}
