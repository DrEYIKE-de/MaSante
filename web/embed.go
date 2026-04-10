// Package web embeds the frontend static files into the binary.
package web

import "embed"

//go:embed index.html
var Files embed.FS
