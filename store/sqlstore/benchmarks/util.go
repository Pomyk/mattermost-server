package benchmarks

import (
	"os"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store/sqlstore"
	"github.com/mattermost/mattermost-server/store/storetest"
)

func setup(driver string) *sqlstore.SqlSupplier {
	var settings *model.SqlSettings
	switch driver {
	case "mysql":
		settings = storetest.MySQLBenchmarkSettings()
	case "postgres":
		settings = storetest.PostgreSQLBenchmarkSettings()
	}
	supplier := sqlstore.NewSqlSupplier(*settings, nil)
	return supplier
}

func getEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return defaultValue
	}
}
