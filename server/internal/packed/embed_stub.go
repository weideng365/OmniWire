//go:build !embed

package packed

import "io/fs"

var FS fs.FS

var Enabled = false
