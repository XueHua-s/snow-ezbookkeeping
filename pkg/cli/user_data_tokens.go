package cli

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
)

// ListUserTokens returns all tokens of the specified user
func (l *UserDataCli) ListUserTokens(c *core.CliContext, username string) ([]*models.TokenRecord, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.ListUserTokens] user name is empty")
		return nil, errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ListUserTokens] error occurs when getting user id by user name")
		return nil, err
	}

	tokens, err := l.tokens.GetAllUnexpiredNormalAndMCPTokensByUid(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.ListUserTokens] failed to get tokens of user \"%s\", because %s", username, err.Error())
		return nil, err
	}

	return tokens, nil
}

// CreateNewUserToken returns a new token for the specified user
func (l *UserDataCli) CreateNewUserToken(c *core.CliContext, username string, tokenType string, expiresInSeconds int64) (*models.TokenRecord, string, error) {
	if username == "" {
		log.CliErrorf(c, "[user_data.CreateNewUserToken] user name is empty")
		return nil, "", errs.ErrUsernameIsEmpty
	}

	user, err := l.GetUserByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.CreateNewUserToken] error occurs when getting user by user name")
		return nil, "", err
	}

	var token string
	var tokenRecord *models.TokenRecord

	if tokenType == "api" {
		if !l.CurrentConfig().EnableAPIToken {
			return nil, "", errs.ErrAPITokenNotEnabled
		}

		if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_GENERATE_API_TOKEN) {
			return nil, "", errs.ErrNotPermittedToPerformThisAction
		}

		token, tokenRecord, err = l.tokens.CreateAPITokenViaCli(c, user, expiresInSeconds)
	} else if tokenType == "mcp" {
		if !l.CurrentConfig().EnableMCPServer {
			return nil, "", errs.ErrMCPServerNotEnabled
		}

		if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_MCP_ACCESS) {
			return nil, "", errs.ErrNotPermittedToPerformThisAction
		}

		token, tokenRecord, err = l.tokens.CreateMCPTokenViaCli(c, user, expiresInSeconds)
	} else {
		return nil, "", errs.ErrParameterInvalid
	}

	if err != nil {
		log.CliErrorf(c, "[user_data.CreateNewUserToken] failed to create token for user \"%s\", because %s", username, err.Error())
		return nil, "", err
	}

	return tokenRecord, token, nil
}

// RevokeUserToken revokes the specified token of the user
func (l *UserDataCli) RevokeUserToken(c *core.CliContext, token string) error {
	_, claims, _, err := l.tokens.ParseToken(c, token)

	if err != nil {
		log.CliErrorf(c, "[user_data.RevokeUserToken] failed to parse token, because %s", err.Error())
		return err
	}

	userTokenId, err := utils.StringToInt64(claims.UserTokenId)

	if err != nil {
		log.CliErrorf(c, "[user_data.RevokeUserToken] failed to get user token id, because %s", err.Error())
		return err
	}

	tokenRecord := &models.TokenRecord{
		Uid:             claims.Uid,
		UserTokenId:     userTokenId,
		CreatedUnixTime: claims.IssuedAt,
	}

	tokenId := l.tokens.GenerateTokenId(tokenRecord)
	err = l.tokens.DeleteToken(c, tokenRecord)

	if err != nil {
		log.Errorf(c, "[user_data.RevokeUserToken] failed to revoke token \"id:%s\" for user \"uid:%d\", because %s", tokenId, claims.Uid, err.Error())
		return err
	}

	return nil
}

// ClearUserTokens clears all tokens of the specified user
func (l *UserDataCli) ClearUserTokens(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.ClearUserTokens] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.ClearUserTokens] error occurs when getting user id by user name")
		return err
	}

	now := time.Now().Unix()
	err = l.tokens.DeleteTokensBeforeTime(c, uid, now)

	if err != nil {
		log.CliErrorf(c, "[user_data.ClearUserTokens] failed to delete tokens of user \"%s\", because %s", username, err.Error())
		return err
	}

	return nil
}

// DisableUserTwoFactorAuthorization disables 2fa for the specified user
func (l *UserDataCli) DisableUserTwoFactorAuthorization(c *core.CliContext, username string) error {
	if username == "" {
		log.CliErrorf(c, "[user_data.DisableUserTwoFactorAuthorization] user name is empty")
		return errs.ErrUsernameIsEmpty
	}

	uid, err := l.getUserIdByUsername(c, username)

	if err != nil {
		log.CliErrorf(c, "[user_data.DisableUserTwoFactorAuthorization] error occurs when getting user id by user name")
		return err
	}

	enableTwoFactor, err := l.twoFactorAuthorizations.ExistsTwoFactorSetting(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.DisableUserTwoFactorAuthorization] failed to check two-factor setting, because %s", err.Error())
		return err
	}

	if !enableTwoFactor {
		return errs.ErrTwoFactorIsNotEnabled
	}

	err = l.twoFactorAuthorizations.DeleteTwoFactorRecoveryCodes(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.DisableUserTwoFactorAuthorization] failed to delete two-factor recovery codes for user \"%s\"", username)
		return err
	}

	err = l.twoFactorAuthorizations.DeleteTwoFactorSetting(c, uid)

	if err != nil {
		log.CliErrorf(c, "[user_data.DisableUserTwoFactorAuthorization] failed to delete two-factor setting for user \"%s\"", username)
		return err
	}

	return nil
}
