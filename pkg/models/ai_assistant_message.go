package models

// AssistantMessage is an immutable question or one answer attempt (also its run).
type AssistantMessage struct {
	Id                         int64  `xorm:"PK NOT NULL"`
	Uid                        int64  `xorm:"INDEX(am_turn) UNIQUE(am_attempt) NOT NULL"`
	ConversationId             int64  `xorm:"INDEX(am_turn) NOT NULL"`
	TurnSequence               int64  `xorm:"INDEX(am_turn) NOT NULL"`
	Role                       string `xorm:"VARCHAR(16)"`
	UserMessageId              int64  `xorm:"UNIQUE(am_attempt) NOT NULL"`
	AttemptNumber              int    `xorm:"UNIQUE(am_attempt) NOT NULL"`
	Content                    string `xorm:"TEXT"`
	ProcessSummary             string `xorm:"TEXT"`
	CreatedAt                  int64
	CompletedAt                int64
	SelectedAssistantMessageId int64
	PreviousAssistantMessageId int64
	RunState                   AssistantRunState   `xorm:"VARCHAR(24) INDEX(am_work)"`
	Progress                   string              `xorm:"VARCHAR(32)"`
	ReasonCode                 AssistantReasonCode `xorm:"VARCHAR(48)"`
	QueriesJSON                string              `xorm:"TEXT"`
	ReferencesJSON             string              `xorm:"TEXT"`
	Continuation               string              `xorm:"TEXT"`
	ContinuationExpiresAt      int64
	ContinuationQueryId        string `xorm:"VARCHAR(128)"`
	ContinuationFilterJSON     string `xorm:"TEXT"`
	ClientRequestId            string `xorm:"VARCHAR(128)"`
	SnapshotSavedAt            int64
	SnapshotRevision           int64
	Owner                      string `xorm:"VARCHAR(128)"`
	LeaseEpoch                 int64
	LeaseExpiresAt             int64 `xorm:"INDEX(am_work)"`
	Deadline                   int64 `xorm:"INDEX(am_work)"`
	CancelRequested            bool
	InputDependenciesJSON      string `xorm:"TEXT"`
	RuntimeJSON                string `xorm:"TEXT"`
	LimitsJSON                 string `xorm:"TEXT"`
	ExpectedContextVersion     int64
	DeletionEpoch              int64
	UsedBytes                  int64
	ReservedBytes              int64
	ReservationSettled         bool
	QuotaState                 string `xorm:"VARCHAR(16)"`
	RestartQuery               bool
	SourceMessageId            int64
	Mode                       string `xorm:"VARCHAR(16)"`
	Locale                     string `xorm:"VARCHAR(32)"`
	Timezone                   string `xorm:"VARCHAR(128)"`
	SchemaVersion              int
}
