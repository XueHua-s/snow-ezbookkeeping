package cmd

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

func printUserInfo(user *models.User) {
	fmt.Printf("[Uid] %d\n", user.Uid)
	fmt.Printf("[Username] %s\n", user.Username)
	fmt.Printf("[Email] %s\n", user.Email)
	fmt.Printf("[Nickname] %s\n", user.Nickname)
	fmt.Printf("[Password] %s\n", user.Password)
	fmt.Printf("[Salt] %s\n", user.Salt)
	fmt.Printf("[DefaultAccountId] %d\n", user.DefaultAccountId)
	fmt.Printf("[UseLastReconciledTime] %t\n", user.UseLastReconciledTime)
	fmt.Printf("[TransactionEditScope] %s (%d)\n", user.TransactionEditScope, user.TransactionEditScope)
	fmt.Printf("[Language] %s\n", user.Language)
	fmt.Printf("[DefaultCurrency] %s\n", user.DefaultCurrency)
	fmt.Printf("[FirstDayOfWeek] %s (%d)\n", user.FirstDayOfWeek, user.FirstDayOfWeek)
	fmt.Printf("[FiscalYearStart] %s (%d)\n", user.FiscalYearStart, user.FiscalYearStart)
	fmt.Printf("[CalendarDisplayType] %s (%d)\n", user.CalendarDisplayType, user.CalendarDisplayType)
	fmt.Printf("[DateDisplayType] %s (%d)\n", user.DateDisplayType, user.DateDisplayType)
	fmt.Printf("[LongDateFormat] %s (%d)\n", user.LongDateFormat, user.LongDateFormat)
	fmt.Printf("[ShortDateFormat] %s (%d)\n", user.ShortDateFormat, user.ShortDateFormat)
	fmt.Printf("[LongTimeFormat] %s (%d)\n", user.LongTimeFormat, user.LongTimeFormat)
	fmt.Printf("[ShortTimeFormat] %s (%d)\n", user.ShortTimeFormat, user.ShortTimeFormat)
	fmt.Printf("[FiscalYearFormat] %s (%d)\n", user.FiscalYearFormat, user.FiscalYearFormat)
	fmt.Printf("[CurrencyDisplayType] %s (%d)\n", user.CurrencyDisplayType, user.CurrencyDisplayType)
	fmt.Printf("[NumeralSystem] %s (%d)\n", user.NumeralSystem, user.NumeralSystem)
	fmt.Printf("[DecimalSeparator] %s (%d)\n", user.DecimalSeparator, user.DecimalSeparator)
	fmt.Printf("[DigitGroupingSymbol] %s (%d)\n", user.DigitGroupingSymbol, user.DigitGroupingSymbol)
	fmt.Printf("[DigitGrouping] %s (%d)\n", user.DigitGrouping, user.DigitGrouping)
	fmt.Printf("[CoordinateDisplayType] %s (%d)\n", user.CoordinateDisplayType, user.CoordinateDisplayType)
	fmt.Printf("[ExpenseAmountColor] %s (%d)\n", user.ExpenseAmountColor, user.ExpenseAmountColor)
	fmt.Printf("[IncomeAmountColor] %s (%d)\n", user.IncomeAmountColor, user.IncomeAmountColor)
	fmt.Printf("[FeatureRestriction] %s (%d)\n", user.FeatureRestriction, user.FeatureRestriction)
	fmt.Printf("[Deleted] %t\n", user.Deleted)
	fmt.Printf("[EmailVerified] %t\n", user.EmailVerified)
	fmt.Printf("[CreatedAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(user.CreatedUnixTime), user.CreatedUnixTime)

	if user.UpdatedUnixTime > 0 {
		fmt.Printf("[UpdatedAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(user.UpdatedUnixTime), user.UpdatedUnixTime)
	}

	if user.DeletedUnixTime > 0 {
		fmt.Printf("[DeletedAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(user.DeletedUnixTime), user.DeletedUnixTime)
	}

	if user.LastLoginUnixTime > 0 {
		fmt.Printf("[LastLoginAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(user.LastLoginUnixTime), user.LastLoginUnixTime)
	}
}

func printTokenInfo(token *models.TokenRecord) {
	fmt.Printf("[CreatedAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(token.CreatedUnixTime), token.CreatedUnixTime)
	fmt.Printf("[ExpiredAt] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(token.ExpiredUnixTime), token.ExpiredUnixTime)
	fmt.Printf("[LastSeen] %s (%d)\n", utils.FormatUnixTimeToLongDateTimeInServerTimezone(token.LastSeenUnixTime), token.LastSeenUnixTime)
	fmt.Printf("[UserAgent] %s\n", token.UserAgent)
}
