package cmd

import (
	"path/filepath"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

func TestDatabaseUpdateRetainsLegacyEmbeddingSchemaAndRows(t *testing.T) {
	previous := *datastore.Container
	require.NoError(t, datastore.InitializeDataStore(&settings.Config{DatabaseConfig: &settings.DatabaseConfig{DatabaseType: settings.Sqlite3DbType, DatabasePath: filepath.Join(t.TempDir(), "schema-retention.db"), MaxOpenConnection: 1, MaxIdleConnection: 1}}))
	session := datastore.Container.UserDataStore.Choose(7).NewSession(core.NewNullContext())
	engine := session.Engine()
	session.Close()
	t.Cleanup(func() { engine.Close(); *datastore.Container = previous })
	// Exercise the actual startup/database-update registration, not a test-only
	// reconstruction of its model list.
	require.NoError(t, updateAllDatabaseTablesStructure(nil))
	exists, err := engine.IsTableExist(&models.AIAssistantEmbedding{})
	require.NoError(t, err)
	require.True(t, exists)
	legacy := &models.AIAssistantEmbedding{Uid: 7, TransactionId: 999, EmbeddingModel: "legacy", ContentHash: "rollback-hash", VectorData: "[0.1,0.2]", CreatedUnixTime: 123, UpdatedUnixTime: 456}
	_, err = engine.Insert(legacy)
	require.NoError(t, err)
	require.NoError(t, updateAllDatabaseTablesStructure(nil))
	var rows []models.AIAssistantEmbedding
	require.NoError(t, engine.Find(&rows))
	require.Equal(t, []models.AIAssistantEmbedding{*legacy}, rows)
}
