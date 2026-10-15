package cmd

import (
	"fmt"
	clis "github.com/mayswind/ezbookkeeping/pkg/cli"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
)

func listUserTokens(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	tokens, err := clis.UserData.ListUserTokens(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.listUserTokens] error occurs when getting user tokens")
		return err
	}

	for i := 0; i < len(tokens); i++ {
		printTokenInfo(tokens[i])

		if i < len(tokens)-1 {
			fmt.Printf("---\n")
		}
	}

	return nil
}

func createNewUserToken(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	tokenType := c.String("type")
	expiresInSeconds := c.Int64("expiresInSeconds")

	if tokenType == "" {
		tokenType = "api"
	}

	if tokenType != "api" && tokenType != "mcp" {
		log.CliErrorf(c, "[user_data.createNewUserToken] token type is invalid")
		return nil
	}

	if expiresInSeconds < 0 || expiresInSeconds > 4294967295 {
		log.CliErrorf(c, "[user_data.createNewUserToken] expiresInSeconds is out of range (0 - 4294967295)")
		return nil
	}

	token, tokenString, err := clis.UserData.CreateNewUserToken(c, username, tokenType, expiresInSeconds)

	if err != nil {
		log.CliErrorf(c, "[user_data.createNewUserToken] error occurs when creating user token")
		return err
	}

	printTokenInfo(token)
	fmt.Printf("[NewToken] %s\n", tokenString)

	return nil
}

func revokeUserToken(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	token := c.String("token")
	err = clis.UserData.RevokeUserToken(c, token)

	if err != nil {
		log.CliErrorf(c, "[user_data.revokeUserToken] error occurs when revoking user token")
		return err
	}

	log.CliInfof(c, "[user_data.revokeUserToken] the specified user token has been revoked successfully")

	return nil
}

func clearUserTokens(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	username := c.String("username")
	err = clis.UserData.ClearUserTokens(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.clearUserTokens] error occurs when clearing user tokens")
		return err
	}

	log.CliInfof(c, "[user_data.clearUserTokens] all tokens of user \"%s\" has been cleared", username)

	return nil
}
