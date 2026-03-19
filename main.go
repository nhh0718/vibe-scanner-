package main

import (
	"github.com/nhh0718/vibe-scanner-/cmd"
	"github.com/nhh0718/vibe-scanner-/internal/output"
)

func init() {
	// Initialize the web filesystem provider
	output.GetWebFSFunc = GetWebFS
}

func main() {
	cmd.Execute()
}
