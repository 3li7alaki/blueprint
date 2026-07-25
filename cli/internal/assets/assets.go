package assets

import "embed"

//go:embed templates/*.md questions/*.toml
var FS embed.FS
