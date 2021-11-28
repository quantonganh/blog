package ui

import (
	"embed"
)

//go:embed html/*.html
var HTMLFS embed.FS

//go:embed static
var StaticFS embed.FS
