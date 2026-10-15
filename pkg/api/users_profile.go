package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/locales"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strings"
	"time"
)

// UserUpdateProfileHandler saves user profile by request parameters for current user
func (a *UsersApi) UserUpdateProfileHandler(c *core.WebContext) (any, *errs.Error) {
	var userUpdateReq models.UserProfileUpdateRequest
	err := c.ShouldBindJSON(&userUpdateReq)

	if err != nil {
		log.Warnf(c, "[users.UserUpdateProfileHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	claims := c.GetTokenClaims()

	if claims == nil {
		log.Warnf(c, "[users.UserUpdateProfileHandler] current token is null")
		return nil, errs.ErrInvalidToken
	} else if claims.Type != core.USER_TOKEN_TYPE_NORMAL {
		log.Warnf(c, "[users.UserUpdateProfileHandler] token type \"%d\" is not allowed to update user profile", claims.Type)
		return nil, errs.ErrInvalidToken
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[users.UserUpdateProfileHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	userUpdateReq.Email = strings.TrimSpace(userUpdateReq.Email)
	userUpdateReq.Nickname = strings.TrimSpace(userUpdateReq.Nickname)

	modifyProfileBasicInfo := false
	modifyUseLastReconciledTime := false
	anythingUpdate := false
	userNew := &models.User{
		Uid:  user.Uid,
		Salt: user.Salt,
	}

	if userUpdateReq.Email != "" && userUpdateReq.Email != user.Email {
		if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_UPDATE_EMAIL) {
			return nil, errs.ErrNotPermittedToPerformThisAction
		}

		user.Email = userUpdateReq.Email
		userNew.Email = userUpdateReq.Email
		anythingUpdate = true
	}

	if userUpdateReq.Password != "" {
		if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_UPDATE_PASSWORD) {
			return nil, errs.ErrNotPermittedToPerformThisAction
		}

		if user.Password != "" && !a.users.IsPasswordEqualsUserPassword(userUpdateReq.OldPassword, user) {
			return nil, errs.ErrUserPasswordWrong
		}

		if !a.users.IsPasswordEqualsUserPassword(userUpdateReq.Password, user) {
			userNew.Password = userUpdateReq.Password
			anythingUpdate = true
		}
	}

	if userUpdateReq.Nickname != "" && userUpdateReq.Nickname != user.Nickname {
		user.Nickname = userUpdateReq.Nickname
		userNew.Nickname = userUpdateReq.Nickname
		modifyProfileBasicInfo = true
		anythingUpdate = true
	}

	if userUpdateReq.DefaultAccountId > 0 && userUpdateReq.DefaultAccountId != user.DefaultAccountId {
		accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, []int64{userUpdateReq.DefaultAccountId})

		if err != nil || len(accountMap) < 1 {
			return nil, errs.Or(err, errs.ErrUserDefaultAccountIsInvalid)
		}

		if _, exists := accountMap[userUpdateReq.DefaultAccountId]; !exists {
			log.Warnf(c, "[users.UserUpdateProfileHandler] account \"id:%d\" does not exist for user \"uid:%d\"", userUpdateReq.DefaultAccountId, uid)
			return nil, errs.ErrUserDefaultAccountIsInvalid
		}

		if accountMap[userUpdateReq.DefaultAccountId].Hidden {
			log.Warnf(c, "[users.UserUpdateProfileHandler] account \"id:%d\" is hidden of user \"uid:%d\"", userUpdateReq.DefaultAccountId, uid)
			return nil, errs.ErrUserDefaultAccountIsHidden
		}

		user.DefaultAccountId = userUpdateReq.DefaultAccountId
		userNew.DefaultAccountId = userUpdateReq.DefaultAccountId
		modifyProfileBasicInfo = true
		anythingUpdate = true
	}

	if userUpdateReq.UseLastReconciledTime != nil && *userUpdateReq.UseLastReconciledTime != user.UseLastReconciledTime {
		user.UseLastReconciledTime = *userUpdateReq.UseLastReconciledTime
		userNew.UseLastReconciledTime = *userUpdateReq.UseLastReconciledTime
		modifyProfileBasicInfo = true
		modifyUseLastReconciledTime = true
		anythingUpdate = true
	} else {
		modifyUseLastReconciledTime = false
	}

	if userUpdateReq.TransactionEditScope != nil && *userUpdateReq.TransactionEditScope != user.TransactionEditScope {
		user.TransactionEditScope = *userUpdateReq.TransactionEditScope
		userNew.TransactionEditScope = *userUpdateReq.TransactionEditScope
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.TransactionEditScope = models.TRANSACTION_EDIT_SCOPE_INVALID
	}

	modifyUserLanguage := false

	if userUpdateReq.Language != user.Language {
		user.Language = userUpdateReq.Language
		userNew.Language = userUpdateReq.Language
		modifyUserLanguage = true
		modifyProfileBasicInfo = true
		anythingUpdate = true
	}

	if userUpdateReq.DefaultCurrency != "" && userUpdateReq.DefaultCurrency != user.DefaultCurrency {
		user.DefaultCurrency = userUpdateReq.DefaultCurrency
		userNew.DefaultCurrency = userUpdateReq.DefaultCurrency
		modifyProfileBasicInfo = true
		anythingUpdate = true
	}

	if userUpdateReq.FirstDayOfWeek != nil && *userUpdateReq.FirstDayOfWeek != user.FirstDayOfWeek {
		user.FirstDayOfWeek = *userUpdateReq.FirstDayOfWeek
		userNew.FirstDayOfWeek = *userUpdateReq.FirstDayOfWeek
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.FirstDayOfWeek = core.WEEKDAY_INVALID
	}

	if userUpdateReq.FiscalYearStart != nil && *userUpdateReq.FiscalYearStart != user.FiscalYearStart {
		user.FiscalYearStart = *userUpdateReq.FiscalYearStart
		userNew.FiscalYearStart = *userUpdateReq.FiscalYearStart
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.FiscalYearStart = core.FISCAL_YEAR_START_INVALID
	}

	if userUpdateReq.CalendarDisplayType != nil && *userUpdateReq.CalendarDisplayType != user.CalendarDisplayType {
		user.CalendarDisplayType = *userUpdateReq.CalendarDisplayType
		userNew.CalendarDisplayType = *userUpdateReq.CalendarDisplayType
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.CalendarDisplayType = core.CALENDAR_DISPLAY_TYPE_INVALID
	}

	if userUpdateReq.DateDisplayType != nil && *userUpdateReq.DateDisplayType != user.DateDisplayType {
		user.DateDisplayType = *userUpdateReq.DateDisplayType
		userNew.DateDisplayType = *userUpdateReq.DateDisplayType
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.DateDisplayType = core.DATE_DISPLAY_TYPE_INVALID
	}

	if userUpdateReq.LongDateFormat != nil && *userUpdateReq.LongDateFormat != user.LongDateFormat {
		user.LongDateFormat = *userUpdateReq.LongDateFormat
		userNew.LongDateFormat = *userUpdateReq.LongDateFormat
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.LongDateFormat = core.LONG_DATE_FORMAT_INVALID
	}

	if userUpdateReq.ShortDateFormat != nil && *userUpdateReq.ShortDateFormat != user.ShortDateFormat {
		user.ShortDateFormat = *userUpdateReq.ShortDateFormat
		userNew.ShortDateFormat = *userUpdateReq.ShortDateFormat
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.ShortDateFormat = core.SHORT_DATE_FORMAT_INVALID
	}

	if userUpdateReq.LongTimeFormat != nil && *userUpdateReq.LongTimeFormat != user.LongTimeFormat {
		user.LongTimeFormat = *userUpdateReq.LongTimeFormat
		userNew.LongTimeFormat = *userUpdateReq.LongTimeFormat
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.LongTimeFormat = core.LONG_TIME_FORMAT_INVALID
	}

	if userUpdateReq.ShortTimeFormat != nil && *userUpdateReq.ShortTimeFormat != user.ShortTimeFormat {
		user.ShortTimeFormat = *userUpdateReq.ShortTimeFormat
		userNew.ShortTimeFormat = *userUpdateReq.ShortTimeFormat
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.ShortTimeFormat = core.SHORT_TIME_FORMAT_INVALID
	}

	if userUpdateReq.FiscalYearFormat != nil && *userUpdateReq.FiscalYearFormat != user.FiscalYearFormat {
		user.FiscalYearFormat = *userUpdateReq.FiscalYearFormat
		userNew.FiscalYearFormat = *userUpdateReq.FiscalYearFormat
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.FiscalYearFormat = core.FISCAL_YEAR_FORMAT_INVALID
	}

	if userUpdateReq.CurrencyDisplayType != nil && *userUpdateReq.CurrencyDisplayType != user.CurrencyDisplayType {
		user.CurrencyDisplayType = *userUpdateReq.CurrencyDisplayType
		userNew.CurrencyDisplayType = *userUpdateReq.CurrencyDisplayType
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.CurrencyDisplayType = core.CURRENCY_DISPLAY_TYPE_INVALID
	}

	if userUpdateReq.NumeralSystem != nil && *userUpdateReq.NumeralSystem != user.NumeralSystem {
		user.NumeralSystem = *userUpdateReq.NumeralSystem
		userNew.NumeralSystem = *userUpdateReq.NumeralSystem
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.NumeralSystem = core.NUMERAL_SYSTEM_INVALID
	}

	if userUpdateReq.DecimalSeparator != nil && *userUpdateReq.DecimalSeparator != user.DecimalSeparator {
		user.DecimalSeparator = *userUpdateReq.DecimalSeparator
		userNew.DecimalSeparator = *userUpdateReq.DecimalSeparator
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.DecimalSeparator = core.DECIMAL_SEPARATOR_INVALID
	}

	if userUpdateReq.DigitGroupingSymbol != nil && *userUpdateReq.DigitGroupingSymbol != user.DigitGroupingSymbol {
		user.DigitGroupingSymbol = *userUpdateReq.DigitGroupingSymbol
		userNew.DigitGroupingSymbol = *userUpdateReq.DigitGroupingSymbol
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.DigitGroupingSymbol = core.DIGIT_GROUPING_SYMBOL_INVALID
	}

	if userUpdateReq.DigitGrouping != nil && *userUpdateReq.DigitGrouping != user.DigitGrouping {
		user.DigitGrouping = *userUpdateReq.DigitGrouping
		userNew.DigitGrouping = *userUpdateReq.DigitGrouping
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.DigitGrouping = core.DIGIT_GROUPING_TYPE_INVALID
	}

	if userUpdateReq.CoordinateDisplayType != nil && *userUpdateReq.CoordinateDisplayType != user.CoordinateDisplayType {
		user.CoordinateDisplayType = *userUpdateReq.CoordinateDisplayType
		userNew.CoordinateDisplayType = *userUpdateReq.CoordinateDisplayType
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.CoordinateDisplayType = core.COORDINATE_DISPLAY_TYPE_INVALID
	}

	if userUpdateReq.ExpenseAmountColor != nil && *userUpdateReq.ExpenseAmountColor != user.ExpenseAmountColor {
		user.ExpenseAmountColor = *userUpdateReq.ExpenseAmountColor
		userNew.ExpenseAmountColor = *userUpdateReq.ExpenseAmountColor
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.ExpenseAmountColor = models.AMOUNT_COLOR_TYPE_INVALID
	}

	if userUpdateReq.IncomeAmountColor != nil && *userUpdateReq.IncomeAmountColor != user.IncomeAmountColor {
		user.IncomeAmountColor = *userUpdateReq.IncomeAmountColor
		userNew.IncomeAmountColor = *userUpdateReq.IncomeAmountColor
		modifyProfileBasicInfo = true
		anythingUpdate = true
	} else {
		userNew.IncomeAmountColor = models.AMOUNT_COLOR_TYPE_INVALID
	}

	if modifyProfileBasicInfo && user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_UPDATE_PROFILE_BASIC_INFO) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	if modifyUserLanguage || userNew.DecimalSeparator != core.DECIMAL_SEPARATOR_INVALID || userNew.DigitGroupingSymbol != core.DIGIT_GROUPING_SYMBOL_INVALID {
		decimalSeparator := userNew.DecimalSeparator
		digitGroupingSymbol := userNew.DigitGroupingSymbol

		if userNew.DecimalSeparator == core.DECIMAL_SEPARATOR_INVALID {
			decimalSeparator = user.DecimalSeparator
		}

		if userNew.DigitGroupingSymbol == core.DIGIT_GROUPING_SYMBOL_INVALID {
			digitGroupingSymbol = user.DigitGroupingSymbol
		}

		locale := user.Language

		if modifyUserLanguage {
			locale = userNew.Language
		}

		if locale == "" {
			locale = c.GetClientLocale()
		}

		if locales.IsDecimalSeparatorEqualsDigitGroupingSymbol(decimalSeparator, digitGroupingSymbol, locale) {
			return nil, errs.ErrDecimalSeparatorAndDigitGroupingSymbolCannotBeEqual
		}
	}

	if !anythingUpdate {
		return nil, errs.ErrNothingWillBeUpdated
	}

	keyProfileUpdated, emailSetToUnverified, err := a.users.UpdateUser(c, userNew, modifyUserLanguage, modifyUseLastReconciledTime)

	if err != nil {
		log.Errorf(c, "[users.UserUpdateProfileHandler] failed to update user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if emailSetToUnverified {
		user.EmailVerified = false
	}

	log.Infof(c, "[users.UserUpdateProfileHandler] user \"uid:%d\" has updated successfully", user.Uid)

	resp := &models.UserProfileUpdateResponse{
		User: a.GetUserBasicInfo(user),
	}

	if emailSetToUnverified && a.CurrentConfig().EnableUserVerifyEmail && a.CurrentConfig().EnableSMTP {
		err = a.tokens.DeleteTokensByType(c, uid, core.USER_TOKEN_TYPE_EMAIL_VERIFY)

		if err != nil {
			log.Errorf(c, "[users.UserUpdateProfileHandler] failed to revoke old email verify tokens for user \"uid:%d\", because %s", user.Uid, err.Error())
		} else {
			token, _, err := a.tokens.CreateEmailVerifyToken(c, user)

			if err != nil {
				log.Errorf(c, "[users.UserUpdateProfileHandler] failed to create email verify token for user \"uid:%d\", because %s", user.Uid, err.Error())
			} else {
				go func() {
					err = a.users.SendVerifyEmail(user, token, c.GetClientLocale())

					if err != nil {
						log.Warnf(c, "[users.UserUpdateProfileHandler] cannot send verify email to \"%s\", because %s", user.Email, err.Error())
					}
				}()
			}
		}
	}

	if keyProfileUpdated {
		now := time.Now().Unix()
		err = a.tokens.DeleteTokensBeforeTime(c, uid, now)

		if err == nil {
			log.Infof(c, "[users.UserUpdateProfileHandler] revoke old tokens before unix time \"%d\" for user \"uid:%d\"", now, user.Uid)
		} else {
			log.Warnf(c, "[users.UserUpdateProfileHandler] failed to revoke old tokens for user \"uid:%d\", because %s", user.Uid, err.Error())
		}

		token, claims, err := a.tokens.CreateToken(c, user)

		if err != nil {
			log.Warnf(c, "[users.UserUpdateProfileHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
			return resp, nil
		}

		resp.NewToken = token
		c.SetTextualToken(token)
		c.SetTokenClaims(claims)
		c.SetTokenContext("")

		log.Infof(c, "[users.UserUpdateProfileHandler] user \"uid:%d\" token refreshed, new token will be expired at %d", user.Uid, claims.ExpiresAt)

		return resp, nil
	}

	return resp, nil
}
