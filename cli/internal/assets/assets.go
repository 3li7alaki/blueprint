package assets

import "embed"

//go:embed templates/*.md questions/*.toml vague.txt
var FS embed.FS
