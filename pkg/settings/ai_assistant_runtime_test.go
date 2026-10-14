package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestAIAssistantRuntimeConfigurationDefaults(t *testing.T) {
	file := ini.Empty()
	// Runtime budgets only belong to the primary assistant section.
	file.Section("llm_assistant_fallback").Key("run_timeout").SetValue("0")
	file.Section("llm_assistant_fallback").Key("query_timeout").SetValue("-1")
	file.Section("llm_assistant_fallback").Key("max_rounds").SetValue("4294967296")
	config, err := loadAIAssistantRuntimeConfiguration(file)
	require.NoError(t, err)
	require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 180000, QueryTimeout: 30000, MaxRounds: 12}, config)
}

func TestGetAIAssistantRuntimeConfigDefaults(t *testing.T) {
	config := &Config{}
	require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 180000, QueryTimeout: 30000, MaxRounds: 12}, config.GetAIAssistantRuntimeConfig())
	require.Equal(t, AIAssistantRuntimeConfig{}, config.AIAssistantRuntimeConfig)
	config.AIAssistantRuntimeConfig.QueryTimeout = 1
	require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 180000, QueryTimeout: 1, MaxRounds: 12}, config.GetAIAssistantRuntimeConfig())
}

func TestAIAssistantRuntimeConfigurationValidation(t *testing.T) {
	for _, key := range []string{"run_timeout", "query_timeout", "max_rounds"} {
		for _, value := range []string{"", "0", "-1", "1.5", "1e3", "invalid", "4294967296", "18446744073709551616"} {
			t.Run(key+"/"+value, func(t *testing.T) {
				file := ini.Empty()
				file.Section("llm_assistant").Key(key).SetValue(value)
				_, err := loadAIAssistantRuntimeConfiguration(file)
				require.ErrorContains(t, err, "invalid [llm_assistant] "+key)
			})
		}
	}
	for _, value := range []string{"1", "4294967295"} {
		file := ini.Empty()
		for _, key := range []string{"run_timeout", "query_timeout", "max_rounds"} {
			file.Section("llm_assistant").Key(key).SetValue(value)
		}
		config, err := loadAIAssistantRuntimeConfiguration(file)
		require.NoError(t, err)
		if value == "1" {
			require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 1, QueryTimeout: 1, MaxRounds: 1}, config)
		} else {
			require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: 4294967295, QueryTimeout: 4294967295, MaxRounds: 4294967295}, config)
		}
	}
}

func TestAIAssistantRuntimeEnvironmentOverrides(t *testing.T) {
	for _, key := range []string{"run_timeout", "query_timeout", "max_rounds"} {
		t.Run(key, func(t *testing.T) {
			file := ini.Empty()
			file.Section("llm_assistant").Key(key).SetValue("2")
			valueEnv := getConfigItemValueEnvironmentKey("llm_assistant", key)
			fileEnv := getConfigItemFilePathEnvironmentKey("llm_assistant", key)
			t.Setenv(valueEnv, "3")
			t.Setenv(fileEnv, "")
			value, err := getPositiveConfigItemUint32Value(file, "llm_assistant", key, 1)
			require.NoError(t, err)
			require.Equal(t, uint32(3), value)

			path := filepath.Join(t.TempDir(), "budget")
			require.NoError(t, os.WriteFile(path, []byte("4\r\n"), 0600))
			t.Setenv(fileEnv, path)
			value, err = getPositiveConfigItemUint32Value(file, "llm_assistant", key, 1)
			require.NoError(t, err)
			require.Equal(t, uint32(4), value)

			for _, invalid := range []string{"", "0", "-1", "4294967296", "secret-canary"} {
				require.NoError(t, os.WriteFile(path, []byte(invalid+"\n"), 0600))
				_, err = getPositiveConfigItemUint32Value(file, "llm_assistant", key, 1)
				require.ErrorContains(t, err, key)
				require.NotContains(t, err.Error(), "secret-canary")
			}
			t.Setenv(fileEnv, filepath.Join(t.TempDir(), "missing"))
			_, err = getPositiveConfigItemUint32Value(file, "llm_assistant", key, 1)
			require.ErrorContains(t, err, "cannot read environment override file")

			t.Setenv(fileEnv, "")
			for _, invalid := range []string{"", "0", "-1", "4294967296", "secret-canary"} {
				t.Setenv(valueEnv, invalid)
				_, err = getPositiveConfigItemUint32Value(file, "llm_assistant", key, 1)
				require.ErrorContains(t, err, key)
				require.NotContains(t, err.Error(), "secret-canary")
			}
		})
	}
}

func TestLoadConfigurationAIAssistantRuntime(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		wantFail bool
	}{
		{"defaults", "", false},
		{"custom", "240000", false},
		{"invalid", "0", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			file, err := ini.Load("../../conf/ezbookkeeping.ini")
			require.NoError(t, err)
			section := file.Section("llm_assistant")
			for _, key := range []string{"run_timeout", "query_timeout", "max_rounds"} {
				section.DeleteKey(key)
			}
			section.Key("llm_provider").SetValue("openai")
			if test.value != "" {
				section.Key("run_timeout").SetValue(test.value)
			}
			dir := t.TempDir()
			t.Setenv(ebkWorkDirEnvName, dir)
			file.Section("server").Key("static_root_path").SetValue(dir)
			file.Section("storage").Key("local_filesystem_path").SetValue(dir)
			path := filepath.Join(dir, "config.ini")
			require.NoError(t, file.SaveTo(path))
			config, err := LoadConfiguration(path)
			if test.wantFail {
				require.ErrorContains(t, err, "invalid [llm_assistant] run_timeout")
				require.Nil(t, config)
				return
			}
			require.NoError(t, err)
			wantTimeout := uint32(180000)
			if test.name == "custom" {
				wantTimeout = 240000
			}
			require.Equal(t, AIAssistantRuntimeConfig{RunTimeout: wantTimeout, QueryTimeout: 30000, MaxRounds: 12}, config.AIAssistantRuntimeConfig)
			require.Equal(t, defaultAIAssistantOpenAIModelID, config.AIAssistantLLMConfig.OpenAIModelID)
		})
	}
}
