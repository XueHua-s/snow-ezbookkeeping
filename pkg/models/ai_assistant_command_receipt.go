package models

// AssistantCommandReceipt contains no user text or response body.
type AssistantCommandReceipt struct {
	Uid                   int64  `xorm:"PK NOT NULL"`
	ClientRequestId       string `xorm:"PK VARCHAR(128) NOT NULL"`
	CommandType           string `xorm:"VARCHAR(64) NOT NULL"`
	CanonicalRequestHash  string `xorm:"VARCHAR(64) NOT NULL"`
	RequestIssuedAt       string `xorm:"VARCHAR(40) NOT NULL"`
	ExpectedDeletionEpoch int64
	ResultJSON            string `xorm:"TEXT"`
	ConversationId        int64
	MemoryId              int64
	CreatedAt             int64
	ExpiresAt             int64 `xorm:"INDEX"`
	OperationCompleted    bool
	Tombstone             bool
}
