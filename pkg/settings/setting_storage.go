package settings

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"gopkg.in/ini.v1"
	"strings"
)

func loadDatabaseConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	dbConfig := &DatabaseConfig{}

	dbConfig.DatabaseType = getConfigItemStringValue(configFile, sectionName, "type", MySqlDbType)

	if dbConfig.DatabaseType != MySqlDbType &&
		dbConfig.DatabaseType != PostgresDbType &&
		dbConfig.DatabaseType != Sqlite3DbType {
		return errs.ErrDatabaseTypeInvalid
	}

	dbConfig.DatabaseHost = getConfigItemStringValue(configFile, sectionName, "host", defaultDatabaseHost)
	dbConfig.DatabaseName = getConfigItemStringValue(configFile, sectionName, "name", defaultDatabaseName)
	dbConfig.DatabaseUser = getConfigItemStringValue(configFile, sectionName, "user")
	dbConfig.DatabasePassword = getConfigItemStringValue(configFile, sectionName, "passwd")

	if dbConfig.DatabaseType == PostgresDbType {
		dbConfig.DatabaseSSLMode = getConfigItemStringValue(configFile, sectionName, "ssl_mode")
	}

	if dbConfig.DatabaseType == Sqlite3DbType {
		staticDBPath := getConfigItemStringValue(configFile, sectionName, "db_path")
		finalStaticDBPath, _ := getFinalPath(config.WorkingPath, staticDBPath)
		dbConfig.DatabasePath = finalStaticDBPath
	}

	dbConfig.MaxIdleConnection = getConfigItemUint16Value(configFile, sectionName, "max_idle_conn", defaultDatabaseMaxIdleConn)
	dbConfig.MaxOpenConnection = getConfigItemUint16Value(configFile, sectionName, "max_open_conn", defaultDatabaseMaxOpenConn)
	dbConfig.ConnectionMaxLifeTime = getConfigItemUint32Value(configFile, sectionName, "conn_max_lifetime", defaultDatabaseConnMaxLifetime)

	config.DatabaseConfig = dbConfig
	config.EnableQueryLog = getConfigItemBoolValue(configFile, sectionName, "log_query", false)
	config.AutoUpdateDatabase = getConfigItemBoolValue(configFile, sectionName, "auto_update_database", true)

	return nil
}

func loadMailConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.EnableSMTP = getConfigItemBoolValue(configFile, sectionName, "enable_smtp", false)

	smtpConfig := &SMTPConfig{}
	smtpConfig.SMTPHost = getConfigItemStringValue(configFile, sectionName, "smtp_host")
	smtpConfig.SMTPUser = getConfigItemStringValue(configFile, sectionName, "smtp_user")
	smtpConfig.SMTPPasswd = getConfigItemStringValue(configFile, sectionName, "smtp_passwd")
	smtpConfig.SMTPSkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "smtp_skip_tls_verify", false)

	smtpConfig.FromAddress = getConfigItemStringValue(configFile, sectionName, "from_address")

	config.SMTPConfig = smtpConfig

	return nil
}

func loadLogConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	config.LogModes = strings.Split(getConfigItemStringValue(configFile, sectionName, "mode", defaultLogMode), " ")

	for i := 0; i < len(config.LogModes); i++ {
		logMode := config.LogModes[i]

		if logMode == "console" {
			config.EnableConsoleLog = true
		} else if logMode == "file" {
			config.EnableFileLog = true
		} else {
			return errs.ErrInvalidLogMode
		}
	}

	var err error
	config.LogLevel, err = getLogLevel(getConfigItemStringValue(configFile, sectionName, "level"))

	if err != nil {
		return err
	}

	if config.LogLevel != LOGLEVEL_DEBUG &&
		config.LogLevel != LOGLEVEL_INFO &&
		config.LogLevel != LOGLEVEL_WARN &&
		config.LogLevel != LOGLEVEL_ERROR {
		return errs.ErrInvalidLogLevel
	}

	if config.LogLevel == LOGLEVEL_DEBUG {
		config.EnableDebugLog = true
	} else {
		config.EnableDebugLog = false
	}

	if config.EnableFileLog {
		fileLogPath := getConfigItemStringValue(configFile, sectionName, "log_path")
		finalFileLogPath, _ := getFinalPath(config.WorkingPath, fileLogPath)
		config.FileLogPath = finalFileLogPath

		requestFileLogPath := getConfigItemStringValue(configFile, sectionName, "request_log_path")

		if requestFileLogPath != "" {
			finalRequestFileLogPath, _ := getFinalPath(config.WorkingPath, requestFileLogPath)
			config.RequestFileLogPath = finalRequestFileLogPath
		} else {
			config.RequestFileLogPath = ""
		}

		queryFileLogPath := getConfigItemStringValue(configFile, sectionName, "query_log_path")

		if queryFileLogPath != "" {
			finalQueryFileLogPath, _ := getFinalPath(config.WorkingPath, queryFileLogPath)
			config.QueryFileLogPath = finalQueryFileLogPath
		} else {
			config.QueryFileLogPath = ""
		}

		config.LogFileRotate = getConfigItemBoolValue(configFile, sectionName, "log_file_rotate", false)
		config.LogFileMaxSize = getConfigItemUint32Value(configFile, sectionName, "log_file_max_size", defaultLogFileMaxSize)
		config.LogFileMaxDays = getConfigItemUint32Value(configFile, sectionName, "log_file_max_days", defaultLogFileMaxDays)
	}

	return nil
}

func loadStorageConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	if getConfigItemStringValue(configFile, sectionName, "type") == LocalFileSystemObjectStorageType {
		config.StorageType = LocalFileSystemObjectStorageType
	} else if getConfigItemStringValue(configFile, sectionName, "type") == S3StorageType {
		config.StorageType = S3StorageType
	} else if getConfigItemStringValue(configFile, sectionName, "type") == MinIOStorageType {
		config.StorageType = MinIOStorageType
	} else if getConfigItemStringValue(configFile, sectionName, "type") == WebDAVStorageType {
		config.StorageType = WebDAVStorageType
	} else {
		return errs.ErrInvalidStorageType
	}

	localFileSystemRootPath := getConfigItemStringValue(configFile, sectionName, "local_filesystem_path")
	finalLocalFileSystemRootPath, err := getFinalPath(config.WorkingPath, localFileSystemRootPath)
	config.LocalFileSystemPath = finalLocalFileSystemRootPath

	if config.StorageType == LocalFileSystemObjectStorageType && err != nil {
		return errs.ErrInvalidLocalFileSystemStoragePath
	}

	s3Config := &S3Config{}
	s3Config.Endpoint = getConfigItemStringValue(configFile, sectionName, "s3_endpoint")
	s3Config.Region = getConfigItemStringValue(configFile, sectionName, "s3_region")
	s3Config.AccessKeyID = getConfigItemStringValue(configFile, sectionName, "s3_access_key_id")
	s3Config.SecretAccessKey = getConfigItemStringValue(configFile, sectionName, "s3_secret_access_key")
	s3Config.SessionToken = getConfigItemStringValue(configFile, sectionName, "s3_session_token")
	s3Config.UseSSL = getConfigItemBoolValue(configFile, sectionName, "s3_use_ssl", false)
	s3Config.SkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "s3_skip_tls_verify", false)
	s3Config.UsePathStyle = getConfigItemBoolValue(configFile, sectionName, "s3_use_path_style", false)
	s3Config.Bucket = getConfigItemStringValue(configFile, sectionName, "s3_bucket")
	s3Config.RootPath = getConfigItemStringValue(configFile, sectionName, "s3_root_path")
	config.S3Config = s3Config

	minIOConfig := &MinIOConfig{}
	minIOConfig.Endpoint = getConfigItemStringValue(configFile, sectionName, "minio_endpoint")
	minIOConfig.Location = getConfigItemStringValue(configFile, sectionName, "minio_location")
	minIOConfig.AccessKeyID = getConfigItemStringValue(configFile, sectionName, "minio_access_key_id")
	minIOConfig.SecretAccessKey = getConfigItemStringValue(configFile, sectionName, "minio_secret_access_key")
	minIOConfig.UseSSL = getConfigItemBoolValue(configFile, sectionName, "minio_use_ssl", false)
	minIOConfig.SkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "minio_skip_tls_verify", false)
	minIOConfig.Bucket = getConfigItemStringValue(configFile, sectionName, "minio_bucket")
	minIOConfig.RootPath = getConfigItemStringValue(configFile, sectionName, "minio_root_path")
	config.MinIOConfig = minIOConfig

	webDAVConfig := &WebDAVConfig{}
	webDAVConfig.Url = getConfigItemStringValue(configFile, sectionName, "webdav_url")
	webDAVConfig.Username = getConfigItemStringValue(configFile, sectionName, "webdav_username")
	webDAVConfig.Password = getConfigItemStringValue(configFile, sectionName, "webdav_password")
	webDAVConfig.RootPath = getConfigItemStringValue(configFile, sectionName, "webdav_root_path")
	webDAVConfig.RequestTimeout = getConfigItemUint32Value(configFile, sectionName, "webdav_request_timeout", defaultWebDAVRequestTimeout)
	webDAVConfig.Proxy = getConfigItemStringValue(configFile, sectionName, "webdav_proxy", "system")
	webDAVConfig.SkipTLSVerify = getConfigItemBoolValue(configFile, sectionName, "webdav_skip_tls_verify", false)
	config.WebDAVConfig = webDAVConfig

	return nil
}
