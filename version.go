package puppers

import (
	// embed required for go:embed
	_ "embed"
	"strings"
)

var (
	// Version dynamically embedded version string
	Version string = strings.TrimSpace(version)
	//go:embed version.txt
	version string
)
