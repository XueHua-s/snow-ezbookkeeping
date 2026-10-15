package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// UserUpdateAvatarHandler saves user avatar by request parameters for current user
func (a *UsersApi) UserUpdateAvatarHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] token type \"%d\" is not allowed to update user avatar", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserUpdateAvatarHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_UPDATE_AVATAR) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	form, err := c.MultipartForm()

	if err != nil {
		log.Errorf(c, "[users.UserUpdateAvatarHandler] failed to get multi-part form data for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrParameterInvalid
	}

	avatarFiles := form.File["avatar"]

	if len(avatarFiles) < 1 {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] there is no user avatar in request for user \"uid:%d\"", user.Uid)
		return nil, errs.ErrNoUserAvatar
	}

	if avatarFiles[0].Size < 1 {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] the size of user avatar in request is zero for user \"uid:%d\"", user.Uid)
		return nil, errs.ErrUserAvatarIsEmpty
	}

	if avatarFiles[0].Size > int64(a.CurrentConfig().MaxAvatarFileSize) {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] the upload file size \"%d\" exceeds the maximum size \"%d\" of user avatar for user \"uid:%d\"", avatarFiles[0].Size, a.CurrentConfig().MaxAvatarFileSize, uid)
		return nil, errs.ErrExceedMaxUserAvatarFileSize
	}

	fileExtension := utils.GetFileNameExtension(avatarFiles[0].Filename)

	if utils.GetImageContentType(fileExtension) == "" {
		log.Warnf(c, "[users.UserUpdateAvatarHandler] the file extension \"%s\" of user avatar in request is not supported for user \"uid:%d\"", fileExtension, user.Uid)
		return nil, errs.ErrImageTypeNotSupported
	}

	avatarFile, err := avatarFiles[0].Open()

	if err != nil {
		log.Errorf(c, "[users.UserUpdateAvatarHandler] failed to get avatar file from request for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	err = a.users.UpdateUserAvatar(c, user.Uid, avatarFile, fileExtension, user.CustomAvatarType)

	if err != nil {
		log.Errorf(c, "[users.UserUpdateAvatarHandler] failed to update avatar for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	user.CustomAvatarType = fileExtension
	userResp := a.getUserProfileResponse(user)
	return userResp, nil
}

// UserRemoveAvatarHandler removes user avatar by request parameters for current user
func (a *UsersApi) UserRemoveAvatarHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[users.UserRemoveAvatarHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[users.UserRemoveAvatarHandler] token type \"%d\" is not allowed to remove user avatar", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserRemoveAvatarHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_UPDATE_AVATAR) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	if user.CustomAvatarType == "" {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.users.RemoveUserAvatar(c, user.Uid, user.CustomAvatarType)

	if err != nil {
		log.Errorf(c, "[users.UserRemoveAvatarHandler] failed to remove avatar for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	user.CustomAvatarType = ""
	userResp := a.getUserProfileResponse(user)
	return userResp, nil
}
