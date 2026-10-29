// SPDX-License-Identifier: MIT

//go:build !solaris

package processor

import (
	_ "modernc.org/sqlite"
)

// sqliteSupported is false on platforms the SQLite driver does not support
const sqliteSupported = true
