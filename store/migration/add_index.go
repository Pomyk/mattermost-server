// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package migration

import (
	"errors"

	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
)

// CreateIndex - asynchronous migration that adds an index to table
type CreateIndex struct {
	name      string
	table     string
	columns   []string
	indexType string
	unique    bool
}

func NewCreateIndex(indexName string, tableName string, columnNames []string, indexType string, unique bool) *CreateIndex {
	return &CreateIndex{
		name:      indexName,
		table:     tableName,
		columns:   columnNames,
		indexType: indexType,
		unique:    unique,
	}
}

// Name returns name of the migration, should be unique
func (m *CreateIndex) Name() string {
	return "add_index_" + m.name
}

// GetStatus returns if the migration should be executed or not
func (m *CreateIndex) GetStatus(ss SqlStore) (AsyncMigrationStatus, error) {
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		_, errExists := ss.GetMaster().SelectStr("SELECT $1::regclass", m.name)
		// It should fail if the index does not exist
		if errExists == nil {
			return AsyncMigrationStatusSkip, nil
		}
		if m.indexType == sqlstore.INDEX_TYPE_FULL_TEXT && len(m.columns) != 1 {
			return AsyncMigrationStatusFailed, errors.New("Unable to create multi column full text index")
		}
	} else if ss.DriverName() == model.DATABASE_DRIVER_MYSQL {
		if m.indexType == sqlstore.INDEX_TYPE_FULL_TEXT {
			return AsyncMigrationStatusFailed, errors.New("Unable to create full text index concurrently")
		}
		count, err := ss.GetMaster().SelectInt("SELECT COUNT(0) AS index_exists FROM information_schema.statistics WHERE TABLE_SCHEMA = DATABASE() and table_name = ? AND index_name = ?", m.table, m.name)
		if err != nil {
			return AsyncMigrationStatusUnknown, err
		}
		if count > 0 {
			return AsyncMigrationStatusSkip, nil
		}
	}
	return AsyncMigrationStatusRun, nil
}

// Execute runs the migration
func (m *CreateIndex) Execute(ss SqlStore, tx *gorp.Transaction) (AsyncMigrationStatus, error) {
	// TODO
	return AsyncMigrationStatusComplete, nil
}
