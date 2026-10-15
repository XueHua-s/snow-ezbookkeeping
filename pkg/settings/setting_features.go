package settings

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"gopkg.in/ini.v1"
)

func loadUserConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableUserRegister = getConfigItemBoolValue(configFile, sectionName, "enable_register", false)
	config.EnableUserVerifyEmail = getConfigItemBoolValue(configFile, sectionName, "enable_email_verify", false)
	config.EnableUserForceVerifyEmail = getConfigItemBoolValue(configFile, sectionName, "enable_force_email_verify", false)
	config.EnableUserCustomIcon = getConfigItemBoolValue(configFile, sectionName, "enable_custom_icon", false)
	config.MaxUserCustomIconFileSize = getConfigItemUint32Value(configFile, sectionName, "max_user_custom_icon_size", defaultUserCustomIconFileMaxSize)
	config.EnableTransactionPictures = getConfigItemBoolValue(configFile, sectionName, "enable_transaction_picture", false)
	config.MaxTransactionPictureFileSize = getConfigItemUint32Value(configFile, sectionName, "max_transaction_picture_size", defaultTransactionPictureFileMaxSize)
	config.EnableScheduledTransaction = getConfigItemBoolValue(configFile, sectionName, "enable_scheduled_transaction", false)

	if getConfigItemStringValue(configFile, sectionName, "avatar_provider") == string(core.USER_AVATAR_PROVIDER_INTERNAL) {
		config.AvatarProvider = core.USER_AVATAR_PROVIDER_INTERNAL
	} else if getConfigItemStringValue(configFile, sectionName, "avatar_provider") == string(core.USER_AVATAR_PROVIDER_GRAVATAR) {
		config.AvatarProvider = core.USER_AVATAR_PROVIDER_GRAVATAR
	} else if getConfigItemStringValue(configFile, sectionName, "avatar_provider") == "" {
		config.AvatarProvider = ""
	} else {
		return errs.ErrInvalidAvatarProvider
	}

	config.MaxAvatarFileSize = getConfigItemUint32Value(configFile, sectionName, "max_user_avatar_size", defaultUserAvatarFileMaxSize)
	config.DefaultFeatureRestrictions = core.ParseUserFeatureRestrictions(getConfigItemStringValue(configFile, sectionName, "default_feature_restrictions", ""))

	return nil
}

func loadDataConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableDataExport = getConfigItemBoolValue(configFile, sectionName, "enable_export", false)
	config.EnableDataImport = getConfigItemBoolValue(configFile, sectionName, "enable_import", false)
	config.MaxImportFileSize = getConfigItemUint32Value(configFile, sectionName, "max_import_file_size", defaultImportFileMaxSize)

	return nil
}

func loadTipConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.LoginPageTips = getMultiLanguageContentConfig(configFile, sectionName, "enable_tips_in_login_page", "login_page_tips_content")

	return nil
}

func loadNotificationConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.AfterRegisterNotification = getMultiLanguageContentConfig(configFile, sectionName, "enable_notification_after_register", "after_register_notification_content")
	config.AfterLoginNotification = getMultiLanguageContentConfig(configFile, sectionName, "enable_notification_after_login", "after_login_notification_content")
	config.AfterOpenNotification = getMultiLanguageContentConfig(configFile, sectionName, "enable_notification_after_open", "after_open_notification_content")

	return nil
}

func loadMapConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	mapProvider := getConfigItemStringValue(configFile, sectionName, "map_provider")

	if mapProvider == "" {
		config.MapProvider = ""
	} else if mapProvider == OpenStreetMapProvider {
		config.MapProvider = OpenStreetMapProvider
	} else if mapProvider == OpenStreetMapHumanitarianStyleProvider {
		config.MapProvider = OpenStreetMapHumanitarianStyleProvider
	} else if mapProvider == OpenTopoMapProvider {
		config.MapProvider = OpenTopoMapProvider
	} else if mapProvider == OPNVKarteMapProvider {
		config.MapProvider = OPNVKarteMapProvider
	} else if mapProvider == CyclOSMMapProvider {
		config.MapProvider = CyclOSMMapProvider
	} else if mapProvider == CartoDBMapProvider {
		config.MapProvider = CartoDBMapProvider
	} else if mapProvider == TomTomMapProvider {
		config.MapProvider = TomTomMapProvider
	} else if mapProvider == TianDiTuProvider {
		config.MapProvider = TianDiTuProvider
	} else if mapProvider == GoogleMapProvider {
		config.MapProvider = GoogleMapProvider
	} else if mapProvider == BaiduMapProvider {
		config.MapProvider = BaiduMapProvider
	} else if mapProvider == AmapProvider {
		config.MapProvider = AmapProvider
	} else if mapProvider == CustomProvider {
		config.MapProvider = CustomProvider
	} else {
		return errs.ErrInvalidMapProvider
	}

	config.EnableMapDataFetchProxy = getConfigItemBoolValue(configFile, sectionName, "map_data_fetch_proxy", false)
	config.MapProxy = getConfigItemStringValue(configFile, sectionName, "proxy", "system")
	config.TomTomMapAPIKey = getConfigItemStringValue(configFile, sectionName, "tomtom_map_api_key")
	config.TianDiTuAPIKey = getConfigItemStringValue(configFile, sectionName, "tianditu_map_app_key")
	config.GoogleMapAPIKey = getConfigItemStringValue(configFile, sectionName, "google_map_api_key")
	config.BaiduMapAK = getConfigItemStringValue(configFile, sectionName, "baidu_map_ak")
	config.AmapApplicationKey = getConfigItemStringValue(configFile, sectionName, "amap_application_key")

	amapSecurityVerificationMethod := getConfigItemStringValue(configFile, sectionName, "amap_security_verification_method")

	if amapSecurityVerificationMethod == AmapSecurityVerificationInternalProxyMethod {
		config.AmapSecurityVerificationMethod = AmapSecurityVerificationInternalProxyMethod
	} else if amapSecurityVerificationMethod == AmapSecurityVerificationExternalProxyMethod {
		config.AmapSecurityVerificationMethod = AmapSecurityVerificationExternalProxyMethod
	} else if amapSecurityVerificationMethod == AmapSecurityVerificationPlainTextMethod {
		config.AmapSecurityVerificationMethod = AmapSecurityVerificationPlainTextMethod
	} else {
		return errs.ErrInvalidAmapSecurityVerificationMethod
	}

	config.AmapApplicationSecret = getConfigItemStringValue(configFile, sectionName, "amap_application_secret")
	config.AmapApiExternalProxyUrl = getConfigItemStringValue(configFile, sectionName, "amap_api_external_proxy_url")

	config.CustomMapTileServerTileLayerUrl = getConfigItemStringValue(configFile, sectionName, "custom_map_tile_server_url")
	config.CustomMapTileServerAnnotationLayerUrl = getConfigItemStringValue(configFile, sectionName, "custom_map_tile_server_annotation_url")
	config.CustomMapTileServerMinZoomLevel = getConfigItemUint8Value(configFile, sectionName, "custom_map_tile_server_min_zoom_level", 1)
	config.CustomMapTileServerMaxZoomLevel = getConfigItemUint8Value(configFile, sectionName, "custom_map_tile_server_max_zoom_level", 18)
	config.CustomMapTileServerDefaultZoomLevel = getConfigItemUint8Value(configFile, sectionName, "custom_map_tile_server_default_zoom_level", 14)

	return nil
}

func loadExchangeRatesConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	dataSource := getConfigItemStringValue(configFile, sectionName, "data_source")

	if dataSource == CentralBankOfArgentinaDataSource ||
		dataSource == NationalBankOfBelarusDataSource ||
		dataSource == BankOfCanadaDataSource ||
		dataSource == CzechNationalBankDataSource ||
		dataSource == DanmarksNationalbankDataSource ||
		dataSource == EuroCentralBankDataSource ||
		dataSource == NationalBankOfGeorgiaDataSource ||
		dataSource == CentralBankOfHungaryDataSource ||
		dataSource == BankOfIsraelDataSource ||
		dataSource == BankOfItalyDataSource ||
		dataSource == NationalBankOfKazakhstanDataSource ||
		dataSource == CentralBankOfMalaysiaDataSource ||
		dataSource == CentralBankOfMyanmarDataSource ||
		dataSource == NorgesBankDataSource ||
		dataSource == NationalBankOfPolandDataSource ||
		dataSource == NationalBankOfRomaniaDataSource ||
		dataSource == BankOfRussiaDataSource ||
		dataSource == SwissNationalBankDataSource ||
		dataSource == NationalBankOfUkraineDataSource ||
		dataSource == CentralBankOfUzbekistanDataSource ||
		dataSource == UserCustomExchangeRatesDataSource {
		config.ExchangeRatesDataSource = dataSource
	} else {
		return errs.ErrInvalidExchangeRatesDataSource
	}

	config.ExchangeRatesProxy = getConfigItemStringValue(configFile, sectionName, "proxy", "system")
	config.ExchangeRatesRequestTimeout = getConfigItemUint32Value(configFile, sectionName, "request_timeout", defaultExchangeRatesDataRequestTimeout)

	if config.ExchangeRatesRequestTimeout > defaultExchangeRatesDataRequestTimeout {
		config.ExchangeRatesRequestTimeoutExceedDefaultValue = true
	}

	config.ExchangeRatesSkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "skip_tls_verify", false)

	return nil
}

func loadDeveloperTools(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableLanguagePreview = getConfigItemBoolValue(configFile, sectionName, "enable_language_preview", false)

	return nil
}
