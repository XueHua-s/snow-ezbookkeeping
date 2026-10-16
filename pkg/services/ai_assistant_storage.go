package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"xorm.io/xorm"
)

// AssistantStorageService owns the per-user transaction boundary, not model calls.
type AssistantStorageService struct {
	ServiceUsingDB
	now    func() time.Time
	config func() settings.AIAssistantExperienceConfig
}

// AssistantStorageTx is valid only inside a WithUserTransaction callback. State
// is the locked user's accounting/epoch record and is committed with the callback's
// writes; neither this object nor Session may escape into asynchronous work.
// Database mutations must remain scoped to State.Uid, including direct SQL.
type AssistantStorageTx struct {
	Session          *xorm.Session
	State            *models.AssistantUserState
	Now              time.Time
	ReceiptRetention time.Duration
	Limits           models.AssistantLimits
	// Persisted version before this transaction acquired its write position.
	previousVersion int64
}

func NewAssistantStorageService(container *datastore.DataStoreContainer) *AssistantStorageService {
	return &AssistantStorageService{ServiceUsingDB: ServiceUsingDB{container: container}, now: time.Now, config: settings.DefaultAIAssistantExperienceConfig}
}

// WithUserRead always reads the primary, including command confirmation.
func (s *AssistantStorageService) WithUserRead(ctx core.Context, uid int64, read func(*xorm.Session) error) error {
	if uid <= 0 {
		return models.AssistantReasonNotAccessible
	}
	session := s.UserDataDB(uid).NewAssistantSession(ctx)
	defer session.Close()
	return assistantStorageSafeError(read(session))
}

var errAssistantStorageRetry = errors.New("assistant database concurrency retry")

// WithUserTransaction serializes writes using a database row CAS. The callback
// may run again after rollback; it must never perform external side effects.
func (s *AssistantStorageService) WithUserTransaction(ctx core.Context, uid int64, apply func(*AssistantStorageTx) error) error {
	_, err := s.withUserTransactionVersion(ctx, uid, apply)
	return err
}

// withUserTransactionVersion returns proof only after a successful commit.
// The locked pre-write version and committed version hide CAS bookkeeping from
// operations that need to detect intervening assistant writes.
func (s *AssistantStorageService) withUserTransactionVersion(ctx core.Context, uid int64, apply func(*AssistantStorageTx) error) (int64, error) {
	if uid <= 0 {
		return 0, models.AssistantReasonNotAccessible
	}
	for attempt := 0; attempt < 16; attempt++ {
		version, err := s.userTransaction(ctx, uid, apply)
		if err == nil {
			return version, nil
		}
		if !assistantStorageRetryable(err) {
			return 0, assistantStorageSafeError(err)
		}
		delay := time.NewTimer(time.Duration(attempt+1) * 5 * time.Millisecond)
		var done <-chan struct{}
		if ctx != nil {
			done = ctx.Done()
		}
		select {
		case <-done:
			delay.Stop()
			return 0, models.AssistantReasonStorageUnavailable
		case <-delay.C:
		}
	}
	return 0, models.AssistantReasonStorageUnavailable
}

func (s *AssistantStorageService) userTransaction(ctx core.Context, uid int64, apply func(*AssistantStorageTx) error) (int64, error) {
	session := s.UserDataDB(uid).NewAssistantSession(ctx)
	defer session.Close()
	if err := session.Begin(); err != nil {
		return 0, err
	}
	defer session.Rollback()
	// Acquire the database write position before reading: SQLite deferred
	// transactions otherwise let competing read snapshots deadlock on upgrade.
	affected, err := session.Where("uid = ?", uid).Incr("version", 1).Update(&models.AssistantUserState{})
	if err != nil {
		return 0, err
	}
	state := &models.AssistantUserState{}
	if affected == 0 {
		state = &models.AssistantUserState{Uid: uid, Version: 1, SettingsVersion: 1}
		if _, err = session.Insert(state); err != nil {
			return 0, err
		}
	} else {
		found, err := session.Where("uid = ?", uid).Get(state)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, errAssistantStorageRetry
		}
	}
	tx := &AssistantStorageTx{Session: session, State: state, Now: s.now().UTC(), ReceiptRetention: time.Duration(s.config().ReceiptRetentionHours) * time.Hour, Limits: s.config().Limits, previousVersion: state.Version - 1}
	if err := apply(tx); err != nil {
		return 0, err
	}
	if state.UsedBytes < 0 || state.ReservedBytes < 0 || state.ConversationCount < 0 || state.RunningCount < 0 || state.QueuedCount < 0 {
		return 0, models.AssistantReasonStorageUnavailable
	}
	state.UpdatedAt = tx.Now.UnixMilli()
	// MySQL reports changed rows, so even a read-only receipt replay must
	// advance the CAS version when committing this locked transaction.
	previousVersion := state.Version
	state.Version++
	affected, err = session.Where("uid = ? AND version = ?", uid, previousVersion).AllCols().Update(state)
	if err != nil {
		return 0, err
	}
	if affected != 1 {
		return 0, errAssistantStorageRetry
	}
	if err := session.Commit(); err != nil {
		return 0, err
	}
	return state.Version, nil
}

func (s *AssistantStorageService) GetUserState(ctx core.Context, uid int64) (*models.AssistantUserState, error) {
	var result models.AssistantUserState
	err := s.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		found, err := session.Where("uid = ?", uid).Get(&result)
		if !found && err == nil {
			result = models.AssistantUserState{Uid: uid, SettingsVersion: 1}
		}
		return err
	})
	return &result, err
}

func assistantStorageSafeError(err error) error {
	if err == nil {
		return nil
	}
	var reason models.AssistantReasonCode
	if errors.As(err, &reason) {
		return reason
	}
	return models.AssistantReasonStorageUnavailable
}

func assistantStorageRetryable(err error) bool {
	if errors.Is(err, errAssistantStorageRetry) {
		return true
	}
	var se sqlite3.Error
	if errors.As(err, &se) {
		return se.Code == sqlite3.ErrBusy || se.Code == sqlite3.ErrLocked || se.ExtendedCode == sqlite3.ErrConstraintPrimaryKey || se.ExtendedCode == sqlite3.ErrConstraintUnique
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062 || me.Number == 1213 || me.Number == 1205
	}
	var pe *pq.Error
	if errors.As(err, &pe) {
		return pe.Code == "23505" || pe.Code == "40001" || pe.Code == "40P01"
	}
	return !errors.Is(err, context.Canceled) && strings.Contains(err.Error(), "database is locked")
}

// SetConfigProvider supplies live validated configuration before the service is used.
func (s *AssistantStorageService) SetConfigProvider(provider func() settings.AIAssistantExperienceConfig) {
	s.config = provider
}
