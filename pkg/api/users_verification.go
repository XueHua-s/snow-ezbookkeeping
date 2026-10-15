package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// UserSendVerifyEmailByUnloginUserHandler sends unlogin user verify email
func (a *UsersApi) UserSendVerifyEmailByUnloginUserHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableUserVerifyEmail {
		return nil, errs.ErrEmailValidationNotAllowed
	}

	var userResendVerifyEmailReq models.UserResendVerifyEmailRequest
	err := c.ShouldBindJSON(&userResendVerifyEmailReq)

	user, err := a.users.GetUserByEmail(c, userResendVerifyEmailReq.Email)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(userResendVerifyEmailReq.Password, user) {
		log.Warnf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] request password not equals to the user password")
		return nil, errs.ErrUserPasswordWrong
	}

	if user.Disabled {
		log.Warnf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] user \"uid:%d\" is disabled", user.Uid)
		return nil, errs.ErrUserIsDisabled
	}

	if user.EmailVerified {
		log.Warnf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] user \"uid:%d\" email has been verified", user.Uid)
		return nil, errs.ErrEmailIsVerified
	}

	if !a.CurrentConfig().EnableSMTP {
		return nil, errs.ErrSMTPServerNotEnabled
	}

	token, _, err := a.tokens.CreateEmailVerifyToken(c, user)

	if err != nil {
		log.Errorf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrTokenGenerating
	}

	go func() {
		err = a.users.SendVerifyEmail(user, token, c.GetClientLocale())

		if err != nil {
			log.Warnf(c, "[users.UserSendVerifyEmailByUnloginUserHandler] cannot send email to \"%s\", because %s", user.Email, err.Error())
		}
	}()

	return true, nil
}

// UserSendVerifyEmailByLoginedUserHandler sends logined user verify email
func (a *UsersApi) UserSendVerifyEmailByLoginedUserHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableUserVerifyEmail {
		return nil, errs.ErrEmailValidationNotAllowed
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserSendVerifyEmailByLoginedUserHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.EmailVerified {
		log.Warnf(c, "[users.UserSendVerifyEmailByLoginedUserHandler] user \"uid:%d\" email has been verified", user.Uid)
		return nil, errs.ErrEmailIsVerified
	}

	if !a.CurrentConfig().EnableSMTP {
		return nil, errs.ErrSMTPServerNotEnabled
	}

	token, _, err := a.tokens.CreateEmailVerifyToken(c, user)

	if err != nil {
		log.Errorf(c, "[users.UserSendVerifyEmailByLoginedUserHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrTokenGenerating
	}

	go func() {
		err = a.users.SendVerifyEmail(user, token, c.GetClientLocale())

		if err != nil {
			log.Warnf(c, "[users.UserSendVerifyEmailByLoginedUserHandler] cannot send email to \"%s\", because %s", user.Email, err.Error())
		}
	}()

	return true, nil
}

// UserGetAvatarHandler returns user avatar data for current user
func (a *UsersApi) UserGetAvatarHandler(c *core.WebContext) ([]byte, string, *errs.Error) {
	fileName := c.Param("fileName")
	fileExtension := utils.GetFileNameExtension(fileName)
	contentType := utils.GetImageContentType(fileExtension)

	if contentType == "" {
		return nil, "", errs.ErrImageTypeNotSupported
	}

	uid := c.GetCurrentUid()
	fileBaseName := utils.GetFileNameWithoutExtension(fileName)

	if utils.Int64ToString(uid) != fileBaseName {
		log.Warnf(c, "[users.UserGetAvatarHandler] cannot get other user avatar \"uid:%s\" for user \"uid:%d\"", fileBaseName, uid)
		return nil, "", errs.ErrUserIdInvalid
	}

	avatarData, err := a.users.GetUserAvatar(c, uid, fileExtension)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserGetAvatarHandler] failed to get user avatar, because %s", err.Error())
		}

		return nil, "", errs.Or(err, errs.ErrOperationFailed)
	}

	return avatarData, contentType, nil
}

func (a *UsersApi) getUserProfileResponse(user *models.User) *models.UserProfileResponse {
	return user.ToUserProfileResponse(a.GetUserBasicInfo(user))
}
