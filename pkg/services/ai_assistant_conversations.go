package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"xorm.io/xorm"
)

type AssistantConversationOptions struct {
	Config              func() settings.AIAssistantExperienceConfig
	Runtime             func() settings.AIAssistantRuntimeConfig
	GenerateID          func(uuid.UuidType) int64
	Secret              []byte
	GenerationAvailable func(core.Context, int64) bool
}
type AssistantConversationService struct {
	storage *AssistantStorageService
	options AssistantConversationOptions
	now     func() time.Time
}

func NewAssistantConversationService(storage *AssistantStorageService, options AssistantConversationOptions) *AssistantConversationService {
	if options.Config == nil {
		options.Config = settings.DefaultAIAssistantExperienceConfig
	}
	if options.Runtime == nil {
		options.Runtime = (&settings.Config{}).GetAIAssistantRuntimeConfig
	}
	if options.GenerateID == nil {
		options.GenerateID = uuid.Container.GenerateUuid
	}
	return &AssistantConversationService{storage: storage, options: options, now: time.Now}
}
func (s *AssistantConversationService) limits() models.AssistantLimits {
	return s.options.Config().Limits
}
func assistantConversationLoad(session *xorm.Session, uid, id int64) (*models.AssistantConversation, error) {
	c := &models.AssistantConversation{}
	found, err := session.Where("uid = ? AND id = ? AND deleted = ?", uid, id, false).Get(c)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.AssistantReasonNotAccessible
	}
	return c, nil
}
func assistantConversationMessage(session *xorm.Session, uid, id int64) (*models.AssistantMessage, error) {
	m := &models.AssistantMessage{}
	found, err := session.Where("uid = ? AND id = ?", uid, id).Get(m)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.AssistantReasonNotAccessible
	}
	if _, err = assistantConversationLoad(session, uid, m.ConversationId); err != nil {
		return nil, err
	}
	return m, nil
}
func assistantConversationSave(tx *AssistantStorageTx, c *models.AssistantConversation) error {
	if !c.Deleted {
		if err := assistantSettleConversationHeader(tx, c); err != nil {
			return err
		}
	}
	_, err := tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, c.Id).AllCols().Update(c)
	return err
}
func assistantConversationPageLimit(requested, def, max int) int {
	if requested <= 0 {
		return def
	}
	if requested > max {
		return max
	}
	return requested
}
