// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package migration

import (
	"sync"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/store/storetest"
	"github.com/mattermost/mattermost-server/v5/testlib"
)

func MigrationTest(t *testing.T, fn func(*testing.T, *MigrationRunner)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range storeTypes {
		t.Run(st.Name, func(t *testing.T) {
			runner := NewMigrationRunner(st.SqlSupplier)
			fn(t, runner)
		})
	}
}

type storeType struct {
	Name        string
	SqlSettings *model.SqlSettings
	SqlSupplier *sqlstore.SqlSupplier
	Store       store.Store
}

var storeTypes []*storeType

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	mainHelper = testlib.NewMainHelperWithOptions(nil)
	defer mainHelper.Close()

	initStores()

	mainHelper.Main(m)
	tearDownStores()
}

func initStores() {
	storeTypes = append(storeTypes, &storeType{
		Name:        "MySQL",
		SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_MYSQL),
	})
	storeTypes = append(storeTypes, &storeType{
		Name:        "PostgreSQL",
		SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_POSTGRES),
	})

	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	var wg sync.WaitGroup
	for _, st := range storeTypes {
		st := st
		wg.Add(1)
		go func() {
			defer wg.Done()
			st.SqlSupplier = sqlstore.NewSqlSupplier(*st.SqlSettings, nil)
			st.Store = st.SqlSupplier
			st.Store.DropAllTables()
			st.Store.MarkSystemRanUnitTests()
		}()
	}
	wg.Wait()
}

var tearDownStoresOnce sync.Once

func tearDownStores() {
	tearDownStoresOnce.Do(func() {
		var wg sync.WaitGroup
		wg.Add(len(storeTypes))
		for _, st := range storeTypes {
			st := st
			go func() {
				if st.Store != nil {
					st.Store.Close()
				}
				if st.SqlSettings != nil {
					storetest.CleanupSqlSettings(st.SqlSettings)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}
