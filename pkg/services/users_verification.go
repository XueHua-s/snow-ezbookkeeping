package services

import (
	"bytes"
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/locales"
	"github.com/mayswind/ezbookkeeping/pkg/mail"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/templates"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"net/url"
)

// ExistsUsername returns whether the given user name exists
func (s *UserService) ExistsUsername(c core.Context, username string) (bool, error) {
	if username == "" {
		return false, errs.ErrUsernameIsEmpty
	}

	return s.UserDB().NewSession(c).Cols("username").Where("username=? AND deleted=?", username, false).Exist(&models.User{})
}

// ExistsEmail returns whether the given user email exists
func (s *UserService) ExistsEmail(c core.Context, email string) (bool, error) {
	if email == "" {
		return false, errs.ErrEmailIsEmpty
	}

	return s.UserDB().NewSession(c).Cols("email").Where("email=? AND deleted=?", email, false).Exist(&models.User{})
}

// SendVerifyEmail sends verify email according to specified parameters
func (s *UserService) SendVerifyEmail(user *models.User, verifyEmailToken string, backupLocale string) error {
	if !s.CurrentConfig().EnableSMTP {
		return errs.ErrSMTPServerNotEnabled
	}

	locale := user.Language

	if locale == "" {
		locale = backupLocale
	}

	localeTextItems := locales.GetLocaleTextItems(locale)
	verifyEmailTextItems := localeTextItems.VerifyEmailTextItems

	expireTimeInMinutes := s.CurrentConfig().EmailVerifyTokenExpiredTimeDuration.Minutes()
	verifyEmailUrl := fmt.Sprintf(verifyEmailUrlFormat, s.CurrentConfig().RootUrl, url.QueryEscape(verifyEmailToken))

	tmpl, err := templates.GetTemplate(templates.TEMPLATE_VERIFY_EMAIL)

	if err != nil {
		return err
	}

	templateParams := map[string]any{
		"AppName": localeTextItems.GlobalTextItems.AppName,
		"VerifyEmail": map[string]any{
			"Title":               verifyEmailTextItems.Title,
			"Salutation":          fmt.Sprintf(verifyEmailTextItems.SalutationFormat, user.Nickname),
			"DescriptionAboveBtn": verifyEmailTextItems.DescriptionAboveBtn,
			"VerifyEmailUrl":      verifyEmailUrl,
			"VerifyEmail":         verifyEmailTextItems.VerifyEmail,
			"DescriptionBelowBtn": fmt.Sprintf(verifyEmailTextItems.DescriptionBelowBtnFormat, localeTextItems.GlobalTextItems.AppName, expireTimeInMinutes),
		},
	}

	var bodyBuffer bytes.Buffer
	err = tmpl.Execute(&bodyBuffer, templateParams)

	if err != nil {
		return err
	}

	message := &mail.MailMessage{
		To:      user.Email,
		Subject: verifyEmailTextItems.Title,
		Body:    bodyBuffer.String(),
	}

	err = s.SendMail(message)

	return err
}

// IsPasswordEqualsUserPassword returns whether the given password is correct
func (s *UserService) IsPasswordEqualsUserPassword(password string, user *models.User) bool {
	return user.Password == utils.EncodePassword(password, user.Salt)
}
