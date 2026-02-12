package assets

import (
	"embed"
)

//go:embed all:static
var StaticAssets embed.FS
