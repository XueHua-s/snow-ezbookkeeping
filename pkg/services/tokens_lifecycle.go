package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"time"
	"xorm.io/xorm"
)

// UpdateTokenLastSeen updates the last seen time of specified token
func (s *TokenService) UpdateTokenLastSeen(c core.Context, tokenRecord *models.TokenRecord) error {
	if tokenRecord.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if tokenRecord.UserTokenId <= 0 {
		return errs.ErrInvalidUserTokenId
	}

	tokenRecord.LastSeenUnixTime = time.Now().Unix()

	return s.TokenDB(tokenRecord.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("last_seen_unix_time").Where("uid=? AND user_token_id=? AND created_unix_time=?", tokenRecord.Uid, tokenRecord.UserTokenId, tokenRecord.CreatedUnixTime).Update(tokenRecord)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTokenRecordNotFound
		}

		return nil
	})
}

// DeleteToken deletes given token from database
func (s *TokenService) DeleteToken(c core.Context, tokenRecord *models.TokenRecord) error {
	if tokenRecord.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if tokenRecord.UserTokenId <= 0 {
		return errs.ErrInvalidUserTokenId
	}

	return s.TokenDB(tokenRecord.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.Where("uid=? AND user_token_id=? AND created_unix_time=?", tokenRecord.Uid, tokenRecord.UserTokenId, tokenRecord.CreatedUnixTime).Delete(&models.TokenRecord{})

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTokenRecordNotFound
		}

		return nil
	})
}

// DeleteTokens deletes given tokens from database
func (s *TokenService) DeleteTokens(c core.Context, uid int64, tokenRecords []*models.TokenRecord) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.TokenDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(tokenRecords); i++ {
			tokenRecord := tokenRecords[i]
			deletedRows, err := sess.Where("uid=? AND user_token_id=? AND created_unix_time=?", uid, tokenRecord.UserTokenId, tokenRecord.CreatedUnixTime).Delete(&models.TokenRecord{})

			if err != nil {
				return err
			} else if deletedRows < 1 {
				return errs.ErrTokenRecordNotFound
			}
		}

		return nil
	})
}

// DeleteTokenByClaims deletes given token from database
func (s *TokenService) DeleteTokenByClaims(c core.Context, claims *core.UserTokenClaims) error {
	userTokenId, err := utils.StringToInt64(claims.UserTokenId)

	if err != nil {
		return errs.ErrInvalidUserTokenId
	}

	return s.DeleteToken(c, &models.TokenRecord{
		Uid:             claims.Uid,
		UserTokenId:     userTokenId,
		CreatedUnixTime: claims.IssuedAt,
	})
}

// DeleteTokensBeforeTime deletes tokens that is created before specific time
func (s *TokenService) DeleteTokensBeforeTime(c core.Context, uid int64, createTime int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.TokenDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=? AND created_unix_time<?", uid, createTime).Delete(&models.TokenRecord{})
		return err
	})
}

// DeleteTokensByType deletes specified type tokens
func (s *TokenService) DeleteTokensByType(c core.Context, uid int64, tokenType core.TokenType) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.TokenDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=? AND token_type=?", uid, tokenType).Delete(&models.TokenRecord{})
		return err
	})
}

// DeleteAllExpiredTokens deletes all expired tokens
func (s *TokenService) DeleteAllExpiredTokens(c core.Context) error {
	var errors []error
	totalCount := int64(0)

	for i := 0; i < s.TokenDBCount(); i++ {
		err := s.TokenDBByIndex(i).DoTransaction(c, func(sess *xorm.Session) error {
			count, err := sess.Where("expired_unix_time<=?", time.Now().Unix()).Delete(&models.TokenRecord{})
			totalCount += count
			return err
		})

		if err != nil {
			errors = append(errors, err)
		}
	}

	if totalCount > 0 {
		log.Infof(c, "[tokens.DeleteAllExpiredTokens] %d expired tokens have been deleted", totalCount)
	} else if len(errors) == 0 {
		log.Infof(c, "[tokens.DeleteAllExpiredTokens] no expired tokens have been deleted")
	}

	return errs.NewMultiErrorOrNil(errors...)
}

// ExistsValidTokenByType returns whether the given token type exists
func (s *TokenService) ExistsValidTokenByType(c core.Context, uid int64, tokenType core.TokenType) (bool, error) {
	if uid <= 0 {
		return false, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	return s.TokenDB(uid).NewSession(c).Cols("uid", "user_token_id", "expired_unix_time").Where("uid=? AND token_type=? AND expired_unix_time>?", uid, tokenType, now).Exist(&models.TokenRecord{})
}
