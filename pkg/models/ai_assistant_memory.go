package models

// AssistantMemory retains revocation metadata independently of removable text.
type AssistantMemory struct {
	Id              int64  `xorm:"PK NOT NULL"`
	Uid             int64  `xorm:"INDEX(amem_key) NOT NULL"`
	Key             string `xorm:"VARCHAR(128) INDEX(amem_key)"`
	Status          string `xorm:"VARCHAR(24) INDEX(amem_key) INDEX(amem_cleanup_status)"`
	Text            string `xorm:"TEXT"`
	Type            string `xorm:"VARCHAR(32)"`
	SourcesJSON     string `xorm:"TEXT"`
	CreatedAt       int64
	UpdatedAt       int64
	ExpiresAt       int64
	Version         int64
	Explicit        bool
	Excluded        bool
	GenerationEpoch int64
	UsedBytes       int64
	SchemaVersion   int
}
