package services

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"time"
	"xorm.io/xorm"
)

// UpdateUserLastLoginTime updates the last login time field
func (s *UserService) UpdateUserLastLoginTime(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.UserDB().DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.ID(uid).Cols("last_login_unix_time").Where("deleted=?", false).Update(&models.User{LastLoginUnixTime: time.Now().Unix()})
		return err
	})
}

// EnableUser sets user enabled
func (s *UserService) EnableUser(c core.Context, username string) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	user, lookupErr := s.GetUserByUsername(c, username)
	if lookupErr != nil {
		return lookupErr
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		Disabled:        false,
		UpdatedUnixTime: now,
	}

	updatedRows, err := s.UserDB().NewSession(c).Cols("disabled", "updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return s.fenceAssistantAccount(c, user.Uid, false, false)
}

// DisableUser sets user disabled
func (s *UserService) DisableUser(c core.Context, username string) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	user, lookupErr := s.GetUserByUsername(c, username)
	if lookupErr != nil {
		return lookupErr
	}
	if err := s.fenceAssistantAccount(c, user.Uid, true, false); err != nil {
		return err
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		Disabled:        true,
		UpdatedUnixTime: now,
	}

	updatedRows, err := s.UserDB().NewSession(c).Cols("disabled", "updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// UpdateUserFeatureRestriction sets user user feature restrictions
func (s *UserService) UpdateUserFeatureRestriction(c core.Context, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		FeatureRestriction: featureRestriction,
		UpdatedUnixTime:    now,
	}

	updatedRows, err := s.UserDB().NewSession(c).Cols("feature_restriction", "updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// AddUserFeatureRestriction adds user user feature restrictions
func (s *UserService) AddUserFeatureRestriction(c core.Context, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		FeatureRestriction: featureRestriction,
		UpdatedUnixTime:    now,
	}

	updatedRows, err := s.UserDB().NewSession(c).SetExpr("feature_restriction", fmt.Sprintf("feature_restriction|(%d)", updateModel.FeatureRestriction)).Cols("updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// RemoveUserFeatureRestriction removes user user feature restrictions
func (s *UserService) RemoveUserFeatureRestriction(c core.Context, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		FeatureRestriction: ^featureRestriction,
		UpdatedUnixTime:    now,
	}

	updatedRows, err := s.UserDB().NewSession(c).SetExpr("feature_restriction", fmt.Sprintf("feature_restriction&(%d)", updateModel.FeatureRestriction)).Cols("updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// SetUserEmailVerified sets user email address verified
func (s *UserService) SetUserEmailVerified(c core.Context, username string) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		EmailVerified:   true,
		UpdatedUnixTime: now,
	}

	updatedRows, err := s.UserDB().NewSession(c).Cols("email_verified", "updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// SetUserEmailUnverified sets user email address unverified
func (s *UserService) SetUserEmailUnverified(c core.Context, username string) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		EmailVerified:   false,
		UpdatedUnixTime: now,
	}

	updatedRows, err := s.UserDB().NewSession(c).Cols("email_verified", "updated_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}

// DeleteUser deletes an existed user from database
func (s *UserService) DeleteUser(c core.Context, username string) error {
	if username == "" {
		return errs.ErrUsernameIsEmpty
	}

	user, lookupErr := s.GetUserByUsername(c, username)
	if lookupErr != nil {
		return lookupErr
	}
	if err := s.fenceAssistantAccount(c, user.Uid, true, true); err != nil {
		return err
	}

	now := time.Now().Unix()

	updateModel := &models.User{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	deletedRows, err := s.UserDB().NewSession(c).Cols("deleted", "deleted_unix_time").Where("username=? AND deleted=?", username, false).Update(updateModel)

	if err != nil {
		return err
	} else if deletedRows < 1 {
		return errs.ErrUserNotFound
	}
	return nil
}
