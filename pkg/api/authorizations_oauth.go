package api

import (
	"encoding/json/v2"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// OAuth2CallbackAuthorizeHandler verifies and authorizes current OAuth 2.0 callback login
func (a *AuthorizationsApi) OAuth2CallbackAuthorizeHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableOAuth2Login {
		return nil, errs.ErrOAuth2NotEnabled
	}

	var credential models.OAuth2CallbackLoginRequest
	err := c.ShouldBindJSON(&credential)

	if err != nil {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	var tokenContext models.OAuth2CallbackTokenContext
	err = json.Unmarshal([]byte(c.GetTokenContext()), &tokenContext)

	if err != nil {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] parse token context failed, because %s", err.Error())
		return nil, errs.ErrOperationFailed
	}

	if !tokenContext.ExternalAuthType.IsValid() {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] external auth type \"%s\" is invalid", tokenContext.ExternalAuthType)
		return nil, errs.ErrInvalidOAuth2Provider
	}

	uid := c.GetCurrentUid()
	err = a.CheckFailureCount(c, uid)

	if err != nil {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to get user \"uid:%d\" info, because %s", uid, err.Error())
		return nil, errs.ErrUserNotFound
	}

	if user.Disabled {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] user \"uid:%d\" is disabled", user.Uid)
		return nil, errs.ErrUserIsDisabled
	}

	if a.CurrentConfig().EnableUserForceVerifyEmail && !user.EmailVerified {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] user \"uid:%d\" has not verified email", user.Uid)
		return nil, errs.ErrEmailIsNotVerified
	}

	oldTokenClaims := c.GetTokenClaims()

	if oldTokenClaims.Type == core.USER_TOKEN_TYPE_OAUTH2_CALLBACK_REQUIRE_VERIFY {
		if credential.Password == "" {
			return nil, errs.ErrPasswordIsEmpty
		}

		if !a.users.IsPasswordEqualsUserPassword(credential.Password, user) {
			failureCheckErr := a.CheckAndIncreaseFailureCount(c, uid)

			if failureCheckErr != nil {
				log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] cannot login for user \"uid:%d\", because %s", user.Uid, failureCheckErr.Error())
				return nil, errs.Or(failureCheckErr, errs.ErrFailureCountLimitReached)
			}

			return nil, errs.ErrUserPasswordWrong
		}

		if a.CurrentConfig().EnableTwoFactor {
			twoFactorSetting, err := a.twoFactorAuthorizations.GetUserTwoFactorSettingByUid(c, uid)

			if err != nil && !errors.Is(err, errs.ErrTwoFactorIsNotEnabled) {
				log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to check two-factor setting for user \"uid:%d\", because %s", user.Uid, err.Error())
				return nil, errs.Or(err, errs.ErrSystemError)
			}

			if twoFactorSetting != nil {
				if credential.Passcode == "" {
					return nil, errs.ErrPasscodeEmpty
				}

				if !a.twoFactorAuthorizations.ValidateTwoFactorPasscode(c, credential.Passcode, twoFactorSetting.Secret) {
					log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] passcode is invalid for user \"uid:%d\"", uid)

					err = a.CheckAndIncreaseFailureCount(c, uid)

					if err != nil {
						log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
						return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
					}

					return nil, errs.ErrPasscodeInvalid
				}

				passcodeUsed := a.update2FAPasscodeUsed(c, uid, credential.Passcode)

				if passcodeUsed {
					err = a.CheckAndIncreaseFailureCount(c, uid)

					if err != nil {
						log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] cannot auth for user \"uid:%d\", because %s", uid, err.Error())
						return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
					}

					return nil, errs.ErrPasscodeInvalid
				}
			}
		}

		userExternalAuth := &models.UserExternalAuth{
			Uid:              user.Uid,
			ExternalAuthType: tokenContext.ExternalAuthType,
			ExternalUsername: tokenContext.ExternalUsername,
			ExternalEmail:    tokenContext.ExternalEmail,
		}

		err = a.userExternalAuths.CreateUserExternalAuth(c, userExternalAuth)

		if err != nil {
			log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to create user external auth for user \"uid:%d\", because %s", user.Uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		log.Infof(c, "[authorizations.OAuth2CallbackAuthorizeHandler] user external auth has been created for user \"uid:%d\"", user.Uid)
	} else if oldTokenClaims.Type == core.USER_TOKEN_TYPE_OAUTH2_CALLBACK {
		_, err = a.userExternalAuths.GetUserExternalAuthByUid(c, uid, tokenContext.ExternalAuthType)

		if err != nil {
			log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to get user external auth for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrUserExternalAuthNotFound)
		}
	} else {
		return nil, errs.ErrSystemError
	}

	err = a.tokens.DeleteTokenByClaims(c, oldTokenClaims)

	if err != nil {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to revoke temporary token \"utid:%s\" for user \"uid:%d\", because %s", oldTokenClaims.UserTokenId, user.Uid, err.Error())
	}

	var token string
	var claims *core.UserTokenClaims

	if credential.Token != "" {
		_, claims, _, err = a.tokens.ParseToken(c, credential.Token)

		if err != nil {
			log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to parse token, because %s", err.Error())
			return nil, errs.ErrInvalidToken
		}

		if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
			log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] token type \"%d\" is not allowed to login via oauth 2.0", claims.Type)
			token = ""
			claims = nil
		} else if claims.Uid != user.Uid {
			log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] oauth 2.0 user \"uid:%d\" does not match current user \"uid:%d\"", user.Uid, claims.Uid)
			token = ""
			claims = nil
		} else {
			token = credential.Token
		}
	}

	if token == "" {
		token, claims, err = a.tokens.CreateToken(c, user)

		if err != nil {
			log.Errorf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
			return nil, errs.ErrTokenGenerating
		}
	}

	c.SetTextualToken(token)
	c.SetTokenClaims(claims)
	c.SetTokenContext("")

	userApplicationCloudSettings, err := a.userAppCloudSettings.GetUserApplicationCloudSettingsByUid(c, user.Uid)
	var applicationCloudSettingSlice *models.ApplicationCloudSettingSlice = nil

	if err != nil {
		log.Warnf(c, "[authorizations.OAuth2CallbackAuthorizeHandler] failed to get latest user application cloud settings for user \"uid:%d\", because %s", user.Uid, err.Error())
	} else if userApplicationCloudSettings != nil && len(userApplicationCloudSettings.Settings) > 0 {
		applicationCloudSettingSlice = &userApplicationCloudSettings.Settings
	}

	log.Infof(c, "[authorizations.OAuth2CallbackAuthorizeHandler] user \"uid:%d\" has logged in, token will be expired at %d", user.Uid, claims.ExpiresAt)

	authResp := a.getAuthResponse(c, token, false, user, applicationCloudSettingSlice)
	return authResp, nil
}
