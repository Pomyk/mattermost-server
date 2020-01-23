// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localcachelayer

import (
	"sync"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/store/storetest"
)

type storeType struct {
	Name        string
	SqlSettings *model.SqlSettings
	SqlSupplier *sqlstore.SqlSupplier
	Store       store.Store
}

var storeTypes []*storeType
var benchmarkStoreTypes []*storeType

func StoreTest(t *testing.T, f func(*testing.T, store.Store)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range storeTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) { f(t, st.Store) })
	}
}

func StoreTestWithSqlSupplier(t *testing.T, f func(*testing.T, store.Store, storetest.SqlSupplier)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range storeTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) { f(t, st.Store, st.SqlSupplier) })
	}
}

func StoreBenchmark(b *testing.B, f func(*testing.B, store.Store)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range benchmarkStoreTypes {
		st := st
		b.Run(st.Name, func(b *testing.B) { f(b, st.Store) })
	}
}

func initStores() {
	storeTypes = append(storeTypes, &storeType{
		Name:        "LocalCache+MySQL",
		SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_MYSQL),
	})
	storeTypes = append(storeTypes, &storeType{
		Name:        "LocalCache+PostgreSQL",
		SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_POSTGRES),
	})

	benchmarkStoreTypes = append(benchmarkStoreTypes, &storeType{
		Name:        "LocalCache+MySQL",
		SqlSettings: storetest.MySQLBenchmarkSettings(),
	})
	benchmarkStoreTypes = append(benchmarkStoreTypes, &storeType{
		Name:        "LocalCache+PostgreSQL",
		SqlSettings: storetest.PostgreSQLBenchmarkSettings(),
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
			st.Store = NewLocalCacheLayer(st.SqlSupplier, nil, nil, getMockCacheProvider())
			st.Store.DropAllTables()
			st.Store.MarkSystemRanUnitTests()
		}()
	}
	for _, st := range benchmarkStoreTypes {
		st.SqlSupplier = sqlstore.NewSqlSupplier(*st.SqlSettings, nil)
		st.Store = NewLocalCacheLayer(st.SqlSupplier, nil, nil, getMockCacheProvider())
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
