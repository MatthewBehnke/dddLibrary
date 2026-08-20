// Package migrations embeds the SQL schema migrations so they travel inside the
// binary and can be applied programmatically at startup or from tests, without
// depending on files being present on disk.
package migrations

import "embed"

// FS holds the versioned golang-migrate .sql files.
//
//go:embed *.sql
var FS embed.FS
