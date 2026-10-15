package cmd

import (
	clis "github.com/mayswind/ezbookkeeping/pkg/cli"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
)

func addNewUser(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	email := c.String("email")
	nickname := c.String("nickname")
	password := c.String("password")
	defaultCurrency := c.String("default-currency")

	user, err := clis.UserData.AddNewUser(c, username, email, nickname, password, defaultCurrency)

	if err != nil {
		log.CliErrorf(c, "[user_data.addNewUser] error occurs when adding new user")
		return err
	}

	printUserInfo(user)

	return nil
}

func getUserInfo(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	user, err := clis.UserData.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.getUserInfo] error occurs when getting user data")
		return err
	}

	printUserInfo(user)

	return nil
}

func modifyUserPassword(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	password := c.String("password")
	err = clis.UserData.ModifyUserPassword(c, username, password)

	if err != nil {
		log.CliErrorf(c, "[user_data.modifyUserPassword] error occurs when modifying user password")
		return err
	}

	log.CliInfof(c, "[user_data.modifyUserPassword] password of user \"%s\" has been changed", username)

	return nil
}

func sendPasswordResetMail(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.SendPasswordResetMail(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.sendPasswordResetMail] error occurs when sending password reset email")
		return err
	}

	log.CliInfof(c, "[user_data.sendPasswordResetMail] a password reset email for user \"%s\" has been sent", username)

	return nil
}
