// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	migrationDefaultLockTimeoutSecs = 3
	migrationDefaultBackoffTimeSecs = 5
	migrationDefaultNumRetries      = 3
)

type asyncMigrationStatus string

const (
	unknown  asyncMigrationStatus = ""
	run      asyncMigrationStatus = "run"      // migration should be run
	skip     asyncMigrationStatus = "skip"     // migration should be skipped (not sure if needed?)
	complete asyncMigrationStatus = "complete" // migration was already executed
	failed   asyncMigrationStatus = "failed"   // migration has failed
)

// AsyncMigration executes a single database migration that allows concurrent DML
type AsyncMigration interface {
	// name of migration, must be unique among migrations - used for saving status in database
	Name() string
	// returns if migration should be run / was already executed
	GetStatus(SqlStore) (asyncMigrationStatus, error)
	// exectutes migration, gets started transaction with lock timeouts set
	Execute(context.Context, SqlStore, *sql.Conn) (asyncMigrationStatus, error)
}

// MigrationRunner runs queued async migrations
type MigrationRunner struct {
	supplier   SqlStore
	migrations []AsyncMigration
	done       chan struct{}
	options    MigrationOptions
	cancelFn   context.CancelFunc
}

// MigrationOptions describes options for MigrationRunner
type MigrationOptions struct {
	LockTimeoutSecs int // Lock timeout in seconds
	BackoffTimeSecs int // Time to wait between retries
	NumRetries      int // Number of retries in case of failed migration
}

// NewMigrationRunner creates a migration runner for the given store
func NewMigrationRunner(s SqlStore, o MigrationOptions) *MigrationRunner {
	if o.LockTimeoutSecs == 0 {
		o.LockTimeoutSecs = migrationDefaultLockTimeoutSecs
	}
	if o.BackoffTimeSecs == 0 {
		o.BackoffTimeSecs = migrationDefaultBackoffTimeSecs
	}
	if o.NumRetries == 0 {
		o.NumRetries = migrationDefaultNumRetries
	}
	return &MigrationRunner{
		supplier: s,
		done:     make(chan struct{}),
		options:  o,
	}
}

// Add checks if the migration should be executed and adds it to queue
func (r *MigrationRunner) Add(m AsyncMigration) error {
	// check status in Systems table
	currentStatus, appErr := r.supplier.System().GetByName("migration_" + m.Name())
	if appErr == nil {
		if currentStatus.Value == string(complete) || currentStatus.Value == string(skip) {
			return nil
		}
	}
	// get status from migration
	status, err := m.GetStatus(r.supplier)
	if err != nil {
		return err
	}
	if status == complete || status == skip {
		return nil
	}
	r.migrations = append(r.migrations, m)
	return nil
}

// Run all queued migrations sequentially
func (r *MigrationRunner) Run() error {
	go func() {
		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()
		defer close(r.done)
		r.cancelFn = cancelFn

		for _, m := range r.migrations {
			// function that will try to execute migration
			migrate := func() error {
				conn, err := createConnectionWithLockTimeout(ctx, r.supplier, r.options.LockTimeoutSecs)
				if err != nil {
					return errors.Wrap(err, "failed to setup connection")
				}
				defer releaseConnection(ctx, r.supplier, conn)

				// run migration
				status, err := m.Execute(ctx, r.supplier, conn)
				if err != nil {
					return errors.Wrap(err, "failed to execute migration")
				}
				switch status {
				case complete, skip:
					// save migration status
					r.supplier.System().SaveOrUpdate(&model.System{Name: "migration_" + m.Name(), Value: string(status)})
				case failed:
					return errors.New("failed migration " + m.Name())
				default:
					// should we return error here to retry?
					return errors.New("invalid result from migration")
				}
				return nil
			}
			// retry migration if it fails
			for i := 0; i < r.options.NumRetries; i++ {
				err := migrate()
				if err == nil {
					break
				}
				mlog.Error("Migration error", mlog.Err(err))
				select {
				case <-ctx.Done():
					return
				// wait before trying again
				case <-time.After(time.Duration(r.options.BackoffTimeSecs) * time.Second):
				}
			}
		}
	}()
	return nil
}

// Wait returns after all migrations are processed (executed, skipped or failed)
func (r *MigrationRunner) Wait() {
	<-r.done
}

// WaitWithTimeout returns after all migrations are processed or if timeout passes it cancels
func (r *MigrationRunner) WaitWithTimeout(timeout time.Duration) {
	select {
	case <-r.done:
	case <-time.After(timeout):
		r.Cancel()
	}
}

// Cancel running migrations
func (r *MigrationRunner) Cancel() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}

// get an explicit connection because that guarantees a single session for all queries
func createConnectionWithLockTimeout(ctx context.Context, s SqlStore, timeout int) (*sql.Conn, error) {
	conn, err := s.GetMaster().Db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var setTimeoutSQL string
	if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		setTimeoutSQL = "SET SESSION lock_timeout = '%ds'"
	} else if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		setTimeoutSQL = "SET SESSION lock_wait_timeout = %d"
	} else {
		return nil, errors.New("Unsupported driver")
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf(setTimeoutSQL, timeout))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// releaseConnection reverts session variables to defaults and returns connection to pool
func releaseConnection(ctx context.Context, s SqlStore, conn *sql.Conn) {
	var revertTimeoutSQL string
	if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		revertTimeoutSQL = "SET SESSION lock_timeout TO DEFAULT"
	} else if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		revertTimeoutSQL = "SET SESSION lock_wait_timeout = @@GLOBAL.lock_wait_timeout"
	}
	_, err := conn.ExecContext(ctx, revertTimeoutSQL)
	if err != nil {
		// in case of error force discarding of this connection
		conn.Raw(func(interface{}) error { return driver.ErrBadConn })
	} else {
		// release connection
		conn.Close()
	}
}
