package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/mail"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/storage"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
	"io"
)

const verifyEmailUrlFormat = "%sdesktop#/verify_email?token=%s"

// UserService represents user service
type UserService struct {
	ServiceUsingDB
	ServiceUsingConfig
	ServiceUsingMailer
	ServiceUsingUuid
	ServiceUsingStorage
}

// Initialize a user service singleton instance
var (
	Users = &UserService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingConfig: ServiceUsingConfig{
			container: settings.Container,
		},
		ServiceUsingMailer: ServiceUsingMailer{
			container: mail.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
		ServiceUsingStorage: ServiceUsingStorage{
			container: storage.Container,
		},
	}
)

// GetUserByUsernameOrEmailAndPassword returns the user model according to login name and password
func (s *UserService) GetUserByUsernameOrEmailAndPassword(c core.Context, loginname string, password string) (*models.User, int64, error) {
	var user *models.User
	var err error

	if utils.IsValidUsername(loginname) {
		user, err = s.GetUserByUsername(c, loginname)
	} else if utils.IsValidEmail(loginname) {
		user, err = s.GetUserByEmail(c, loginname)
	} else {
		err = errs.ErrLoginNameInvalid
	}

	if err != nil {
		return nil, 0, err
	}

	if user == nil {
		return nil, 0, errs.ErrUserNotFound
	}

	if !s.IsPasswordEqualsUserPassword(password, user) {
		return nil, user.Uid, errs.ErrUserPasswordWrong
	}

	return user, user.Uid, nil
}

// GetUserById returns the user model according to user uid
func (s *UserService) GetUserById(c core.Context, uid int64) (*models.User, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	user := &models.User{}
	has, err := s.UserDB().NewSession(c).ID(uid).Where("deleted=?", false).Get(user)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}

// GetUserByUsername returns the user model according to user name
func (s *UserService) GetUserByUsername(c core.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errs.ErrUsernameIsEmpty
	}

	user := &models.User{}
	has, err := s.UserDB().NewSession(c).Where("username=? AND deleted=?", username, false).Get(user)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}

// GetUserByEmail returns the user model according to user email
func (s *UserService) GetUserByEmail(c core.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, errs.ErrEmailIsEmpty
	}

	user := &models.User{}
	has, err := s.UserDB().NewSession(c).Where("email=? AND deleted=?", email, false).Get(user)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}

// GetUserAvatar returns the user avatar image data according to user uid
func (s *UserService) GetUserAvatar(c core.Context, uid int64, fileExtension string) ([]byte, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	user := &models.User{}
	has, err := s.UserDB().NewSession(c).ID(uid).Cols("uid", "deleted", "custom_avatar_type").Where("deleted=?", false).Get(user)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrUserNotFound
	}

	if user.CustomAvatarType == "" {
		return nil, errs.ErrUserAvatarNotSet
	}

	if user.CustomAvatarType != fileExtension {
		return nil, errs.ErrUserAvatarExtensionInvalid
	}

	avatarFile, exists, err := s.ReadAvatar(c, user.Uid, user.CustomAvatarType)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errs.ErrUserAvatarNoExists
	}

	defer avatarFile.Close()

	avatarData, err := io.ReadAll(avatarFile)

	if err != nil {
		return nil, err
	}

	return avatarData, nil
}
