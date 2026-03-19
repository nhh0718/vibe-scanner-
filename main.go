package main

import (
	"github.com/vibescanner/vibescanner/cmd"
	"github.com/vibescanner/vibescanner/internal/output"
)

func init() {
	// Initialize the web filesystem provider
	output.GetWebFSFunc = GetWebFS
}

func main() {
	cmd.Execute()
}
