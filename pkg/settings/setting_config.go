package settings

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"net"
	"time"
)

// MultiLanguageContentConfig represents a multi-language content setting config
type MultiLanguageContentConfig struct {
	Enabled              bool
	DefaultContent       string
	MultiLanguageContent map[string]string
}

// Config represents the global setting config
type Config struct {
	// Global
	Mode        SystemMode
	WorkingPath string

	// Server
	Protocol Scheme
	HttpAddr string
	HttpPort uint16

	Domain  string
	RootUrl string

	CertFile    string
	CertKeyFile string

	UnixSocketPath string

	StaticRootPath string

	EnableGZip            bool
	EnableRequestLog      bool
	EnableRequestIdHeader bool

	// MCP
	EnableMCPServer     bool
	MCPAllowedRemoteIPs []*core.IPPattern

	// Database
	DatabaseConfig     *DatabaseConfig
	EnableQueryLog     bool
	AutoUpdateDatabase bool

	// Mail
	EnableSMTP bool
	SMTPConfig *SMTPConfig

	// Log
	LogModes         []string
	EnableConsoleLog bool
	EnableFileLog    bool

	EnableDebugLog     bool
	LogLevel           Level
	FileLogPath        string
	RequestFileLogPath string
	QueryFileLogPath   string
	LogFileRotate      bool
	LogFileMaxSize     uint32
	LogFileMaxDays     uint32

	// Storage
	StorageType         string
	LocalFileSystemPath string
	S3Config            *S3Config
	MinIOConfig         *MinIOConfig
	WebDAVConfig        *WebDAVConfig

	// Large Language Model
	TransactionFromAITextRecognition  bool
	TransactionFromAIImageRecognition bool
	InsightsExplorerCodingAssistant   bool
	MaxAIRecognitionPictureFileSize   uint32
	EnableAIAssistant                 bool

	// Large Language Model for Transaction Text Recognition
	TextRecognitionLLMConfig *LLMConfig

	// Large Language Model for Receipt Image Recognition
	ReceiptImageRecognitionLLMConfig         *LLMConfig
	ReceiptImageRecognitionFallbackLLMConfig *LLMConfig
	AIAssistantLLMConfig                     *LLMConfig
	AIAssistantFallbackLLMConfig             *LLMConfig
	AIAssistantRuntimeConfig                 AIAssistantRuntimeConfig

	CodingAssistantLLMConfig *LLMConfig

	// Uuid
	UuidGeneratorType string
	UuidServerId      uint8

	// Duplicate Checker
	DuplicateCheckerType                            string
	InMemoryDuplicateCheckerCleanupInterval         uint32
	InMemoryDuplicateCheckerCleanupIntervalDuration time.Duration
	EnableDuplicateSubmissionsCheck                 bool
	DuplicateSubmissionsInterval                    uint32
	DuplicateSubmissionsIntervalDuration            time.Duration

	// Cron
	EnableRemoveExpiredTokens        bool
	EnableCreateScheduledTransaction bool

	// Secret
	SecretKeyNoSet                        bool
	SecretKey                             string
	TrustedProxyIPs                       []*net.IPNet
	TrustedProxyTextualIPs                []string
	TokenExpiredTime                      uint32
	TokenExpiredTimeDuration              time.Duration
	TokenMinRefreshInterval               uint32
	TemporaryTokenExpiredTime             uint32
	TemporaryTokenExpiredTimeDuration     time.Duration
	EmailVerifyTokenExpiredTime           uint32
	EmailVerifyTokenExpiredTimeDuration   time.Duration
	PasswordResetTokenExpiredTime         uint32
	PasswordResetTokenExpiredTimeDuration time.Duration
	EnableAPIToken                        bool
	APITokenAllowedRemoteIPs              []*core.IPPattern
	MaxFailuresPerIpPerMinute             uint32
	MaxFailuresPerUserPerMinute           uint32

	// Auth
	EnableInternalAuth                bool
	EnableOAuth2Login                 bool
	EnableTwoFactor                   bool
	EnableUserForgetPassword          bool
	ForgetPasswordRequireVerifyEmail  bool
	OAuth2ClientID                    string
	OAuth2ClientSecret                string
	OAuth2UsePKCE                     bool
	OAuth2UserIdentifier              string
	OAuth2AutoRegister                bool
	OAuth2Provider                    string
	OAuth2StateExpiredTime            uint32
	OAuth2StateExpiredTimeDuration    time.Duration
	OAuth2RequestTimeout              uint32
	OAuth2Proxy                       string
	OAuth2SkipTLSVerify               bool
	OAuth2OIDCProviderIssuerURL       string
	OAuth2OIDCProviderCheckIssuerURL  bool
	OAuth2OIDCCustomDisplayNameConfig MultiLanguageContentConfig
	OAuth2NextcloudBaseUrl            string
	OAuth2GiteaBaseUrl                string

	// User
	EnableUserRegister            bool
	EnableUserVerifyEmail         bool
	EnableUserForceVerifyEmail    bool
	EnableUserCustomIcon          bool
	MaxUserCustomIconFileSize     uint32
	EnableTransactionPictures     bool
	MaxTransactionPictureFileSize uint32
	EnableScheduledTransaction    bool
	AvatarProvider                core.UserAvatarProviderType
	MaxAvatarFileSize             uint32
	DefaultFeatureRestrictions    core.UserFeatureRestrictions

	// Data
	EnableDataExport  bool
	EnableDataImport  bool
	MaxImportFileSize uint32

	// Tip
	LoginPageTips MultiLanguageContentConfig

	// Notification
	AfterRegisterNotification MultiLanguageContentConfig
	AfterLoginNotification    MultiLanguageContentConfig
	AfterOpenNotification     MultiLanguageContentConfig

	// Map
	MapProvider                           string
	EnableMapDataFetchProxy               bool
	MapProxy                              string
	TomTomMapAPIKey                       string
	TianDiTuAPIKey                        string
	GoogleMapAPIKey                       string
	BaiduMapAK                            string
	AmapApplicationKey                    string
	AmapSecurityVerificationMethod        string
	AmapApplicationSecret                 string
	AmapApiExternalProxyUrl               string
	CustomMapTileServerTileLayerUrl       string
	CustomMapTileServerAnnotationLayerUrl string
	CustomMapTileServerMinZoomLevel       uint8
	CustomMapTileServerMaxZoomLevel       uint8
	CustomMapTileServerDefaultZoomLevel   uint8

	// Exchange Rates
	ExchangeRatesDataSource                       string
	ExchangeRatesRequestTimeout                   uint32
	ExchangeRatesRequestTimeoutExceedDefaultValue bool
	ExchangeRatesProxy                            string
	ExchangeRatesSkipTLSVerify                    bool

	// Developer Tools
	EnableLanguagePreview bool
}
