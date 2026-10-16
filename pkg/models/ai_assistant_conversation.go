package models

// AssistantConversation stores metadata separately from bounded message bodies.
type AssistantConversation struct {
	Id                            int64               `xorm:"PK NOT NULL"`
	Uid                           int64               `xorm:"INDEX(ac_list) NOT NULL"`
	Deleted                       bool                `xorm:"INDEX(ac_list) INDEX(ac_cleanup_deleted) NOT NULL"`
	Archived                      bool                `xorm:"INDEX(ac_list) NOT NULL"`
	ActivityAt                    int64               `xorm:"INDEX(ac_list) NOT NULL"`
	Title                         string              `xorm:"VARCHAR(400) NOT NULL"`
	LastRunState                  AssistantRunState   `xorm:"VARCHAR(24)"`
	LastReasonCode                AssistantReasonCode `xorm:"VARCHAR(48)"`
	TitleManual                   bool
	CreatedAt                     int64
	DeletedAt                     int64
	MetadataVersion               int64
	ContextVersion                int64
	SourceVersion                 int64
	NextTurnSequence              int64
	ActiveRunId                   int64
	UsedBytes                     int64
	ReservedBytes                 int64
	HeaderUsedBytes               int64
	HeaderReservedBytes           int64
	TurnCount                     int
	Summary                       string `xorm:"TEXT"`
	ContextInvalidatedThroughTurn int64
	SummaryThrough                int64
	SummarySourceVersion          int64
	SummarySourcesJSON            string `xorm:"TEXT"`
	SeedContext                   string `xorm:"TEXT"`
	SeedSourcesJSON               string `xorm:"TEXT"`
	FrozenPrefixJSON              string `xorm:"TEXT"`
	SourceConversationId          int64
	SourceMessageId               int64
	SourceInvalidated             bool
	MemoryMode                    string `xorm:"VARCHAR(16)"`
	BranchKind                    string `xorm:"VARCHAR(16)"`
	PendingQuerySourceMessageId   int64
	PendingContinuation           bool
	SchemaVersion                 int
}
