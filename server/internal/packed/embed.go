//go:build embed

package packed

import "embed"

//go:embed public
var FS embed.FS

var Enabled = true
