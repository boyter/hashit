// SPDX-License-Identifier: MIT

package assets

import _ "embed"

//go:embed db/migrations.sql
var Migrations string
