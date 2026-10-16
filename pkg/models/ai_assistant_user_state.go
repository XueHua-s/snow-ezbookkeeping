package models

// AssistantUserState is the database CAS boundary for all per-user mutations.
type AssistantUserState struct {
	Uid                      int64 `xorm:"PK NOT NULL"`
	UsedBytes                int64
	ReservedBytes            int64
	ConversationCount        int
	RunningCount             int
	QueuedCount              int
	ActiveMemoryCount        int
	CandidateMemoryCount     int
	AccountUnavailable       bool
	MemoryEnabled            bool
	SettingsVersion          int64
	MemoryGenerationEpoch    int64
	MemoryUseEpoch           int64
	DeletionEpoch            int64
	Version                  int64
	SubscriptionLeasesJSON   string `xorm:"TEXT"`
	BackgroundDay            int64
	BackgroundCalls          int
	BackgroundOwner          string `xorm:"VARCHAR(128)"`
	BackgroundLeaseExpiresAt int64
	UpdatedAt                int64
}
