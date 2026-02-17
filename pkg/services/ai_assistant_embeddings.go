package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// AIAssistantEmbeddingService represents ai assistant embedding cache service
type AIAssistantEmbeddingService struct {
	ServiceUsingDB
}

// Initialize an ai assistant embedding service singleton instance
var (
	AIAssistantEmbeddings = &AIAssistantEmbeddingService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
	}
)

// GetEmbeddingsByTransactionIds returns ai assistant embeddings by transaction ids
func (s *AIAssistantEmbeddingService) GetEmbeddingsByTransactionIds(c core.Context, uid int64, embeddingModel string, transactionIds []int64) (map[int64]*models.AIAssistantEmbedding, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if embeddingModel == "" {
		return nil, errs.ErrAIAssistantEmbeddingModelInvalid
	}

	if len(transactionIds) < 1 {
		return map[int64]*models.AIAssistantEmbedding{}, nil
	}

	var embeddings []*models.AIAssistantEmbedding
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND embedding_model=?", uid, embeddingModel).In("transaction_id", transactionIds).Find(&embeddings)

	if err != nil {
		return nil, err
	}

	embeddingMap := make(map[int64]*models.AIAssistantEmbedding, len(embeddings))

	for i := 0; i < len(embeddings); i++ {
		embedding := embeddings[i]

		if embedding == nil {
			continue
		}

		embeddingMap[embedding.TransactionId] = embedding
	}

	return embeddingMap, nil
}

// SaveEmbeddings saves ai assistant embeddings
func (s *AIAssistantEmbeddingService) SaveEmbeddings(c core.Context, embeddings []*models.AIAssistantEmbedding) error {
	if len(embeddings) < 1 {
		return nil
	}

	firstEmbedding := embeddings[0]

	if firstEmbedding == nil {
		return errs.ErrOperationFailed
	}

	uid := firstEmbedding.Uid
	embeddingModel := firstEmbedding.EmbeddingModel

	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if embeddingModel == "" {
		return errs.ErrAIAssistantEmbeddingModelInvalid
	}

	transactionIds := make([]int64, 0, len(embeddings))
	now := time.Now().Unix()

	for i := 0; i < len(embeddings); i++ {
		embedding := embeddings[i]

		if embedding == nil || embedding.Uid != uid || embedding.EmbeddingModel != embeddingModel || embedding.TransactionId <= 0 || embedding.VectorData == "" {
			return errs.ErrOperationFailed
		}

		if embedding.ContentHash == "" {
			return errs.ErrOperationFailed
		}

		embedding.UpdatedUnixTime = now

		if embedding.CreatedUnixTime == 0 {
			embedding.CreatedUnixTime = now
		}

		transactionIds = append(transactionIds, embedding.TransactionId)
	}

	transactionIds = utils.ToUniqueInt64Slice(transactionIds)

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=? AND embedding_model=?", uid, embeddingModel).In("transaction_id", transactionIds).Delete(&models.AIAssistantEmbedding{})

		if err != nil {
			return err
		}

		for i := 0; i < len(embeddings); i++ {
			_, err = sess.Insert(embeddings[i])

			if err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteEmbeddingsNotInTransactionIds deletes embeddings that are not in transaction ids
func (s *AIAssistantEmbeddingService) DeleteEmbeddingsNotInTransactionIds(c core.Context, uid int64, embeddingModel string, transactionIds []int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if embeddingModel == "" {
		return errs.ErrAIAssistantEmbeddingModelInvalid
	}

	sess := s.UserDataDB(uid).NewSession(c).Where("uid=? AND embedding_model=?", uid, embeddingModel)

	if len(transactionIds) > 0 {
		_, err := sess.NotIn("transaction_id", transactionIds).Delete(&models.AIAssistantEmbedding{})
		return err
	}

	_, err := sess.Delete(&models.AIAssistantEmbedding{})
	return err
}
