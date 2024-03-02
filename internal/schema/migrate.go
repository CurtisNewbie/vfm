package schema

import (
	"embed"
)

//go:embed scripts/*.sql
var SchemaFs embed.FS

const (
	BaseDir     = "scripts"
	StartingVer = "v0.1.17.sql"
)
