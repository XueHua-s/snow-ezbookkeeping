package cmd

import (
	"fmt"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/mayswind/ezbookkeeping/pkg/api"
	"github.com/mayswind/ezbookkeeping/pkg/auth/oauth2"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/cron"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/mcp"
	"github.com/mayswind/ezbookkeeping/pkg/middlewares"
	"github.com/mayswind/ezbookkeeping/pkg/requestid"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/validators"
	"github.com/urfave/cli/v3"
	"net/http"
	"path/filepath"
	"time"
)

// WebServer represents the server command
var WebServer = &cli.Command{
	Name:  "server",
	Usage: "ezBookkeeping web server operation",
	Commands: []*cli.Command{
		{
			Name:   "run",
			Usage:  "Run ezBookkeeping web server",
			Action: bindAction(startWebServer),
		},
	},
}

func startWebServer(c *core.CliContext) error {
	config, err := initializeSystem(c)

	if err != nil {
		return err
	}

	log.BootInfof(c, "[webserver.startWebServer] static root path is %s", config.StaticRootPath)

	if config.AutoUpdateDatabase {
		err = updateAllDatabaseTablesStructure(c)

		if err != nil {
			log.BootErrorf(c, "[webserver.startWebServer] update database table structure failed, because %s", err.Error())
			return err
		}
	}

	err = requestid.InitializeRequestIdGenerator(c, config)

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] initializes requestid generator failed, because %s", err.Error())
		return err
	}

	err = mcp.InitializeMCPHandlers(config)

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] initializes mcp handlers failed, because %s", err.Error())
		return err
	}

	err = oauth2.InitializeOAuth2Provider(config)

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] initializes oauth 2.0 provider failed, because %s", err.Error())
		return err
	}

	err = cron.InitializeCronJobSchedulerContainer(c, config, true)

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] initializes cron job scheduler failed, because %s", err.Error())
		return err
	}

	serverInfo := fmt.Sprintf("current server id is %d, current instance id is %d", requestid.Container.GetCurrentServerUniqId(), requestid.Container.GetCurrentInstanceUniqId())
	uuidServerInfo := ""
	if config.UuidGeneratorType == settings.InternalUuidGeneratorType {
		uuidServerInfo = fmt.Sprintf(", current uuid server id is %d", config.UuidServerId)
	}

	log.BootInfof(c, "[webserver.startWebServer] %s%s", serverInfo, uuidServerInfo)

	if config.Mode == settings.MODE_PRODUCTION {
		gin.SetMode(gin.ReleaseMode)
	}

	workboxFileNames := utils.ListFileNamesWithPrefixAndSuffix(config.StaticRootPath, "workbox-", ".js")

	router := gin.New()
	router.Use(bindMiddleware(middlewares.Recovery, config))

	err = router.SetTrustedProxies(config.TrustedProxyTextualIPs)

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] set trusted proxy failed, because %s", err.Error())
		return err
	}

	if config.EnableGZip {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("notBlank", validators.NotBlank)
		_ = v.RegisterValidation("validUsername", validators.ValidUsername)
		_ = v.RegisterValidation("validEmail", validators.ValidEmail)
		_ = v.RegisterValidation("validNickname", validators.ValidNickname)
		_ = v.RegisterValidation("validCurrency", validators.ValidCurrency)
		_ = v.RegisterValidation("validHexRGBColor", validators.ValidHexRGBColor)
		_ = v.RegisterValidation("validAmountFilter", validators.ValidAmountFilter)
		_ = v.RegisterValidation("validTransactionAmount", validators.ValidTransactionAmount)
		_ = v.RegisterValidation("validTagFilter", validators.ValidTagFilter)
		_ = v.RegisterValidation("validFiscalYearStart", validators.ValidateFiscalYearStart)
	}

	router.NoRoute(bindApi(api.Default.ApiNotFound, config))
	router.NoMethod(bindApi(api.Default.MethodNotAllowed, config))

	serverSettingsCacheStore := persistence.NewInMemoryStore(time.Minute)

	router.StaticFile("/", filepath.Join(config.StaticRootPath, "index.html"))
	router.Static("/js", filepath.Join(config.StaticRootPath, "js"))
	router.Static("/css", filepath.Join(config.StaticRootPath, "css"))
	router.Static("/img", filepath.Join(config.StaticRootPath, "img"))
	router.Static("/fonts", filepath.Join(config.StaticRootPath, "fonts"))

	router.StaticFile("robots.txt", filepath.Join(config.StaticRootPath, "robots.txt"))
	router.StaticFile("favicon.ico", filepath.Join(config.StaticRootPath, "favicon.ico"))
	router.StaticFile("favicon.png", filepath.Join(config.StaticRootPath, "favicon.png"))
	router.StaticFile("touchicon.png", filepath.Join(config.StaticRootPath, "touchicon.png"))
	router.StaticFile("manifest.json", filepath.Join(config.StaticRootPath, "manifest.json"))
	router.StaticFile("sw.js", filepath.Join(config.StaticRootPath, "sw.js"))
	router.GET("/server_settings.js", bindCachedJs(api.ServerSettings.ServerSettingsJavascriptHandler, config, serverSettingsCacheStore))

	for i := 0; i < len(workboxFileNames); i++ {
		router.StaticFile("/"+workboxFileNames[i], filepath.Join(config.StaticRootPath, workboxFileNames[i]))
	}

	router.StaticFile("/mobile", filepath.Join(config.StaticRootPath, "mobile.html"))
	router.Match([]string{http.MethodHead, http.MethodGet}, "/mobile#/*fragment", bindLocalFile(filepath.Join(config.StaticRootPath, "mobile.html")))  // add compatibility for browsers that send the full URL with the fragment to the server
	router.Match([]string{http.MethodHead, http.MethodGet}, "/mobile#!/*fragment", bindLocalFile(filepath.Join(config.StaticRootPath, "mobile.html"))) // add compatibility for browsers that send the full URL with the fragment to the server
	router.Static("/mobile/js", filepath.Join(config.StaticRootPath, "js"))
	router.Static("/mobile/css", filepath.Join(config.StaticRootPath, "css"))
	router.Static("/mobile/img", filepath.Join(config.StaticRootPath, "img"))
	router.Static("/mobile/fonts", filepath.Join(config.StaticRootPath, "fonts"))
	router.StaticFile("/mobile/favicon.ico", filepath.Join(config.StaticRootPath, "favicon.ico"))
	router.StaticFile("/mobile/favicon.png", filepath.Join(config.StaticRootPath, "favicon.png"))
	router.StaticFile("/mobile/touchicon.png", filepath.Join(config.StaticRootPath, "touchicon.png"))
	router.StaticFile("/mobile/manifest.json", filepath.Join(config.StaticRootPath, "manifest.json"))
	router.StaticFile("/mobile/sw.js", filepath.Join(config.StaticRootPath, "sw.js"))
	router.GET("/mobile/server_settings.js", bindCachedJs(api.ServerSettings.ServerSettingsJavascriptHandler, config, serverSettingsCacheStore))

	for i := 0; i < len(workboxFileNames); i++ {
		router.StaticFile("/mobile/"+workboxFileNames[i], filepath.Join(config.StaticRootPath, workboxFileNames[i]))
	}

	router.StaticFile("/desktop", filepath.Join(config.StaticRootPath, "desktop.html"))
	router.Match([]string{http.MethodHead, http.MethodGet}, "/desktop#/*fragment", bindLocalFile(filepath.Join(config.StaticRootPath, "desktop.html"))) // add compatibility for browsers that send the full URL with the fragment to the server
	router.Static("/desktop/js", filepath.Join(config.StaticRootPath, "js"))
	router.Static("/desktop/css", filepath.Join(config.StaticRootPath, "css"))
	router.Static("/desktop/img", filepath.Join(config.StaticRootPath, "img"))
	router.Static("/desktop/fonts", filepath.Join(config.StaticRootPath, "fonts"))
	router.StaticFile("/desktop/favicon.ico", filepath.Join(config.StaticRootPath, "favicon.ico"))
	router.StaticFile("/desktop/favicon.png", filepath.Join(config.StaticRootPath, "favicon.png"))
	router.StaticFile("/desktop/touchicon.png", filepath.Join(config.StaticRootPath, "touchicon.png"))
	router.StaticFile("/desktop/manifest.json", filepath.Join(config.StaticRootPath, "manifest.json"))
	router.StaticFile("/desktop/sw.js", filepath.Join(config.StaticRootPath, "sw.js"))
	router.GET("/desktop/server_settings.js", bindCachedJs(api.ServerSettings.ServerSettingsJavascriptHandler, config, serverSettingsCacheStore))

	for i := 0; i < len(workboxFileNames); i++ {
		router.StaticFile("/desktop/"+workboxFileNames[i], filepath.Join(config.StaticRootPath, workboxFileNames[i]))
	}

	if config.AvatarProvider == core.USER_AVATAR_PROVIDER_INTERNAL {
		avatarRoute := router.Group("/avatar")
		avatarRoute.Use(bindMiddleware(middlewares.JWTAuthorizationByQueryString(config), config))
		{
			avatarRoute.GET("/:fileName", bindImage(api.Users.UserGetAvatarHandler, config))
		}
	}

	if config.EnableTransactionPictures {
		pictureRoute := router.Group("/pictures")
		pictureRoute.Use(bindMiddleware(middlewares.JWTAuthorizationByQueryString(config), config))
		{
			pictureRoute.GET("/:fileName", bindImage(api.TransactionPictures.TransactionPictureGetHandler, config))
		}
	}

	if config.EnableUserCustomIcon {
		customIconRoute := router.Group("/icons")
		customIconRoute.Use(bindMiddleware(middlewares.JWTAuthorizationByQueryString(config), config))
		{
			customIconRoute.GET("/:fileName", bindImage(api.UserCustomIcons.CustomIconGetHandler, config))
		}
	}

	router.GET("/healthz.json", bindApi(api.Healths.HealthStatusHandler, config))

	proxyRoute := router.Group("/proxy")
	proxyRoute.Use(bindMiddleware(middlewares.JWTAuthorizationByQueryString(config), config))
	{
		if config.EnableMapDataFetchProxy {
			if config.MapProvider == settings.OpenStreetMapProvider ||
				config.MapProvider == settings.OpenStreetMapHumanitarianStyleProvider ||
				config.MapProvider == settings.OpenTopoMapProvider ||
				config.MapProvider == settings.OPNVKarteMapProvider ||
				config.MapProvider == settings.CyclOSMMapProvider ||
				config.MapProvider == settings.CartoDBMapProvider ||
				config.MapProvider == settings.TomTomMapProvider ||
				config.MapProvider == settings.TianDiTuProvider ||
				config.MapProvider == settings.CustomProvider {
				proxyRoute.GET("/map/tile/:zoomLevel/:coordinateX/:fileName", bindProxy(api.MapImages.MapTileImageProxyHandler, config))
			}

			if config.MapProvider == settings.TianDiTuProvider ||
				(config.MapProvider == settings.CustomProvider && config.CustomMapTileServerAnnotationLayerUrl != "") {
				proxyRoute.GET("/map/annotation/:zoomLevel/:coordinateX/:fileName", bindProxy(api.MapImages.MapAnnotationImageProxyHandler, config))
			}
		}
	}

	if config.MapProvider == settings.AmapProvider && config.AmapSecurityVerificationMethod == settings.AmapSecurityVerificationInternalProxyMethod {
		amapApiProxyRoute := router.Group("/_AMapService")
		amapApiProxyRoute.Use(bindMiddleware(middlewares.JWTAuthorizationByCookie(config), config))
		{
			amapApiProxyRoute.GET("/*action", bindProxy(api.AmapApis.AmapApiProxyHandler, config))
		}
	}

	qrCodeRoute := router.Group("/qrcode")
	qrCodeRoute.Use(bindMiddleware(middlewares.RequestId(config), config))
	{
		qrCodeCacheStore := persistence.NewInMemoryStore(time.Minute)
		qrCodeRoute.GET("/mobile_url.png", bindCachedImage(api.QrCodes.MobileUrlQrCodeHandler, config, qrCodeCacheStore))
	}

	if config.EnableMCPServer {
		mcpRoute := router.Group("/mcp")
		mcpRoute.Use(bindMiddleware(middlewares.RequestId(config), config))
		mcpRoute.Use(bindMiddleware(middlewares.RequestLog, config))
		mcpRoute.Use(bindMiddleware(middlewares.MCPServerIpLimit(config), config))
		mcpRoute.Use(bindMiddleware(middlewares.JWTMCPAuthorization(config), config))
		{
			mcpRoute.POST("", bindJSONRPCApi(map[string]core.JSONRPCApiHandlerFunc{
				"initialize":     api.ModelContextProtocols.InitializeHandler,
				"resources/list": api.ModelContextProtocols.ListResourcesHandler,
				"resources/read": api.ModelContextProtocols.ReadResourceHandler,
				"tools/list":     api.ModelContextProtocols.ListToolsHandler,
				"tools/call":     api.ModelContextProtocols.CallToolHandler,
				"ping":           api.ModelContextProtocols.PingHandler,
			}, map[string]int{
				"notifications/initialized": http.StatusAccepted,
			}, config))
			mcpRoute.GET("", bindApi(api.Default.MethodNotAllowed, config))
		}
	}

	if config.EnableOAuth2Login {
		oauth2Route := router.Group("/oauth2")
		oauth2Route.Use(bindMiddleware(middlewares.RequestId(config), config))
		oauth2Route.Use(bindMiddleware(middlewares.RequestLog, config))
		{
			oauth2Route.GET("/login", bindRedirect(api.OAuth2Authentications.LoginHandler, config))
			oauth2Route.GET("/callback", bindRedirect(api.OAuth2Authentications.CallbackHandler, config))
		}
	}

	registerAPIRoutes(router, config)

	listenAddr := fmt.Sprintf("%s:%d", config.HttpAddr, config.HttpPort)

	if config.Protocol == settings.SCHEME_SOCKET {
		log.BootInfof(c, "[webserver.startWebServer] will run at socks:%s", config.UnixSocketPath)
		err = router.RunUnix(config.UnixSocketPath)
	} else if config.Protocol == settings.SCHEME_HTTP {
		log.BootInfof(c, "[webserver.startWebServer] will run at http://%s", listenAddr)
		err = router.Run(listenAddr)
	} else if config.Protocol == settings.SCHEME_HTTPS {
		log.BootInfof(c, "[webserver.startWebServer] will run at https://%s", listenAddr)
		err = router.RunTLS(listenAddr, config.CertFile, config.CertKeyFile)
	} else {
		err = errs.ErrInvalidProtocol
	}

	if err != nil {
		log.BootErrorf(c, "[webserver.startWebServer] cannot start, because %s", err)
		return err
	}

	return nil
}
