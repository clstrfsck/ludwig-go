package highlight

import "embed"

//go:embed syntax/*.yaml
var SyntaxFiles embed.FS
