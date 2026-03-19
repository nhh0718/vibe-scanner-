// Package main - embed web dashboard
package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var webDist embed.FS

// GetWebFS trả về filesystem chứa web dashboard
func GetWebFS() (fs.FS, error) {
	return fs.Sub(webDist, "web/dist")
}
