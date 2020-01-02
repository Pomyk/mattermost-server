// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storetest

import "github.com/mattermost/mattermost-server/v5/model"

// MySQLBenchmarkSettings returns the database settings to connect to the MySQL benchmark database.
func MySQLBenchmarkSettings() *model.SqlSettings {
	dsn := getEnv("BENCHMARK_DATABASE_MYSQL_DSN", defaultMysqlDSN)
	return databaseSettings("mysql", dsn)
}

// PostgreSQLBenchmarkSettings returns the database settings to connect to the PostgreSQL benchmark database.
func PostgreSQLBenchmarkSettings() *model.SqlSettings {
	dsn := getEnv("BENCHMARK_DATABASE_POSTGRESQL_DSN", defaultPostgresqlDSN)
	return databaseSettings("postgres", dsn)
}
