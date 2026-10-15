//go:build darwin || linux

package settings

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

// A FIFO with no writer blocks even on opening for reading. Running the real
// parser in a bounded subprocess detects secret-file reads without hanging the
// test suite or relying on permission differences between root and other users.
func TestRetiredEmbeddingSecretFilesNeverOpened(t *testing.T) {
	if configPath := os.Getenv("EBK_TEST_RETIRED_EMBEDDING_CONFIG"); configPath != "" {
		_, err := LoadConfiguration(configPath)
		require.NoError(t, err)
		return
	}
	for _, section := range []string{"llm_assistant_embedding", "llm_assistant"} {
		t.Run(section, func(t *testing.T) {
			dir := t.TempDir()
			fifo := filepath.Join(dir, "retired-secret-fifo")
			require.NoError(t, syscall.Mkfifo(fifo, 0600))
			file, err := ini.Load("../../conf/ezbookkeeping.ini")
			require.NoError(t, err)
			file.Section("server").Key("static_root_path").SetValue(dir)
			file.Section("storage").Key("local_filesystem_path").SetValue(dir)
			file.Section("llm_assistant").Key("llm_provider").SetValue("openai")
			file.Section("llm_assistant_embedding").Key("llm_provider").SetValue("openai")
			file.Section("llm_assistant_embedding").Key("openai_embedding_model_id").SetValue("legacy")
			configPath := filepath.Join(dir, "config.ini")
			require.NoError(t, file.SaveTo(configPath))
			key := "openai_embedding_model_id"
			if section == "llm_assistant_embedding" {
				key = "openai_api_key"
			}
			t.Setenv(getConfigItemFilePathEnvironmentKey(section, key), fifo)
			t.Setenv(ebkWorkDirEnvName, dir)
			t.Setenv("EBK_TEST_RETIRED_EMBEDDING_CONFIG", configPath)
			binary, err := os.Executable()
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, binary, "-test.run=^TestRetiredEmbeddingSecretFilesNeverOpened$").CombinedOutput()
			require.NoError(t, ctx.Err(), "parser opened the retired secret FIFO: %s", output)
			require.NoError(t, err, "%s", output)
		})
	}
}
