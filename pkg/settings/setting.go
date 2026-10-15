package settings

const (
	ebkWorkDirEnvName                  = "EBK_WORK_DIR"
	ebkConfigItemValueEnvNamePrefix    = "EBK"
	ebkConfigItemFilePathEnvNamePrefix = "EBKCFP"
	defaultConfigPath                  = "/conf/ezbookkeeping.ini"
	defaultRootUrl                     = "%(protocol)s://%(domain)s:%(http_port)s/"
	defaultStaticRootPath              = "public"
)

// SystemMode represents running mode of system
type SystemMode string

// System running modes
const (
	MODE_DEVELOPMENT SystemMode = "development"
	MODE_PRODUCTION  SystemMode = "production"
)

// Scheme represents how the web backend service exposes
type Scheme string

// Scheme types
const (
	SCHEME_HTTP   Scheme = "http"
	SCHEME_HTTPS  Scheme = "https"
	SCHEME_SOCKET Scheme = "socket"
)

// Level represents log level
type Level string

// Log levels
const (
	LOGLEVEL_DEBUG Level = "debug"
	LOGLEVEL_INFO  Level = "info"
	LOGLEVEL_WARN  Level = "warn"
	LOGLEVEL_ERROR Level = "error"
)

// Database types
const (
	MySqlDbType    string = "mysql"
	PostgresDbType string = "postgres"
	Sqlite3DbType  string = "sqlite3"
)

// Object Storage types
const (
	LocalFileSystemObjectStorageType string = "local_filesystem"
	S3StorageType                    string = "s3"
	MinIOStorageType                 string = "minio"
	WebDAVStorageType                string = "webdav"
)

const (
	OpenAILLMProvider                    string = "openai"
	OpenAICompatibleLLMProvider          string = "openai_compatible"
	OpenAIResponsesCompatibleLLMProvider string = "openai_responses_compatible"
	AnthropicLLMProvider                 string = "anthropic"
	AnthropicCompatibleLLMProvider       string = "anthropic_compatible"
	OpenRouterLLMProvider                string = "openrouter"
	OllamaLLMProvider                    string = "ollama"
	LMStudioLLMProvider                  string = "lm_studio"
	GoogleAILLMProvider                  string = "google_ai"
)

// LLMThinkingLevel represents the thinking level of a large language model
type LLMThinkingLevel string

// LLM thinking levels
const (
	LLMThinkingDefault  LLMThinkingLevel = ""
	LLMThinkingDisabled LLMThinkingLevel = "off"
	LLMThinkingEnabled  LLMThinkingLevel = "on"
	LLMThinkingLow      LLMThinkingLevel = "low"
	LLMThinkingMedium   LLMThinkingLevel = "medium"
	LLMThinkingHigh     LLMThinkingLevel = "high"
	LLMThinkingXHigh    LLMThinkingLevel = "xhigh"
)

// Uuid generator types
const (
	InternalUuidGeneratorType string = "internal"
)

// Duplicate checker types
const (
	InMemoryDuplicateCheckerType string = "in_memory"
)

// OAuth 2.0 user identifier types
const (
	OAuth2UserIdentifierEmail    string = "email"
	OAuth2UserIdentifierUsername string = "username"
)

// OAuth 2.0 provider types
const (
	OAuth2ProviderOIDC      string = "oidc"
	OAuth2ProviderNextcloud string = "nextcloud"
	OAuth2ProviderGitea     string = "gitea"
	OAuth2ProviderGithub    string = "github"
)

// Map provider types
const (
	OpenStreetMapProvider                  string = "openstreetmap"
	OpenStreetMapHumanitarianStyleProvider string = "openstreetmap_humanitarian"
	OpenTopoMapProvider                    string = "opentopomap"
	OPNVKarteMapProvider                   string = "opnvkarte"
	CyclOSMMapProvider                     string = "cyclosm"
	CartoDBMapProvider                     string = "cartodb"
	TomTomMapProvider                      string = "tomtom"
	TianDiTuProvider                       string = "tianditu"
	GoogleMapProvider                      string = "googlemap"
	BaiduMapProvider                       string = "baidumap"
	AmapProvider                           string = "amap"
	CustomProvider                         string = "custom"
)

// Amap security verification method
const (
	AmapSecurityVerificationInternalProxyMethod string = "internal_proxy"
	AmapSecurityVerificationExternalProxyMethod string = "external_proxy"
	AmapSecurityVerificationPlainTextMethod     string = "plain_text"
)

// Exchange rates data source types
const (
	CentralBankOfArgentinaDataSource   string = "central_bank_of_argentina"
	NationalBankOfBelarusDataSource    string = "national_bank_of_belarus"
	BankOfCanadaDataSource             string = "bank_of_canada"
	CzechNationalBankDataSource        string = "czech_national_bank"
	DanmarksNationalbankDataSource     string = "danmarks_national_bank"
	EuroCentralBankDataSource          string = "euro_central_bank"
	NationalBankOfGeorgiaDataSource    string = "national_bank_of_georgia"
	CentralBankOfHungaryDataSource     string = "central_bank_of_hungary"
	BankOfIsraelDataSource             string = "bank_of_israel"
	BankOfItalyDataSource              string = "bank_of_italy"
	NationalBankOfKazakhstanDataSource string = "national_bank_of_kazakhstan"
	CentralBankOfMalaysiaDataSource    string = "central_bank_of_malaysia"
	CentralBankOfMyanmarDataSource     string = "central_bank_of_myanmar"
	NorgesBankDataSource               string = "norges_bank"
	NationalBankOfPolandDataSource     string = "national_bank_of_poland"
	NationalBankOfRomaniaDataSource    string = "national_bank_of_romania"
	BankOfRussiaDataSource             string = "bank_of_russia"
	SwissNationalBankDataSource        string = "swiss_national_bank"
	NationalBankOfUkraineDataSource    string = "national_bank_of_ukraine"
	CentralBankOfUzbekistanDataSource  string = "central_bank_of_uzbekistan"
	UserCustomExchangeRatesDataSource  string = "user_custom"
)

const (
	defaultHttpAddr string = "0.0.0.0"
	defaultHttpPort uint16 = 8080
	defaultDomain   string = "localhost"

	defaultDatabaseHost            string = "127.0.0.1:3306"
	defaultDatabaseName            string = "ezbookkeeping"
	defaultDatabaseMaxIdleConn     uint16 = 2
	defaultDatabaseMaxOpenConn     uint16 = 0
	defaultDatabaseConnMaxLifetime uint32 = 14400

	defaultLogMode        string = "console"
	defaultLogFileMaxSize uint32 = 104857600 // 100 MB
	defaultLogFileMaxDays uint32 = 7         // days

	defaultWebDAVRequestTimeout uint32 = 10000 // 10 seconds

	defaultAIRecognitionPictureMaxSize                 uint32 = 10485760 // 10MB
	defaultAnthropicLargeLanguageModelAPIMaximumTokens uint32 = 1024
	defaultAnthropicLargeLanguageModelThinkingBudget   uint32 = 1024
	defaultLargeLanguageModelAPIRequestTimeout         uint32 = 60000 // 60 seconds
	defaultOpenAIBaseURL                               string = "https://api.openai.com/v1/"
	defaultAIAssistantOpenAIModelID                    string = "gpt-5.4-mini"
	defaultAIAssistantRunTimeout                       uint32 = 180000 // 180 seconds
	defaultAIAssistantQueryTimeout                     uint32 = 30000  // 30 seconds
	defaultAIAssistantMaxRounds                        uint32 = 12

	defaultInMemoryDuplicateCheckerCleanupInterval uint32 = 60  // 1 minutes
	defaultDuplicateSubmissionsInterval            uint32 = 300 // 5 minutes

	defaultSecretKey                     string = "ezbookkeeping"
	defaultTrustedProxyIPs               string = "10.0.0.0/8,169.254.0.0/16,127.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
	defaultTokenExpiredTime              uint32 = 2592000 // 30 days
	defaultTokenMinRefreshInterval       uint32 = 86400   // 1 day
	defaultTemporaryTokenExpiredTime     uint32 = 300     // 5 minutes
	defaultEmailVerifyTokenExpiredTime   uint32 = 3600    // 60 minutes
	defaultPasswordResetTokenExpiredTime uint32 = 3600    // 60 minutes
	defaultMaxFailuresPerIpPerMinute     uint32 = 5
	defaultMaxFailuresPerUserPerMinute   uint32 = 5

	defaultOAuth2StateExpiredTime uint32 = 300   // 5 minutes
	defaultOAuth2RequestTimeout   uint32 = 10000 // 10 seconds

	defaultUserCustomIconFileMaxSize     uint32 = 1048576  // 1MB
	defaultTransactionPictureFileMaxSize uint32 = 10485760 // 10MB
	defaultUserAvatarFileMaxSize         uint32 = 1048576  // 1MB

	defaultImportFileMaxSize uint32 = 10485760 // 10MB

	defaultExchangeRatesDataRequestTimeout uint32 = 10000 // 10 seconds
)
