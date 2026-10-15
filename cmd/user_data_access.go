package cmd

import (
	clis "github.com/mayswind/ezbookkeeping/pkg/cli"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
)

func enableUser(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.EnableUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.enableUser] error occurs when setting user enabled")
		return err
	}

	log.CliInfof(c, "[user_data.enableUser] user \"%s\" has been set enabled", username)

	return nil
}

func disableUser(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.DisableUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.disableUser] error occurs when setting user disabled")
		return err
	}

	log.CliInfof(c, "[user_data.disableUser] user \"%s\" has been set disabled", username)

	return nil
}

func setUserFeatureRestriction(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	featureRestriction := core.ParseUserFeatureRestrictions(c.String("features"))
	err = clis.UserData.SetUserFeatureRestrictions(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.setUserFeatureRestriction] error occurs when setting user feature restriction")
		return err
	}

	log.CliInfof(c, "[user_data.setUserFeatureRestriction] user \"%s\" has been set new feature restriction", username)

	return nil
}

func addUserFeatureRestriction(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	featureRestriction := core.ParseUserFeatureRestrictions(c.String("features"))

	if featureRestriction < 1 {
		log.CliErrorf(c, "[user_data.addUserFeatureRestriction] nothing has been modified")
		return nil
	}

	err = clis.UserData.AddUserFeatureRestrictions(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.addUserFeatureRestriction] error occurs when adding user feature restriction")
		return err
	}

	log.CliInfof(c, "[user_data.addUserFeatureRestriction] user \"%s\" has been add new feature restriction", username)

	return nil
}

func removeUserFeatureRestriction(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	featureRestriction := core.ParseUserFeatureRestrictions(c.String("features"))

	if featureRestriction < 1 {
		log.CliErrorf(c, "[user_data.removeUserFeatureRestriction] nothing has been modified")
		return nil
	}

	err = clis.UserData.RemoveUserFeatureRestrictions(c, username, featureRestriction)

	if err != nil {
		log.CliErrorf(c, "[user_data.removeUserFeatureRestriction] error occurs when removing user feature restriction")
		return err
	}

	log.CliInfof(c, "[user_data.removeUserFeatureRestriction] user \"%s\" has been removed new feature restriction", username)

	return nil
}

func resendUserVerifyEmail(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.ResendVerifyEmail(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.resendUserVerifyEmail] error occurs when resending user verify email")
		return err
	}

	log.CliInfof(c, "[user_data.resendUserVerifyEmail] verify email for user \"%s\" has been resent", username)

	return nil
}

func setUserEmailVerified(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.SetUserEmailVerified(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.setUserEmailVerified] error occurs when setting user email address verified")
		return err
	}

	log.CliInfof(c, "[user_data.setUserEmailVerified] user \"%s\" email address has been set verified", username)

	return nil
}

func setUserEmailUnverified(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.SetUserEmailUnverified(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.setUserEmailUnverified] error occurs when setting user email address unverified")
		return err
	}

	log.CliInfof(c, "[user_data.setUserEmailUnverified] user \"%s\" email address has been set unverified", username)

	return nil
}

func deleteUser(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.DeleteUser(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.deleteUser] error occurs when deleting user")
		return err
	}

	log.CliInfof(c, "[user_data.deleteUser] user \"%s\" has been deleted", username)

	return nil
}

func disableUser2FA(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.DisableUserTwoFactorAuthorization(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.disableUser2FA] error occurs when disabling user two-factor authorization")
		return err
	}

	log.CliInfof(c, "[user_data.disableUser2FA] two-factor authorization of user \"%s\" has been disabled", username)

	return nil
}
