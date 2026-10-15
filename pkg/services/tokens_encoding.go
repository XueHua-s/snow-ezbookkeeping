package services

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"math"
	"strings"
	"time"
	"xorm.io/xorm"
)

// ParseFromTokenId returns token model according to token id
func (s *TokenService) ParseFromTokenId(tokenId string) (*models.TokenRecord, error) {
	pairs := strings.Split(tokenId, ":")

	if len(pairs) != 3 {
		return nil, errs.ErrInvalidTokenId
	}

	uid, err := utils.StringToInt64(pairs[0])

	if err != nil {
		return nil, errs.ErrInvalidTokenId
	}

	createdUnixTime, err := utils.StringToInt64(pairs[1])

	if err != nil {
		return nil, errs.ErrInvalidTokenId
	}

	userTokenId, err := utils.StringToInt64(pairs[2])

	if err != nil {
		return nil, errs.ErrInvalidTokenId
	}

	tokenRecord := &models.TokenRecord{
		Uid:             uid,
		UserTokenId:     userTokenId,
		CreatedUnixTime: createdUnixTime,
	}

	return tokenRecord, nil
}

// GenerateTokenId generates token id according to token model
func (s *TokenService) GenerateTokenId(tokenRecord *models.TokenRecord) string {
	return fmt.Sprintf("%d:%d:%d", tokenRecord.Uid, tokenRecord.CreatedUnixTime, tokenRecord.UserTokenId)
}

func (s *TokenService) parseToken(c core.Context, tokenString string) (*jwt.Token, *core.UserTokenClaims, string, error) {
	claims := &core.UserTokenClaims{}
	tokenContext := ""

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (any, error) {
			now := time.Now().Unix()
			userTokenId, err := utils.StringToInt64(claims.UserTokenId)

			if err != nil {
				log.Warnf(c, "[tokens.parseToken] token \"utid:%s\" in token of user \"uid:%d\" is invalid, because %s", claims.UserTokenId, claims.Uid, err.Error())
				return nil, errs.ErrInvalidUserTokenId
			}

			tokenRecord, err := s.getTokenRecord(c, claims.Uid, userTokenId, claims.IssuedAt)

			if err != nil {
				log.Warnf(c, "[tokens.parseToken] token \"utid:%s\" of user \"uid:%d\" record not found, because %s", claims.UserTokenId, claims.Uid, err.Error())
				return nil, errs.ErrTokenRecordNotFound
			}

			if tokenRecord.ExpiredUnixTime < now {
				log.Warnf(c, "[tokens.parseToken] token \"utid:%s\" of user \"uid:%d\" record is expired", claims.UserTokenId, claims.Uid)
				return nil, errs.ErrTokenExpired
			}

			tokenContext = tokenRecord.Context
			return []byte(tokenRecord.Secret), nil
		},
		jwt.WithIssuedAt(),
	)

	if err != nil {
		if errors.Is(err, request.ErrNoTokenInRequest) {
			return nil, nil, "", errs.ErrTokenIsEmpty
		}

		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenUnverifiable) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			log.Warnf(c, "[tokens.parseToken] token is invalid, because %s", err.Error())
			return nil, nil, "", errs.ErrCurrentInvalidToken
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, "", errs.ErrCurrentTokenExpired
		}

		if errors.Is(err, jwt.ErrTokenUsedBeforeIssued) {
			log.Warnf(c, "[tokens.parseToken] token is invalid, because issue time is later than now")
			return nil, nil, "", errs.ErrCurrentInvalidToken
		}

		return nil, nil, "", err
	}

	return token, claims, tokenContext, err
}

func (s *TokenService) createToken(c core.Context, user *models.User, tokenType core.TokenType, userAgent string, context string, expiryDate time.Duration) (string, *core.UserTokenClaims, *models.TokenRecord, error) {
	var err error
	now := time.Now()

	tokenRecord := &models.TokenRecord{
		Uid:              user.Uid,
		UserTokenId:      s.getUserTokenId(),
		TokenType:        tokenType,
		UserAgent:        userAgent,
		Context:          context,
		CreatedUnixTime:  now.Unix(),
		ExpiredUnixTime:  now.Add(expiryDate).Unix(),
		LastSeenUnixTime: now.Unix(),
	}

	if tokenRecord.Secret, err = utils.GetRandomString(10); err != nil {
		return "", nil, nil, err
	}

	claims := &core.UserTokenClaims{
		UserTokenId: utils.Int64ToString(tokenRecord.UserTokenId),
		Uid:         tokenRecord.Uid,
		Username:    user.Username,
		Type:        tokenRecord.TokenType,
		IssuedAt:    tokenRecord.CreatedUnixTime,
		ExpiresAt:   tokenRecord.ExpiredUnixTime,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(tokenRecord.Secret))

	if err != nil {
		return "", nil, nil, err
	}

	err = s.createTokenRecord(c, tokenRecord)

	if err != nil {
		return "", nil, nil, err
	}

	return tokenString, claims, tokenRecord, err
}

func (s *TokenService) getTokenRecord(c core.Context, uid int64, userTokenId int64, createUnixTime int64) (*models.TokenRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if userTokenId <= 0 {
		return nil, errs.ErrInvalidUserTokenId
	}

	tokenRecord := &models.TokenRecord{}
	has, err := s.TokenDB(uid).NewSession(c).Where("uid=? AND user_token_id=? AND created_unix_time=?", uid, userTokenId, createUnixTime).Limit(1).Get(tokenRecord)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, errs.ErrTokenRecordNotFound
	}

	return tokenRecord, nil
}

func (s *TokenService) createTokenRecord(c core.Context, tokenRecord *models.TokenRecord) error {
	if tokenRecord.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if tokenRecord.UserTokenId <= 0 {
		return errs.ErrInvalidUserTokenId
	}

	return s.TokenDB(tokenRecord.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(tokenRecord)
		return err
	})
}

func (s *TokenService) getUserTokenId() int64 {
	nanoSeconds := time.Now().Nanosecond()
	randomNumber, _ := utils.GetRandomInteger(math.MaxInt32)
	userTokenId := (int64(nanoSeconds) << 32) | int64(randomNumber)

	return userTokenId
}

func (s *TokenService) getUserAgent(ctx *core.WebContext) string {
	userAgent := ""

	if ctx != nil && ctx.Request != nil {
		userAgent = ctx.Request.UserAgent()
	}

	if len(userAgent) > models.TokenMaxUserAgentLength {
		userAgent = utils.SubString(userAgent, 0, models.TokenMaxUserAgentLength)
	}

	return userAgent
}
