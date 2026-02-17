package models

// AIAssistantEmbedding represents one cached embedding vector for transaction knowledge
type AIAssistantEmbedding struct {
	Uid             int64  `xorm:"PK INDEX(IDX_ai_assistant_embedding_uid_model_updated_time) INDEX(IDX_ai_assistant_embedding_uid_model_transaction_id) NOT NULL"`
	TransactionId   int64  `xorm:"PK INDEX(IDX_ai_assistant_embedding_uid_model_transaction_id) NOT NULL"`
	EmbeddingModel  string `xorm:"PK VARCHAR(128) INDEX(IDX_ai_assistant_embedding_uid_model_updated_time) INDEX(IDX_ai_assistant_embedding_uid_model_transaction_id) NOT NULL"`
	ContentHash     string `xorm:"VARCHAR(64) NOT NULL"`
	VectorData      string `xorm:"TEXT NOT NULL"`
	CreatedUnixTime int64  `xorm:"NOT NULL"`
	UpdatedUnixTime int64  `xorm:"INDEX(IDX_ai_assistant_embedding_uid_model_updated_time) NOT NULL"`
}
