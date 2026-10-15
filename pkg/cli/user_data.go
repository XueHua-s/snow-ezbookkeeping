package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/validators"
	"time"
)

const pageCountForGettingTransactions = 1000

const pageCountForDataExport = 1000

// UserDataCli represents user data cli
type UserDataCli struct {
	CliUsingConfig
	accounts                *services.AccountService
	transactions            *services.TransactionService
	categories              *services.TransactionCategoryService
	tags                    *services.TransactionTagService
	users                   *services.UserService
	twoFactorAuthorizations *services.TwoFactorAuthorizationService
	tokens                  *services.TokenService
	forgetPasswords         *services.ForgetPasswordService
}

// Initialize a user data cli singleton instance
var (
	UserData = &UserDataCli{
		CliUsingConfig: CliUsingConfig{
			container: settings.Container,
		},
		accounts:                services.Accounts,
		transactions:            services.Transactions,
		categories:              services.TransactionCategories,
		tags:                    services.TransactionTags,
		users:                   services.Users,
		twoFactorAuthorizations: services.TwoFactorAuthorizations,
		tokens:                  services.Tokens,
		forgetPasswords:         services.ForgetPasswords,
	}
)

// AddNewUser adds a new user according to specified info
func (l *UserDataCli) AddNewUser(c *core.CliContext, username string, email string, nickname string, password string, defaultCurrency string) (*models.User, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.AddNewUser] user name is empty")
		return nil, errs.ErrUsernameIsEmpty
	}

	if email == "" {
		log.CliErrorf(c, "[user_data.AddNewUser] user email is empty")
		return nil, errs.ErrEmailIsEmpty
	}

	if nickname == "" {
		log.CliErrorf(c, "[user_data.AddNewUser] user nickname is empty")
		return nil, errs.ErrNicknameIsEmpty
	}

	if password == "" {
		log.CliErrorf(c, "[user_data.AddNewUser] user password is empty")
		return nil, errs.ErrPasswordIsEmpty
	}

	if defaultCurrency == "" {
		log.CliErrorf(c, "[user_data.AddNewUser] user default currency is empty")
		return nil, errs.ErrUserDefaultCurrencyIsEmpty
	}

	if _, ok := validators.AllCurrencyNames[defaultCurrency]; !ok {
		log.CliErrorf(c, "[user_data.AddNewUser] user default currency is invalid")
		return nil, errs.ErrUserDefaultCurrencyIsInvalid
	}

	user := &models.User{
		Username:             username,
		Email:                email,
		Nickname:             nickname,
		Password:             password,
		DefaultCurrency:      defaultCurrency,
		FirstDayOfWeek:       core.WEEKDAY_SUNDAY,
		TransactionEditScope: models.TRANSACTION_EDIT_SCOPE_ALL,
		FeatureRestriction:   l.CurrentConfig().DefaultFeatureRestrictions,
	}

	err := l.users.CreateUser(c, user, false)

	if err != nil {
		log.CliErrorf(c, "[user_data.AddNewUser] failed to create user \"%s\", because %s", user.Username, err.Error())
		return nil, err
	}

	log.CliInfof(c, "[user_data.AddNewUser] user \"%s\" has add successfully, uid is %d", user.Username, user.Uid)

	return user, nil
}

// GetUserByUsername returns user by user name
func (l *UserDataCli) GetUserByUsername(c *core.CliContext, username string) (*models.User, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.GetUserByUsername] user name is empty")
		return nil, errs.ErrUsernameIsEmpty
	}

	user, err := l.users.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.GetUserByUsername] failed to get user by user name \"%s\", because %s", username, err.Error())
		return nil, err
	}

	return user, nil
}

// ModifyUserPassword modifies user password
func (l *UserDataCli) ModifyUserPassword(c *core.CliContext, username string, password string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.ModifyUserPassword] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	if password == "" {
		log.CliErrorf(c, "[user_data.ModifyUserPassword] user password is empty")
		return errs.ErrPasswordIsEmpty
	}

	user, err := l.users.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ModifyUserPassword] failed to get user by user name \"%s\", because %s", username, err.Error())
		return err
	}

	if l.users.IsPasswordEqualsUserPassword(password, user) {
		return errs.ErrNothingWillBeUpdated
	}

	userNew := &models.User{
		Uid:      user.Uid,
		Salt:     user.Salt,
		Password: password,
	}

	err = l.users.UpdateUserPassword(c, userNew)

	if err != nil {
		log.CliErrorf(c, "[user_data.ModifyUserPassword] failed to update user \"%s\" password, because %s", user.Username, err.Error())
		return err
	}

	now := time.Now().Unix()
	err = l.tokens.DeleteTokensBeforeTime(c, user.Uid, now)

	if err == nil {
		log.CliInfof(c, "[user_data.ModifyUserPassword] revoke old tokens before unix time \"%d\" for user \"%s\"", now, user.Username)
	} else {
		log.CliWarnf(c, "[user_data.ModifyUserPassword] failed to revoke old tokens for user \"%s\", because %s", user.Username, err.Error())
	}

	return nil
}

// SendPasswordResetMail sends an email with password reset link
func (l *UserDataCli) SendPasswordResetMail(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.SendPasswordResetMail] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	user, err := l.users.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.SendPasswordResetMail] failed to get user by user name \"%s\", because %s", username, err.Error())
		return err
	}

	if l.CurrentConfig().ForgetPasswordRequireVerifyEmail && !user.EmailVerified {
		log.CliWarnf(c, "[user_data.SendPasswordResetMail] user \"uid:%d\" has not verified email", user.Uid)
		return errs.ErrEmailIsNotVerified
	}

	token, _, err := l.tokens.CreatePasswordResetTokenWithoutUserAgent(c, user)

	if err != nil {
		log.CliErrorf(c, "[user_data.SendPasswordResetMail] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return err
	}

	err = l.forgetPasswords.SendPasswordResetEmail(c, user, token, "")

	if err != nil {
		log.CliWarnf(c, "[user_data.SendPasswordResetMail] cannot send email to \"%s\", because %s", user.Email, err.Error())
		return err
	}

	return nil
}
