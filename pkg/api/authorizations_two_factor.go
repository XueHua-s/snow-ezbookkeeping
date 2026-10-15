package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// TwoFactorAuthorizeHandler verifies and authorizes current 2fa login by passcode
func (a *AuthorizationsApi) TwoFactorAuthorizeHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableInternalAuth {
		return nil, errs.ErrCannotLoginByPassword
	}

	var credential models.TwoFactorLoginRequest
	err := c.ShouldBindJSON(&credential)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] parse request failed, because %s", err.Error())
		return nil, errs.ErrPasscodeInvalid
	}

	uid := c.GetCurrentUid()
	err = a.CheckFailureCount(c, uid)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
	}

	twoFactorSetting, err := a.twoFactorAuthorizations.GetUserTwoFactorSettingByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[authorizations.TwoFactorAuthorizeHandler] failed to get two-factor setting for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrSystemError)
	}

	if !a.twoFactorAuthorizations.ValidateTwoFactorPasscode(c, credential.Passcode, twoFactorSetting.Secret) {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] passcode is invalid for user \"uid:%d\"", uid)

		err = a.CheckAndIncreaseFailureCount(c, uid)

		if err != nil {
			log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
		}

		return nil, errs.ErrPasscodeInvalid
	}

	passcodeUsed := a.update2FAPasscodeUsed(c, uid, credential.Passcode)

	if passcodeUsed {
		err = a.CheckAndIncreaseFailureCount(c, uid)

		if err != nil {
			log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
		}

		return nil, errs.ErrPasscodeInvalid
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		log.Errorf(c, "[authorizations.TwoFactorAuthorizeHandler] failed to get user \"uid:%d\" info, because %s", uid, err.Error())
		return nil, errs.ErrUserNotFound
	}

	if user.Disabled {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] user \"uid:%d\" is disabled", user.Uid)
		return nil, errs.ErrUserIsDisabled
	}

	if a.CurrentConfig().EnableUserForceVerifyEmail && !user.EmailVerified {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] user \"uid:%d\" has not verified email", user.Uid)
		return nil, errs.ErrEmailIsNotVerified
	}

	oldTokenClaims := c.GetTokenClaims()
	err = a.tokens.DeleteTokenByClaims(c, oldTokenClaims)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] failed to revoke temporary token \"utid:%s\" for user \"uid:%d\", because %s", oldTokenClaims.UserTokenId, user.Uid, err.Error())
	}

	token, claims, err := a.tokens.CreateToken(c, user)

	if err != nil {
		log.Errorf(c, "[authorizations.TwoFactorAuthorizeHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrTokenGenerating
	}

	c.SetTextualToken(token)
	c.SetTokenClaims(claims)
	c.SetTokenContext("")

	userApplicationCloudSettings, err := a.userAppCloudSettings.GetUserApplicationCloudSettingsByUid(c, user.Uid)
	var applicationCloudSettingSlice *models.ApplicationCloudSettingSlice = nil

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeHandler] failed to get latest user application cloud settings for user \"uid:%d\", because %s", user.Uid, err.Error())
	} else if userApplicationCloudSettings != nil && len(userApplicationCloudSettings.Settings) > 0 {
		applicationCloudSettingSlice = &userApplicationCloudSettings.Settings
	}

	log.Infof(c, "[authorizations.TwoFactorAuthorizeHandler] user \"uid:%d\" has authorized two-factor via passcode, token will be expired at %d", user.Uid, claims.ExpiresAt)

	authResp := a.getAuthResponse(c, token, false, user, applicationCloudSettingSlice)
	return authResp, nil
}

// TwoFactorAuthorizeByRecoveryCodeHandler verifies and authorizes current 2fa login by recovery code
func (a *AuthorizationsApi) TwoFactorAuthorizeByRecoveryCodeHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableInternalAuth {
		return nil, errs.ErrCannotLoginByPassword
	}

	var credential models.TwoFactorRecoveryCodeLoginRequest
	err := c.ShouldBindJSON(&credential)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] parse request failed, because %s", err.Error())
		return nil, errs.ErrTwoFactorRecoveryCodeInvalid
	}

	uid := c.GetCurrentUid()
	err = a.CheckFailureCount(c, uid)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
	}

	enableTwoFactor, err := a.twoFactorAuthorizations.ExistsTwoFactorSetting(c, uid)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to get two-factor setting for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrSystemError)
	}

	if !enableTwoFactor {
		return nil, errs.ErrTwoFactorIsNotEnabled
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		log.Errorf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to get user \"uid:%d\" info, because %s", uid, err.Error())
		return nil, errs.ErrUserNotFound
	}

	if user.Disabled {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] user \"uid:%d\" is disabled", user.Uid)
		return nil, errs.ErrUserIsDisabled
	}

	if a.CurrentConfig().EnableUserForceVerifyEmail && !user.EmailVerified {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] user \"uid:%d\" has not verified email", user.Uid)
		return nil, errs.ErrEmailIsNotVerified
	}

	err = a.twoFactorAuthorizations.UseUserTwoFactorRecoveryCode(c, uid, credential.RecoveryCode, user.Salt)

	if errs.IsCustomError(err) {
		failureCheckErr := a.CheckAndIncreaseFailureCount(c, uid)

		if failureCheckErr != nil {
			log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] cannot auth for user \"uid:%d\", because %s", uid, failureCheckErr.Error())
			return nil, errs.Or(failureCheckErr, errs.ErrFailureCountLimitReached)
		}
	}

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to get two-factor recovery code for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrTwoFactorRecoveryCodeNotExist)
	}

	oldTokenClaims := c.GetTokenClaims()
	err = a.tokens.DeleteTokenByClaims(c, oldTokenClaims)

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to revoke temporary token \"utid:%s\" for user \"uid:%d\", because %s", oldTokenClaims.UserTokenId, user.Uid, err.Error())
	}

	token, claims, err := a.tokens.CreateToken(c, user)

	if err != nil {
		log.Errorf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrTokenGenerating
	}

	c.SetTextualToken(token)
	c.SetTokenClaims(claims)
	c.SetTokenContext("")

	userApplicationCloudSettings, err := a.userAppCloudSettings.GetUserApplicationCloudSettingsByUid(c, user.Uid)
	var applicationCloudSettingSlice *models.ApplicationCloudSettingSlice = nil

	if err != nil {
		log.Warnf(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] failed to get latest user application cloud settings for user \"uid:%d\", because %s", user.Uid, err.Error())
	} else if userApplicationCloudSettings != nil && len(userApplicationCloudSettings.Settings) > 0 {
		applicationCloudSettingSlice = &userApplicationCloudSettings.Settings
	}

	log.Infof(c, "[authorizations.TwoFactorAuthorizeByRecoveryCodeHandler] user \"uid:%d\" has authorized two-factor via recovery code \"%s\", token will be expired at %d", user.Uid, credential.RecoveryCode, claims.ExpiresAt)

	authResp := a.getAuthResponse(c, token, false, user, applicationCloudSettingSlice)
	return authResp, nil
}
