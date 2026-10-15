package services

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"time"
)

const tokenMaxExpiredAtUnixTime = int64(253402300799)

// TokenService represents user token service
type TokenService struct {
	ServiceUsingDB
	ServiceUsingConfig
}

// Initialize a user token service singleton instance
var (
	Tokens = &TokenService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingConfig: ServiceUsingConfig{
			container: settings.Container,
		},
	}
)

// GetAllTokensByUid returns all token models of given user
func (s *TokenService) GetAllTokensByUid(c core.Context, uid int64) ([]*models.TokenRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var tokenRecords []*models.TokenRecord
	err := s.TokenDB(uid).NewSession(c).Cols("uid", "user_token_id", "token_type", "user_agent", "created_unix_time", "expired_unix_time").Where("uid=?", uid).Find(&tokenRecords)

	return tokenRecords, err
}

// GetAllUnexpiredNormalAndMCPTokensByUid returns all available token models of given user
func (s *TokenService) GetAllUnexpiredNormalAndMCPTokensByUid(c core.Context, uid int64) ([]*models.TokenRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	var tokenRecords []*models.TokenRecord
	err := s.TokenDB(uid).NewSession(c).Cols("uid", "user_token_id", "token_type", "user_agent", "created_unix_time", "expired_unix_time", "last_seen_unix_time").Where("uid=? AND (token_type=? OR token_type=? OR token_type=?) AND expired_unix_time>?", uid, core.USER_TOKEN_TYPE_NORMAL, core.USER_TOKEN_TYPE_MCP, core.USER_TOKEN_TYPE_API, now).Find(&tokenRecords)

	return tokenRecords, err
}

// ParseToken returns the token model according to token content
func (s *TokenService) ParseToken(c core.Context, token string) (*jwt.Token, *core.UserTokenClaims, string, error) {
	return s.parseToken(c, token)
}

// CreateToken generates a new normal token and saves to database
func (s *TokenService) CreateToken(c *core.WebContext, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_NORMAL, s.getUserAgent(c), "", s.CurrentConfig().TokenExpiredTimeDuration)
	return token, claims, err
}

// CreateRequire2FAToken generates a new token requiring user to verify 2fa passcode and saves to database
func (s *TokenService) CreateRequire2FAToken(c *core.WebContext, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_REQUIRE_2FA, s.getUserAgent(c), "", s.CurrentConfig().TemporaryTokenExpiredTimeDuration)
	return token, claims, err
}

// CreateEmailVerifyToken generates a new email verify token and saves to database
func (s *TokenService) CreateEmailVerifyToken(c *core.WebContext, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_EMAIL_VERIFY, s.getUserAgent(c), "", s.CurrentConfig().EmailVerifyTokenExpiredTimeDuration)
	return token, claims, err
}

// CreateEmailVerifyTokenWithoutUserAgent generates a new email verify token and saves to database
func (s *TokenService) CreateEmailVerifyTokenWithoutUserAgent(c core.Context, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_EMAIL_VERIFY, "", "", s.CurrentConfig().EmailVerifyTokenExpiredTimeDuration)
	return token, claims, err
}

// CreatePasswordResetToken generates a new password reset token and saves to database
func (s *TokenService) CreatePasswordResetToken(c *core.WebContext, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_PASSWORD_RESET, s.getUserAgent(c), "", s.CurrentConfig().PasswordResetTokenExpiredTimeDuration)
	return token, claims, err
}

// CreatePasswordResetTokenWithoutUserAgent generates a new password reset token and saves to database
func (s *TokenService) CreatePasswordResetTokenWithoutUserAgent(c core.Context, user *models.User) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_PASSWORD_RESET, "", "", s.CurrentConfig().PasswordResetTokenExpiredTimeDuration)
	return token, claims, err
}

// CreateAPIToken generates a new API token and saves to database
func (s *TokenService) CreateAPIToken(c *core.WebContext, user *models.User, expiresInSeconds int64) (string, *core.UserTokenClaims, error) {
	var tokenExpiredTimeDuration time.Duration

	if expiresInSeconds > 0 {
		tokenExpiredTimeDuration = time.Duration(expiresInSeconds) * time.Second
	} else {
		tokenExpiredTimeDuration = time.Unix(tokenMaxExpiredAtUnixTime, 0).Sub(time.Now())
	}

	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_API, s.getUserAgent(c), "", tokenExpiredTimeDuration)
	return token, claims, err
}

// CreateAPITokenViaCli generates a new API token and saves to database
func (s *TokenService) CreateAPITokenViaCli(c *core.CliContext, user *models.User, expiresInSeconds int64) (string, *models.TokenRecord, error) {
	var tokenExpiredTimeDuration time.Duration

	if expiresInSeconds > 0 {
		tokenExpiredTimeDuration = time.Duration(expiresInSeconds) * time.Second
	} else {
		tokenExpiredTimeDuration = time.Unix(tokenMaxExpiredAtUnixTime, 0).Sub(time.Now())
	}

	token, _, tokenRecord, err := s.createToken(c, user, core.USER_TOKEN_TYPE_API, core.TokenUserAgentCreatedViaCli, "", tokenExpiredTimeDuration)
	return token, tokenRecord, err
}

// CreateMCPToken generates a new MCP token and saves to database
func (s *TokenService) CreateMCPToken(c *core.WebContext, user *models.User, expiresInSeconds int64) (string, *core.UserTokenClaims, error) {
	var tokenExpiredTimeDuration time.Duration

	if expiresInSeconds > 0 {
		tokenExpiredTimeDuration = time.Duration(expiresInSeconds) * time.Second
	} else {
		tokenExpiredTimeDuration = time.Unix(tokenMaxExpiredAtUnixTime, 0).Sub(time.Now())
	}

	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_MCP, s.getUserAgent(c), "", tokenExpiredTimeDuration)
	return token, claims, err
}

// CreateMCPTokenViaCli generates a new MCP token and saves to database
func (s *TokenService) CreateMCPTokenViaCli(c *core.CliContext, user *models.User, expiresInSeconds int64) (string, *models.TokenRecord, error) {
	var tokenExpiredTimeDuration time.Duration

	if expiresInSeconds > 0 {
		tokenExpiredTimeDuration = time.Duration(expiresInSeconds) * time.Second
	} else {
		tokenExpiredTimeDuration = time.Unix(tokenMaxExpiredAtUnixTime, 0).Sub(time.Now())
	}

	token, _, tokenRecord, err := s.createToken(c, user, core.USER_TOKEN_TYPE_MCP, core.TokenUserAgentCreatedViaCli, "", tokenExpiredTimeDuration)
	return token, tokenRecord, err
}

// CreateOAuth2CallbackRequireVerifyToken generates a new OAuth 2.0 callback token requiring user to verify and saves to database
func (s *TokenService) CreateOAuth2CallbackRequireVerifyToken(c *core.WebContext, user *models.User, context string) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_OAUTH2_CALLBACK_REQUIRE_VERIFY, s.getUserAgent(c), context, s.CurrentConfig().TemporaryTokenExpiredTimeDuration)
	return token, claims, err
}

// CreateOAuth2CallbackToken generates a new OAuth 2.0 callback token and saves to database
func (s *TokenService) CreateOAuth2CallbackToken(c *core.WebContext, user *models.User, context string) (string, *core.UserTokenClaims, error) {
	token, claims, _, err := s.createToken(c, user, core.USER_TOKEN_TYPE_OAUTH2_CALLBACK, s.getUserAgent(c), context, s.CurrentConfig().TemporaryTokenExpiredTimeDuration)
	return token, claims, err
}
