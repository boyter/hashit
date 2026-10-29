// Copyright 2025 The Sqlite Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite // import "modernc.org/sqlite"

import (
	"context"
	"database/sql/driver"
	"sync/atomic"

	sqlite3 "modernc.org/sqlite/lib"
)

type tx struct {
	c *conn
}

func newTx(ctx context.Context, c *conn, opts driver.TxOptions) (*tx, error) {
	r := &tx{c: c}

	sql := "begin"
	if !opts.ReadOnly && c.beginMode != "" {
		sql = "begin " + c.beginMode
	}

	if err := r.exec(ctx, sql); err != nil {
		return nil, err
	}

	return r, nil
}

// Commit implements driver.Tx.
func (t *tx) Commit() (err error) {
	err = t.exec(context.Background(), "commit")
	if err == nil {
		return nil
	}

	// If Commit fails (e.g., SQLITE_BUSY), the connection might still be inside
	// the transaction. database/sql expects the connection to be clean (no active
	// transaction) after Commit returns.
	//
	// We check the low-level autocommit state.  0 means autocommit is disabled =>
	// we are inside a transaction.
	if t.c.db != 0 && sqlite3.Xsqlite3_get_autocommit(t.c.tls, t.c.db) == 0 {
		// Force a rollback to clean up the connection.  We ignore the rollback error
		// because we must return the original Commit error to the user.
		t.exec(context.Background(), "rollback")
	}

	return err
}

// Rollback implements driver.Tx.
func (t *tx) Rollback() (err error) {
	return t.exec(context.Background(), "rollback")
}

// exec runs sql, a BEGIN, COMMIT or ROLLBACK, through the statement path, so
// that with a shared cache it waits for the lock held by another connection,
// as an executed BEGIN does, rather than failing at once with
// SQLITE_LOCKED_SHAREDCACHE, and ends that wait once ctx is done. Unlike
// conn.exec, it reports a statement that has run as a success even if ctx
// is done by then: a BEGIN that took effect must reach database/sql as a
// transaction, which it then rolls back for the done context.
func (t *tx) exec(ctx context.Context, sql string) (err error) {
	s, err := t.c.newStmtContext(ctx, sql)
	if err != nil {
		return err
	}

	defer func() {
		if err2 := s.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	if s.pstmt == 0 { // Not a single statement, which BEGIN, COMMIT and ROLLBACK are
		_, err = s.exec(ctx, nil)
		return err
	}

	if ctx != nil && ctx.Done() != nil {
		var done int32
		defer interruptOnDone(ctx, t.c, &done)()
		defer func() {
			if err != nil && (atomic.LoadInt32(&done) != 0 || t.c.canceled.Load() != 0) {
				err = ctx.Err()
			}
		}()
	}

	_, err = t.c.step(s.pstmt)
	return err
}
