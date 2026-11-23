// SPDX-License-Identifier: MIT

// the solaris constraint also applies to illumos, which the driver does support
//go:build solaris && !illumos

package processor

// sqliteSupported is false as modernc.org/sqlite does not support solaris
const sqliteSupported = false
