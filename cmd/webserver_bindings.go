package cmd

import (
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/middlewares"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"net/http"
	"time"
)

func bindMiddleware(fn core.MiddlewareHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(core.WrapWebContext(c, config.TrustedProxyIPs))
	}
}

func bindLocalFile(filePath string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ginCtx.File(filePath)
	}
}

func bindRedirect(fn core.RedirectHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		url, err := fn(c)

		if err != nil {
			utils.PrintJsonErrorResult(c, err)
		} else {
			c.Redirect(http.StatusFound, url)
		}
	}
}

func bindApi(fn core.ApiHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, err := fn(c)

		if err != nil {
			utils.PrintJsonErrorResult(c, err)
		} else {
			utils.PrintJsonSuccessResult(c, result)
		}
	}
}

func bindApiWithTokenUpdate(fn core.ApiHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, err := fn(c)

		if err == nil && config.MapProvider == settings.AmapProvider && config.AmapSecurityVerificationMethod == settings.AmapSecurityVerificationInternalProxyMethod {
			middlewares.AmapApiProxyAuthCookie(c, config)
		}

		if err != nil {
			utils.PrintJsonErrorResult(c, err)
		} else {
			utils.PrintJsonSuccessResult(c, result)
		}
	}
}

func bindJSONRPCApi(fns map[string]core.JSONRPCApiHandlerFunc, skipMethods map[string]int, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)

		var jsonRPCRequest core.JSONRPCRequest
		reqErr := c.ShouldBindBodyWithJSON(&jsonRPCRequest)

		if reqErr != nil {
			utils.PrintJSONRPCErrorResult(c, nil, errs.NewIncompleteOrIncorrectSubmissionError(reqErr))
			return
		}

		if skipMethods != nil {
			httpStatusCode, exists := skipMethods[jsonRPCRequest.Method]

			if exists {
				c.AbortWithStatus(httpStatusCode)
				return
			}
		}

		fn, exists := fns[jsonRPCRequest.Method]

		if !exists {
			utils.PrintJSONRPCErrorResult(c, &jsonRPCRequest, errs.ErrApiNotFound)
			return
		}

		result, err := fn(c, &jsonRPCRequest)

		if err != nil {
			utils.PrintJSONRPCErrorResult(c, &jsonRPCRequest, err)
		} else {
			utils.PrintJSONRPCSuccessResult(c, &jsonRPCRequest, result)
		}
	}
}

func bindEventStreamApi(fn core.EventStreamApiHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		utils.SetEventStreamHeader(c)
		err := fn(c)

		if err != nil && c.Request.Context().Err() == nil {
			utils.WriteEventStreamJsonErrorResult(c, err)
		}
	}
}

func bindCachedJs(fn core.DataHandlerFunc, config *settings.Config, store persistence.CacheStore) gin.HandlerFunc {
	return cache.CachePage(store, time.Minute, func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, _, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/javascript", err)
		} else {
			utils.PrintDataSuccessResult(c, "text/javascript; charset=utf-8", "", result)
		}
	})
}

func bindCsv(fn core.DataHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, fileName, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/text", err)
		} else {
			utils.PrintDataSuccessResult(c, "text/csv; charset=utf-8", fileName, result)
		}
	}
}

func bindTsv(fn core.DataHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, fileName, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/text", err)
		} else {
			utils.PrintDataSuccessResult(c, "text/tab-separated-values; charset=utf-8", fileName, result)
		}
	}
}

func bindImage(fn core.ImageHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, contentType, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/text", err)
		} else {
			utils.PrintDataSuccessResult(c, contentType, "", result)
		}
	}
}

func bindCachedImage(fn core.ImageHandlerFunc, config *settings.Config, store persistence.CacheStore) gin.HandlerFunc {
	return cache.CachePage(store, time.Minute, func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		result, contentType, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/text", err)
		} else {
			utils.PrintDataSuccessResult(c, contentType, "", result)
		}
	})
}

func bindProxy(fn core.ProxyHandlerFunc, config *settings.Config) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		c := core.WrapWebContext(ginCtx, config.TrustedProxyIPs)
		proxy, err := fn(c)

		if err != nil {
			utils.PrintDataErrorResult(c, "text/text", err)
		} else {
			proxy.ServeHTTP(c.Writer, c.Request)
		}
	}
}
