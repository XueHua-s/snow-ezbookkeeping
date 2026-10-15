package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/avatars"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// AuthorizationsApi represents authorization api
type AuthorizationsApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	ApiWithUserInfo
	users                   *services.UserService
	userAppCloudSettings    *services.UserApplicationCloudSettingsService
	tokens                  *services.TokenService
	twoFactorAuthorizations *services.TwoFactorAuthorizationService
	userExternalAuths       *services.UserExternalAuthService
}

// Initialize a authorization api singleton instance
var (
	Authorizations = &AuthorizationsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		ApiWithUserInfo: ApiWithUserInfo{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			ApiUsingAvatarProvider: ApiUsingAvatarProvider{
				container: avatars.Container,
			},
		},
		users:                   services.Users,
		userAppCloudSettings:    services.UserApplicationCloudSettings,
		tokens:                  services.Tokens,
		twoFactorAuthorizations: services.TwoFactorAuthorizations,
		userExternalAuths:       services.UserExternalAuths,
	}
)

// AuthorizeHandler verifies and authorizes current login request
func (a *AuthorizationsApi) AuthorizeHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableInternalAuth {
		return nil, errs.ErrCannotLoginByPassword
	}

	var credential models.UserLoginRequest
	err := c.ShouldBindJSON(&credential)

	if err != nil {
		log.Warnf(c, "[authorizations.AuthorizeHandler] parse request failed, because %s", err.Error())
		return nil, errs.ErrLoginNameOrPasswordInvalid
	}

	err = a.CheckFailureCount(c, 0)

	if err != nil {
		log.Warnf(c, "[authorizations.AuthorizeHandler] cannot login for user \"%s\", because %s", credential.LoginName, err.Error())
		return nil, errs.Or(err, errs.ErrFailureCountLimitReached)
	}

	user, uid, err := a.users.GetUserByUsernameOrEmailAndPassword(c, credential.LoginName, credential.Password)

	if errs.IsCustomError(err) {
		failureCheckErr := a.CheckAndIncreaseFailureCount(c, uid)

		if failureCheckErr != nil {
			log.Warnf(c, "[authorizations.AuthorizeHandler] cannot login for user \"%s\", because %s", credential.LoginName, failureCheckErr.Error())
			return nil, errs.Or(failureCheckErr, errs.ErrFailureCountLimitReached)
		}
	}

	if err != nil {
		log.Warnf(c, "[authorizations.AuthorizeHandler] login failed for user \"%s\", because %s", credential.LoginName, err.Error())
		return nil, errs.ErrLoginNameOrPasswordWrong
	}

	if user.Disabled {
		log.Warnf(c, "[authorizations.AuthorizeHandler] login failed for user \"%s\", because user is disabled", credential.LoginName)
		return nil, errs.ErrUserIsDisabled
	}

	if a.CurrentConfig().EnableUserForceVerifyEmail && !user.EmailVerified {
		hasValidEmailVerifyToken, err := a.tokens.ExistsValidTokenByType(c, user.Uid, core.USER_TOKEN_TYPE_EMAIL_VERIFY)

		if err != nil {
			log.Warnf(c, "[authorizations.AuthorizeHandler] failed check whether user \"uid:%d\" has valid verify email token, because %s", user.Uid, err.Error())
			hasValidEmailVerifyToken = false
		}

		log.Warnf(c, "[authorizations.AuthorizeHandler] login failed for user \"%s\", because user has not verified email", credential.LoginName)

		return nil, errs.NewErrorWithContext(errs.ErrEmailIsNotVerified, map[string]any{
			"email":                    user.Email,
			"hasValidEmailVerifyToken": hasValidEmailVerifyToken,
		})
	}

	err = a.users.UpdateUserLastLoginTime(c, user.Uid)

	if err != nil {
		log.Warnf(c, "[authorizations.AuthorizeHandler] failed to update last login time for user \"uid:%d\", because %s", user.Uid, err.Error())
	}

	twoFactorEnable := a.tokens.CurrentConfig().EnableTwoFactor

	if twoFactorEnable {
		twoFactorEnable, err = a.twoFactorAuthorizations.ExistsTwoFactorSetting(c, user.Uid)

		if err != nil {
			log.Errorf(c, "[authorizations.AuthorizeHandler] failed to check two-factor setting for user \"uid:%d\", because %s", user.Uid, err.Error())
			return nil, errs.Or(err, errs.ErrSystemError)
		}
	}

	var token string
	var claims *core.UserTokenClaims

	if twoFactorEnable {
		token, claims, err = a.tokens.CreateRequire2FAToken(c, user)
	} else {
		token, claims, err = a.tokens.CreateToken(c, user)
	}

	if err != nil {
		log.Errorf(c, "[authorizations.AuthorizeHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.ErrTokenGenerating
	}

	if !twoFactorEnable {
		c.SetTextualToken(token)
	}

	c.SetTokenClaims(claims)
	c.SetTokenContext("")

	userApplicationCloudSettings, err := a.userAppCloudSettings.GetUserApplicationCloudSettingsByUid(c, user.Uid)
	var applicationCloudSettingSlice *models.ApplicationCloudSettingSlice = nil

	if err != nil {
		log.Warnf(c, "[authorizations.AuthorizeHandler] failed to get latest user application cloud settings for user \"uid:%d\", because %s", user.Uid, err.Error())
	} else if userApplicationCloudSettings != nil && len(userApplicationCloudSettings.Settings) > 0 {
		applicationCloudSettingSlice = &userApplicationCloudSettings.Settings
	}

	log.Infof(c, "[authorizations.AuthorizeHandler] user \"uid:%d\" has logged in, token type is %d, token will be expired at %d", user.Uid, claims.Type, claims.ExpiresAt)

	authResp := a.getAuthResponse(c, token, twoFactorEnable, user, applicationCloudSettingSlice)
	return authResp, nil
}
