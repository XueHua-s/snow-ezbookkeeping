package api

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/mayswind/ezbookkeeping/pkg/avatars"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"strings"
)

// UsersApi represents user api
type UsersApi struct {
	ApiUsingConfig
	ApiWithUserInfo
	users    *services.UserService
	tokens   *services.TokenService
	accounts *services.AccountService
}

// Initialize a user api singleton instance
var (
	Users = &UsersApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiWithUserInfo: ApiWithUserInfo{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			ApiUsingAvatarProvider: ApiUsingAvatarProvider{
				container: avatars.Container,
			},
		},
		users:    services.Users,
		tokens:   services.Tokens,
		accounts: services.Accounts,
	}
)

// UserRegisterHandler saves a new user by request parameters
func (a *UsersApi) UserRegisterHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableUserRegister {
		return nil, errs.ErrUserRegistrationNotAllowed
	}

	var userRegisterReq models.UserRegisterRequest
	err := c.ShouldBindBodyWith(&userRegisterReq, binding.JSON)

	if err != nil {
		log.Warnf(c, "[users.UserRegisterHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if userRegisterReq.DefaultCurrency == core.AccountCurrencyNotSetValue {
		log.Warnf(c, "[users.UserRegisterHandler] user default currency is invalid")
		return nil, errs.ErrUserDefaultCurrencyIsInvalid
	}

	userRegisterReq.Username = strings.TrimSpace(userRegisterReq.Username)
	userRegisterReq.Email = strings.TrimSpace(userRegisterReq.Email)
	userRegisterReq.Nickname = strings.TrimSpace(userRegisterReq.Nickname)

	user := &models.User{
		Username:             userRegisterReq.Username,
		Email:                userRegisterReq.Email,
		Nickname:             userRegisterReq.Nickname,
		Password:             userRegisterReq.Password,
		Language:             userRegisterReq.Language,
		DefaultCurrency:      userRegisterReq.DefaultCurrency,
		FirstDayOfWeek:       userRegisterReq.FirstDayOfWeek,
		FiscalYearStart:      core.FISCAL_YEAR_START_DEFAULT,
		TransactionEditScope: models.TRANSACTION_EDIT_SCOPE_ALL,
		FeatureRestriction:   a.CurrentConfig().DefaultFeatureRestrictions,
	}

	err = a.users.CreateUser(c, user, false)

	if err != nil {
		log.Errorf(c, "[users.UserRegisterHandler] failed to create user \"%s\", because %s", user.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[users.UserRegisterHandler] user \"%s\" has registered successfully, uid is %d", user.Username, user.Uid)

	presetCategoriesSaved := false

	if len(userRegisterReq.Categories) > 0 {
		_, err = TransactionCategories.createBatchCategories(c, user.Uid, &userRegisterReq.TransactionCategoryCreateBatchRequest)

		if err == nil {
			presetCategoriesSaved = true
		}
	}

	authResp := &models.RegisterResponse{
		AuthResponse: models.AuthResponse{
			Need2FA:             false,
			User:                a.GetUserBasicInfo(user),
			NotificationContent: a.GetAfterRegisterNotificationContent(user.Language, c.GetClientLocale()),
		},
		NeedVerifyEmail:       a.CurrentConfig().EnableUserVerifyEmail && a.CurrentConfig().EnableUserForceVerifyEmail,
		PresetCategoriesSaved: presetCategoriesSaved,
	}

	if a.CurrentConfig().EnableUserVerifyEmail && a.CurrentConfig().EnableSMTP {
		token, _, err := a.tokens.CreateEmailVerifyToken(c, user)

		if err != nil {
			log.Errorf(c, "[users.UserRegisterHandler] failed to create email verify token for user \"uid:%d\", because %s", user.Uid, err.Error())
		} else {
			go func() {
				err = a.users.SendVerifyEmail(user, token, c.GetClientLocale())

				if err != nil {
					log.Warnf(c, "[users.UserRegisterHandler] cannot send verify email to \"%s\", because %s", user.Email, err.Error())
				}
			}()
		}
	}

	if a.CurrentConfig().EnableUserForceVerifyEmail {
		return authResp, nil
	}

	token, claims, err := a.tokens.CreateToken(c, user)

	if err != nil {
		log.Warnf(c, "[users.UserRegisterHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return authResp, nil
	}

	authResp.Token = token
	c.SetTextualToken(token)
	c.SetTokenClaims(claims)
	c.SetTokenContext("")

	log.Infof(c, "[users.UserRegisterHandler] user \"uid:%d\" has logged in, token will be expired at %d", user.Uid, claims.ExpiresAt)

	return authResp, nil
}

// UserEmailVerifyHandler sets user email address verified
func (a *UsersApi) UserEmailVerifyHandler(c *core.WebContext) (any, *errs.Error) {
	var userVerifyEmailReq models.UserVerifyEmailRequest
	err := c.ShouldBindJSON(&userVerifyEmailReq)

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserEmailVerifyHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.Disabled {
		log.Warnf(c, "[users.UserEmailVerifyHandler] user \"uid:%d\" is disabled", user.Uid)
		return nil, errs.ErrUserIsDisabled
	}

	if user.EmailVerified {
		log.Warnf(c, "[users.UserEmailVerifyHandler] user \"uid:%d\" email has been verified", user.Uid)
		return nil, errs.ErrEmailIsVerified
	}

	err = a.users.SetUserEmailVerified(c, user.Username)

	if err != nil {
		log.Errorf(c, "[users.UserEmailVerifyHandler] failed to update user \"uid:%d\" email address verified, because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	err = a.tokens.DeleteTokensByType(c, uid, core.USER_TOKEN_TYPE_EMAIL_VERIFY)

	if err == nil {
		log.Infof(c, "[users.UserEmailVerifyHandler] revoke old email verify tokens for user \"uid:%d\"", user.Uid)
	} else {
		log.Warnf(c, "[users.UserEmailVerifyHandler] failed to revoke old email verify tokens for user \"uid:%d\", because %s", user.Uid, err.Error())
	}

	resp := &models.UserVerifyEmailResponse{}

	if userVerifyEmailReq.RequestNewToken {
		token, claims, err := a.tokens.CreateToken(c, user)

		if err != nil {
			log.Warnf(c, "[users.UserEmailVerifyHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
			return resp, nil
		}

		resp.NewToken = token
		resp.User = a.GetUserBasicInfo(user)
		resp.NotificationContent = a.GetAfterLoginNotificationContent(user.Language, c.GetClientLocale())

		c.SetTextualToken(token)
		c.SetTokenClaims(claims)
		c.SetTokenContext("")

		log.Infof(c, "[users.UserEmailVerifyHandler] user \"uid:%d\" token created, new token will be expired at %d", user.Uid, claims.ExpiresAt)
	}

	return resp, nil
}

// UserProfileHandler returns user profile of current user
func (a *UsersApi) UserProfileHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserRegisterHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	userResp := a.getUserProfileResponse(user)
	return userResp, nil
}
