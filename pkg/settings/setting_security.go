package settings

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"gopkg.in/ini.v1"
	"time"
)

func loadUuidConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	if getConfigItemStringValue(configFile, sectionName, "generator_type") == InternalUuidGeneratorType {
		config.UuidGeneratorType = InternalUuidGeneratorType
	} else {
		return errs.ErrInvalidUuidMode
	}

	config.UuidServerId = getConfigItemUint8Value(configFile, sectionName, "server_id", 0)

	return nil
}

func loadDuplicateCheckerConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	if getConfigItemStringValue(configFile, sectionName, "checker_type") == InMemoryDuplicateCheckerType {
		config.DuplicateCheckerType = InMemoryDuplicateCheckerType
	} else {
		return errs.ErrInvalidDuplicateCheckerType
	}

	config.InMemoryDuplicateCheckerCleanupInterval = getConfigItemUint32Value(configFile, sectionName, "cleanup_interval", defaultInMemoryDuplicateCheckerCleanupInterval)

	if config.InMemoryDuplicateCheckerCleanupInterval < 1 {
		return errs.ErrInvalidInMemoryDuplicateCheckerCleanupInterval
	}

	config.InMemoryDuplicateCheckerCleanupIntervalDuration = time.Duration(config.InMemoryDuplicateCheckerCleanupInterval) * time.Second

	duplicateSubmissionsInterval := getConfigItemUint32Value(configFile, sectionName, "duplicate_submissions_interval", defaultDuplicateSubmissionsInterval)

	config.EnableDuplicateSubmissionsCheck = duplicateSubmissionsInterval > 0

	if duplicateSubmissionsInterval < 1 {
		duplicateSubmissionsInterval = defaultDuplicateSubmissionsInterval
	}

	config.DuplicateSubmissionsInterval = duplicateSubmissionsInterval
	config.DuplicateSubmissionsIntervalDuration = time.Duration(config.DuplicateSubmissionsInterval) * time.Second

	return nil
}

func loadCronConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableRemoveExpiredTokens = getConfigItemBoolValue(configFile, sectionName, "enable_remove_expired_tokens", false)
	config.EnableCreateScheduledTransaction = getConfigItemBoolValue(configFile, sectionName, "enable_create_scheduled_transaction", false)

	return nil
}

func loadSecurityConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	var err error

	config.SecretKeyNoSet = !getConfigItemIsSet(configFile, sectionName, "secret_key")
	config.SecretKey = getConfigItemStringValue(configFile, sectionName, "secret_key", defaultSecretKey)
	config.TrustedProxyIPs, config.TrustedProxyTextualIPs, err = getCIDRList(configFile, sectionName, "trusted_proxy_ips", defaultTrustedProxyIPs)

	if err != nil {
		return err
	}

	config.TokenExpiredTime = getConfigItemUint32Value(configFile, sectionName, "token_expired_time", defaultTokenExpiredTime)

	if config.TokenExpiredTime < 60 {
		return errs.ErrInvalidTokenExpiredTime
	}

	config.TokenExpiredTimeDuration = time.Duration(config.TokenExpiredTime) * time.Second

	config.TokenMinRefreshInterval = getConfigItemUint32Value(configFile, sectionName, "token_min_refresh_interval", defaultTokenMinRefreshInterval)

	if config.TokenMinRefreshInterval >= config.TokenExpiredTime {
		return errs.ErrInvalidTokenMinRefreshInterval
	}

	config.TemporaryTokenExpiredTime = getConfigItemUint32Value(configFile, sectionName, "temporary_token_expired_time", defaultTemporaryTokenExpiredTime)

	if config.TemporaryTokenExpiredTime < 60 {
		return errs.ErrInvalidTemporaryTokenExpiredTime
	}

	config.TemporaryTokenExpiredTimeDuration = time.Duration(config.TemporaryTokenExpiredTime) * time.Second

	config.EmailVerifyTokenExpiredTime = getConfigItemUint32Value(configFile, sectionName, "email_verify_token_expired_time", defaultEmailVerifyTokenExpiredTime)

	if config.EmailVerifyTokenExpiredTime < 60 {
		return errs.ErrInvalidEmailVerifyTokenExpiredTime
	}

	config.EmailVerifyTokenExpiredTimeDuration = time.Duration(config.EmailVerifyTokenExpiredTime) * time.Second

	config.PasswordResetTokenExpiredTime = getConfigItemUint32Value(configFile, sectionName, "password_reset_token_expired_time", defaultPasswordResetTokenExpiredTime)

	if config.PasswordResetTokenExpiredTime < 60 {
		return errs.ErrInvalidPasswordResetTokenExpiredTime
	}

	config.PasswordResetTokenExpiredTimeDuration = time.Duration(config.PasswordResetTokenExpiredTime) * time.Second

	config.EnableAPIToken = getConfigItemBoolValue(configFile, sectionName, "enable_api_token", false)
	config.APITokenAllowedRemoteIPs, err = getIPPatterns(configFile, sectionName, "api_token_allowed_remote_ips", "")

	if err != nil {
		return err
	}

	config.MaxFailuresPerIpPerMinute = getConfigItemUint32Value(configFile, sectionName, "max_failures_per_ip_per_minute", defaultMaxFailuresPerIpPerMinute)
	config.MaxFailuresPerUserPerMinute = getConfigItemUint32Value(configFile, sectionName, "max_failures_per_user_per_minute", defaultMaxFailuresPerUserPerMinute)

	return nil
}

func loadAuthConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableInternalAuth = getConfigItemBoolValue(configFile, sectionName, "enable_internal_auth", true)
	config.EnableOAuth2Login = getConfigItemBoolValue(configFile, sectionName, "enable_oauth2_auth", false)
	config.EnableTwoFactor = getConfigItemBoolValue(configFile, sectionName, "enable_two_factor", true)
	config.EnableUserForgetPassword = getConfigItemBoolValue(configFile, sectionName, "enable_forget_password", false)
	config.ForgetPasswordRequireVerifyEmail = getConfigItemBoolValue(configFile, sectionName, "forget_password_require_email_verify", false)
	config.OAuth2ClientID = getConfigItemStringValue(configFile, sectionName, "oauth2_client_id")
	config.OAuth2ClientSecret = getConfigItemStringValue(configFile, sectionName, "oauth2_client_secret")
	config.OAuth2UsePKCE = getConfigItemBoolValue(configFile, sectionName, "oauth2_use_pkce", false)

	oauth2UserIdentifier := getConfigItemStringValue(configFile, sectionName, "oauth2_user_identifier")

	if oauth2UserIdentifier == OAuth2UserIdentifierEmail {
		config.OAuth2UserIdentifier = OAuth2UserIdentifierEmail
	} else if oauth2UserIdentifier == OAuth2UserIdentifierUsername {
		config.OAuth2UserIdentifier = OAuth2UserIdentifierUsername
	} else {
		return errs.ErrInvalidOAuth2UserIdentifier
	}

	config.OAuth2AutoRegister = getConfigItemBoolValue(configFile, sectionName, "oauth2_auto_register", true)

	oauth2Provider := getConfigItemStringValue(configFile, sectionName, "oauth2_provider")

	if oauth2Provider == "" {
		config.OAuth2Provider = ""
	} else if oauth2Provider == OAuth2ProviderOIDC {
		config.OAuth2Provider = OAuth2ProviderOIDC
	} else if oauth2Provider == OAuth2ProviderNextcloud {
		config.OAuth2Provider = OAuth2ProviderNextcloud
	} else if oauth2Provider == OAuth2ProviderGitea {
		config.OAuth2Provider = OAuth2ProviderGitea
	} else if oauth2Provider == OAuth2ProviderGithub {
		config.OAuth2Provider = OAuth2ProviderGithub
	} else {
		return errs.ErrInvalidOAuth2Provider
	}

	config.OAuth2StateExpiredTime = getConfigItemUint32Value(configFile, sectionName, "oauth2_state_expired_time", defaultOAuth2StateExpiredTime)

	if config.OAuth2StateExpiredTime < 60 {
		return errs.ErrInvalidOAuth2StateExpiredTime
	}

	config.OAuth2StateExpiredTimeDuration = time.Duration(config.OAuth2StateExpiredTime) * time.Second

	config.OAuth2Proxy = getConfigItemStringValue(configFile, sectionName, "oauth2_proxy", "system")
	config.OAuth2RequestTimeout = getConfigItemUint32Value(configFile, sectionName, "oauth2_request_timeout", defaultOAuth2RequestTimeout)
	config.OAuth2SkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "oauth2_skip_tls_verify", false)

	config.OAuth2OIDCProviderIssuerURL = getConfigItemStringValue(configFile, sectionName, "oidc_provider_base_url")
	config.OAuth2OIDCProviderCheckIssuerURL = getConfigItemBoolValue(configFile, sectionName, "oidc_provider_check_issuer_url", true)
	config.OAuth2OIDCCustomDisplayNameConfig = getMultiLanguageContentConfig(configFile, sectionName, "enable_oidc_display_name", "oidc_custom_display_name")
	config.OAuth2NextcloudBaseUrl = getConfigItemStringValue(configFile, sectionName, "nextcloud_base_url")
	config.OAuth2GiteaBaseUrl = getConfigItemStringValue(configFile, sectionName, "gitea_base_url")

	return nil
}
