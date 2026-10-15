package settings

import (
	"fmt"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/locales"
	"gopkg.in/ini.v1"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getWorkingPath() (string, error) {
	workingPath := os.Getenv(ebkWorkDirEnvName)

	if workingPath != "" {
		return workingPath, nil
	}

	execFilePath, err := os.Getwd()

	if err != nil {
		return "", err
	}

	return execFilePath, nil
}

func getFinalPath(workingPath, p string) (string, error) {
	var err error

	if !filepath.IsAbs(p) {
		p = filepath.Join(workingPath, p)
	}

	if _, err = os.Stat(p); err == nil {
		return p, nil
	}

	return p, err
}

func getIPPatterns(configFile *ini.File, sectionName string, itemName string, defaultValue string) ([]*core.IPPattern, error) {
	configValue := getConfigItemStringValue(configFile, sectionName, itemName, defaultValue)

	if configValue == "" {
		return nil, nil
	}

	remoteIPs := strings.Split(configValue, ",")
	ipPatterns := make([]*core.IPPattern, 0, len(remoteIPs))

	for i := 0; i < len(remoteIPs); i++ {
		ip := strings.TrimSpace(remoteIPs[i])
		pattern, err := core.ParseIPPattern(ip)

		if err != nil {
			return nil, err
		}

		if pattern == nil {
			continue
		}

		ipPatterns = append(ipPatterns, pattern)
	}

	return ipPatterns, nil
}

func getCIDRList(configFile *ini.File, sectionName string, itemName string, defaultValue string) ([]*net.IPNet, []string, error) {
	configValue := getConfigItemStringValue(configFile, sectionName, itemName, defaultValue)

	if configValue == "" {
		return nil, nil, nil
	}

	cidrs := strings.Split(configValue, ",")
	parsedCIDRs := make([]*net.IPNet, 0, len(cidrs))
	textualCIDRs := make([]string, 0, len(cidrs))

	for i := 0; i < len(cidrs); i++ {
		cidr := strings.TrimSpace(cidrs[i])

		if cidr == "" {
			continue
		}

		_, parsedCIDR, err := net.ParseCIDR(cidr)

		if err != nil {
			return nil, nil, err
		}

		parsedCIDRs = append(parsedCIDRs, parsedCIDR)
		textualCIDRs = append(textualCIDRs, parsedCIDR.String())
	}

	return parsedCIDRs, textualCIDRs, nil
}

func getMultiLanguageContentConfig(configFile *ini.File, sectionName string, enableKey string, contentKey string) MultiLanguageContentConfig {
	config := MultiLanguageContentConfig{
		Enabled:              getConfigItemBoolValue(configFile, sectionName, enableKey, false),
		DefaultContent:       getConfigItemStringValue(configFile, sectionName, contentKey, ""),
		MultiLanguageContent: make(map[string]string),
	}

	for languageTag := range locales.AllLanguages {
		multiLanguageContentKey := strings.ToLower(languageTag)
		multiLanguageContentKey = strings.Replace(multiLanguageContentKey, "-", "_", -1)
		multiLanguageContentKey = contentKey + "_" + multiLanguageContentKey
		content := getConfigItemStringValue(configFile, sectionName, multiLanguageContentKey, "")

		if content != "" {
			config.MultiLanguageContent[languageTag] = content
		}
	}

	return config
}

func getConfigItemIsSet(configFile *ini.File, sectionName string, itemName string) bool {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		return true
	}

	section := configFile.Section(sectionName)

	if !section.HasKey(itemName) {
		return false
	}

	return section.Key(itemName).Value() != ""
}

func getConfigItemStringValue(configFile *ini.File, sectionName string, itemName string, defaultValue ...string) string {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		return environmentValue
	}

	section := configFile.Section(sectionName)
	key := section.Key(itemName)
	value := key.Value()

	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return value
	}
}

func getConfigItemUint8Value(configFile *ini.File, sectionName string, itemName string, defaultValue uint8) uint8 {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		value, err := strconv.ParseUint(environmentValue, 10, 8)

		if err == nil {
			return uint8(value)
		}
	}

	section := configFile.Section(sectionName)
	value, err := strconv.ParseUint(section.Key(itemName).Value(), 10, 8)

	if err == nil {
		return uint8(value)
	}

	return defaultValue
}

func getConfigItemUint16Value(configFile *ini.File, sectionName string, itemName string, defaultValue uint16) uint16 {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		value, err := strconv.ParseUint(environmentValue, 10, 16)

		if err == nil {
			return uint16(value)
		}
	}

	section := configFile.Section(sectionName)
	value, err := strconv.ParseUint(section.Key(itemName).Value(), 10, 16)

	if err == nil {
		return uint16(value)
	}

	return defaultValue
}

func getConfigItemUint32Value(configFile *ini.File, sectionName string, itemName string, defaultValue uint32) uint32 {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		value, err := strconv.ParseUint(environmentValue, 10, 32)

		if err == nil {
			return uint32(value)
		}
	}

	section := configFile.Section(sectionName)
	value, err := strconv.ParseUint(section.Key(itemName).Value(), 10, 32)

	if err == nil {
		return uint32(value)
	}

	return defaultValue
}

func getConfigItemBoolValue(configFile *ini.File, sectionName string, itemName string, defaultValue bool) bool {
	environmentValue := getConfigItemValueFromEnvironment(sectionName, itemName)

	if len(environmentValue) > 0 {
		value, err := strconv.ParseBool(environmentValue)

		if err == nil {
			return value
		}
	}

	section := configFile.Section(sectionName)
	value, err := strconv.ParseBool(section.Key(itemName).Value())

	if err == nil {
		return value
	}

	return defaultValue
}

func getConfigItemValueFromEnvironment(sectionName string, itemName string) string {
	itemFilePathEnvironmentKey := getConfigItemFilePathEnvironmentKey(sectionName, itemName)
	itemFilePath := os.Getenv(itemFilePathEnvironmentKey)

	if itemFilePath != "" {
		content, err := os.ReadFile(itemFilePath)

		if err == nil {
			return strings.TrimRight(string(content), "\r\n")
		}
	}

	itemValueEnvironmentKey := getConfigItemValueEnvironmentKey(sectionName, itemName)
	itemValueEnvironmentValue := os.Getenv(itemValueEnvironmentKey)

	return itemValueEnvironmentValue
}

func getConfigItemFilePathEnvironmentKey(sectionName string, itemName string) string {
	return fmt.Sprintf("%s_%s_%s", ebkConfigItemFilePathEnvNamePrefix, strings.ToUpper(sectionName), strings.ToUpper(itemName))
}

func getConfigItemValueEnvironmentKey(sectionName string, itemName string) string {
	return fmt.Sprintf("%s_%s_%s", ebkConfigItemValueEnvNamePrefix, strings.ToUpper(sectionName), strings.ToUpper(itemName))
}

func getLogLevel(logLevelStr string) (Level, error) {
	if logLevelStr == "debug" {
		return LOGLEVEL_DEBUG, nil
	} else if logLevelStr == "info" {
		return LOGLEVEL_INFO, nil
	} else if logLevelStr == "warn" {
		return LOGLEVEL_WARN, nil
	} else if logLevelStr == "error" {
		return LOGLEVEL_ERROR, nil
	}

	return "", errs.ErrInvalidLogLevel
}
