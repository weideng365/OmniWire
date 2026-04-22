//go:build embed

package packed

import "embed"

//go:embed all:public
var FS embed.FS

var Enabled = true
