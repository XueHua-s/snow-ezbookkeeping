package settings

import (
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"gopkg.in/ini.v1"
	"strconv"
	"strings"
)

func loadGlobalConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	if getConfigItemStringValue(configFile, sectionName, "mode") == "production" {
		config.Mode = MODE_PRODUCTION
	} else if getConfigItemStringValue(configFile, sectionName, "mode") == "development" {
		config.Mode = MODE_DEVELOPMENT
	} else {
		return errs.ErrInvalidServerMode
	}

	return nil
}

func loadServerConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	if getConfigItemStringValue(configFile, sectionName, "protocol") == "http" {
		config.Protocol = SCHEME_HTTP

		config.HttpAddr = getConfigItemStringValue(configFile, sectionName, "http_addr", defaultHttpAddr)
		config.HttpPort = getConfigItemUint16Value(configFile, sectionName, "http_port", defaultHttpPort)
	} else if getConfigItemStringValue(configFile, sectionName, "protocol") == "https" {
		config.Protocol = SCHEME_HTTPS

		config.HttpAddr = getConfigItemStringValue(configFile, sectionName, "http_addr", defaultHttpAddr)
		config.HttpPort = getConfigItemUint16Value(configFile, sectionName, "http_port", defaultHttpPort)

		config.CertFile = getConfigItemStringValue(configFile, sectionName, "cert_file")
		config.CertKeyFile = getConfigItemStringValue(configFile, sectionName, "cert_key_file")
	} else if getConfigItemStringValue(configFile, sectionName, "protocol") == "socket" {
		config.Protocol = SCHEME_SOCKET

		config.UnixSocketPath = getConfigItemStringValue(configFile, sectionName, "unix_socket")
	} else {
		return errs.ErrInvalidProtocol
	}

	config.Domain = getConfigItemStringValue(configFile, sectionName, "domain", defaultDomain)
	config.RootUrl = getConfigItemStringValue(configFile, sectionName, "root_url", defaultRootUrl)
	config.RootUrl = strings.ReplaceAll(config.RootUrl, "%(protocol)s", string(config.Protocol))
	config.RootUrl = strings.ReplaceAll(config.RootUrl, "%(domain)s", config.Domain)
	config.RootUrl = strings.ReplaceAll(config.RootUrl, "%(http_port)s", strconv.Itoa(int(config.HttpPort)))

	if config.RootUrl[len(config.RootUrl)-1] != '/' {
		config.RootUrl += "/"
	}

	staticRootPath := getConfigItemStringValue(configFile, sectionName, "static_root_path", defaultStaticRootPath)
	finalStaticRootPath, err := getFinalPath(config.WorkingPath, staticRootPath)

	if err != nil {
		return err
	}

	config.StaticRootPath = finalStaticRootPath

	config.EnableGZip = getConfigItemBoolValue(configFile, sectionName, "enable_gzip", false)
	config.EnableRequestLog = getConfigItemBoolValue(configFile, sectionName, "log_request", false)
	config.EnableRequestIdHeader = getConfigItemBoolValue(configFile, sectionName, "request_id_header", true)

	return nil
}

func loadMCPServerConfiguration(config *Config, configFile *ini.File, sectionName string) error {
	var err error

	config.EnableMCPServer = getConfigItemBoolValue(configFile, sectionName, "enable_mcp", false)
	config.MCPAllowedRemoteIPs, err = getIPPatterns(configFile, sectionName, "mcp_allowed_remote_ips", "")

	if err != nil {
		return err
	}

	return nil
}
