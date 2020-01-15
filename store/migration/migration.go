// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package migration

import (
	"errors"
	"time"

	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
)

type AsyncMigrationStatus string

const (
	AsyncMigrationStatusUnknown  = ""
	AsyncMigrationStatusRun      = "run"      // migration should be run
	AsyncMigrationStatusSkip     = "skip"     // migration should be skipped (not sure if needed?)
	AsyncMigrationStatusComplete = "complete" // migration was already executed
	AsyncMigrationStatusFailed   = "failed"   // migration has failed
)

// AsyncMigration executes a single database migration that allows concurrent DML
type AsyncMigration interface {
	// name of migration, must be unique among migrations - used for saving status in database
	Name() string
	// returns if migration should be run / was already executed
	GetStatus(*sqlstore.SqlSupplier) (AsyncMigrationStatus, error)
	// exectutes migration, gets started transaction with lock timeouts set
	Execute(*sqlstore.SqlSupplier, *gorp.Transaction) (AsyncMigrationStatus, error)
}

// MigrationRunner runs queued async migrations
type MigrationRunner struct {
	supplier   *sqlstore.SqlSupplier
	migrations []AsyncMigration
}

func NewMigrationRunner(s *sqlStore.Supplier) {
	return &MigrationRunner{
		supplier: s,
	}
}

// Add checks if the migration should be executed and adds it to queue
func (r *MigrationRunner) Add(m AsyncMigration) error {
	// check status in Systems table
	currentStatus, appErr := r.supplier.System().GetByName("migration_" + m.Name())
	if appErr != nil {
		return appErr
	}
	if currentStatus.Value == AsyncMigrationStatusComplete || currentStatus.Value == AsyncMigrationStatusSkip {
		return nil
	}
	// get status from migration
	status, err := m.GetStatus(r.supplier)
	if err != nil {
		return err
	}
	if status == AsyncMigrationStatusComplete || status.Value == AsyncMigrationStatusSkip {
		return nil
	}
	r.migrations = append(r.migrations, m)
	return nil
}

// Run all queued migrations sequentially
func (r *MigrationRunner) Run() error {
	go func() {
		for _, m := range r.migrations {
			// function that will try to execute migration
			migrate := func() error {
				var err error
				tx, err := createTransactionWithLockTimeouts(r.supplier)
				if err != nil {
					mlog.Error("Failed to setup transaction", mlog.Err(err))
					return err
				}
				defer finishTransaction(r.supplier, tx, &err)

				status, err := m.Execute(r.supplier, tx)
				if err != nil {
					mlog.Error("Failed to execute migration", mlog.Err(err))
					return err
				}
				if status == AsyncMigrationStatusComplete || status == AsyncMigrationStatusSkip {
					r.supplier.System().SaveOrUpdate(&model.System{Name: "migration_" + m.Name(), Value: string(status)})
				} else if status == AsyncMigrationStatusFailed {
					mlog.Error("Failed migration: " + m.Name())
					return errors.New("Failed migration")
				} else {
					// should we return error here to retry?
					return errors.New("Unknown ")
				}
				return nil
			}
			// retry migration if it fails
			for i := 0; i < 3; i++ {
				err := migrate()
				if err == nil {
					break
				}
				// wait before trying again
				time.Sleep(3 * time.Second)
			}
		}
	}()
	return nil
}

// use a transaction because that guarantees a single session for all queries
func createTransactionWithLockTimeouts(s *sqlstore.SqlSupplier) (*gorp.Transaction, error) {
	tx, err := s.GetMaster().Begin()
	if err != nil {
		return nil, err
	}
	if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		// in postgresql set local is limited to single transacion
		tx.Exec("SET LOCAL lock_timeout = '3s'")
	} else if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		// set timeout for session, we have to revert it later
		tx.Exec("SET SESSION lock_wait_timeout = 3")
	}
	return tx, nil
}

// finishTransaction reverts session variables and commits or rollbacks transaction
func finishTransaction(s *sqlstore.SqlSupplier, tx *gorp.Transaction, err *error) {
	if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		// revert lock timeout to global value
		tx.Exec("SET SESSION lock_wait_timeout = @@GLOBAL.lock_wait_timeout")
	}
	if *err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
}
