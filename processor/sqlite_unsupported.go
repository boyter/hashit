// SPDX-License-Identifier: MIT

// the solaris constraint also applies to illumos
//go:build solaris

package processor

// sqliteSupported is false as modernc.org/sqlite does not support illumos or solaris, see
// https://gitlab.com/cznic/sqlite/-/work_items/210
const sqliteSupported = false
