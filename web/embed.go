// Package web embeds the frontend static files into the binary.
package web

import "embed"

//go:embed index.html static
var Files embed.FS
