package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
)

// EnableUser sets user enabled according to the specified user name
func (l *UserDataCli) EnableUser(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.EnableUser] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.EnableUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.EnableUser] failed to set user enabled by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// DisableUser sets user disabled according to the specified user name
func (l *UserDataCli) DisableUser(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.DisableUser] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.DisableUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.DisableUser] failed to set user disabled by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// SetUserFeatureRestrictions sets user feature restrictions according to the specified user name
func (l *UserDataCli) SetUserFeatureRestrictions(c *core.CliContext, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.SetUserFeatureRestrictions] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.UpdateUserFeatureRestriction(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.SetUserFeatureRestrictions] failed to set user feature restrictions by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// AddUserFeatureRestrictions adds user feature restrictions according to the specified user name
func (l *UserDataCli) AddUserFeatureRestrictions(c *core.CliContext, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.AddUserFeatureRestrictions] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.AddUserFeatureRestriction(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.AddUserFeatureRestrictions] failed to add user feature restrictions by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// RemoveUserFeatureRestrictions removes user feature restrictions according to the specified user name
func (l *UserDataCli) RemoveUserFeatureRestrictions(c *core.CliContext, username string, featureRestriction core.UserFeatureRestrictions) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.RemoveUserFeatureRestrictions] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.RemoveUserFeatureRestriction(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.RemoveUserFeatureRestrictions] failed to remove user feature restrictions by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// ResendVerifyEmail resends an email with account activation link
func (l *UserDataCli) ResendVerifyEmail(c *core.CliContext, username string) error {
	if !l.CurrentConfig().EnableUserVerifyEmail {
		return errs.ErrEmailValidationNotAllowed
	}

	if username == "" {
		log.CliErrorf(c, "[user_data.ResendVerifyEmail] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	user, err := l.users.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ResendVerifyEmail] failed to get user by user name \"%s\", because %s", username, err.Error())
		return err
	}

	if user.EmailVerified {
		log.CliWarnf(c, "[user_data.ResendVerifyEmail] user \"uid:%d\" email has been verified", user.Uid)
		return errs.ErrEmailIsVerified
	}

	token, _, err := l.tokens.CreateEmailVerifyTokenWithoutUserAgent(c, user)

	if err != nil {
		log.CliErrorf(c, "[user_data.ResendVerifyEmail] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return errs.ErrTokenGenerating
	}

	err = l.users.SendVerifyEmail(user, token, "")

	if err != nil {
		log.CliErrorf(c, "[user_data.ResendVerifyEmail] cannot send email to \"%s\", because %s", user.Email, err.Error())
		return err
	}

	return nil
}

// SetUserEmailVerified sets user email address verified
func (l *UserDataCli) SetUserEmailVerified(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.SetUserEmailVerified] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.SetUserEmailVerified(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.SetUserEmailVerified] failed to set user email address verified by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// SetUserEmailUnverified sets user email address unverified
func (l *UserDataCli) SetUserEmailUnverified(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.SetUserEmailUnverified] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.SetUserEmailUnverified(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.SetUserEmailUnverified] failed to set user email address unverified by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// DeleteUser deletes user according to the specified user name
func (l *UserDataCli) DeleteUser(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.DeleteUser] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	err := l.users.DeleteUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.DeleteUser] failed to delete user by user name \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}
