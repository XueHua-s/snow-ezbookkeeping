package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
)

func (a *AuthorizationsApi) getAuthResponse(c *core.WebContext, token string, need2FA bool, user *models.User, applicationCloudSettings *models.ApplicationCloudSettingSlice) *models.AuthResponse {
	return &models.AuthResponse{
		Token:                    token,
		Need2FA:                  need2FA,
		User:                     a.GetUserBasicInfo(user),
		ApplicationCloudSettings: applicationCloudSettings,
		NotificationContent:      a.GetAfterLoginNotificationContent(user.Language, c.GetClientLocale()),
	}
}

func (a *AuthorizationsApi) update2FAPasscodeUsed(c *core.WebContext, uid int64, passcode string) bool {
	passcodeHash := utils.MD5EncodeToStringWithUidAndSalt([]byte(passcode), uid, a.CurrentConfig().SecretKey)
	found, remark := a.GetOrSetSubmissionRemarkWithCustomExpiration(duplicatechecker.DUPLICATE_CHECKER_TYPE_2FA_PASSCODE, uid, passcodeHash, utils.Int64ToString(time.Now().Unix()), a.twoFactorAuthorizations.GetTwoFactorPasscodeExpirationDuration())

	if found {
		log.Warnf(c, "[authorizations.update2FAPasscodeUsed] passcode \"%s\" has been used for user \"uid:%d\" in unix timestamp %s", passcode, uid, remark)
	}

	return found
}
