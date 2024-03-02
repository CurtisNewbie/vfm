package vfm

import (
	"embed"
	"os"

	"github.com/curtisnewbie/miso/middleware/svc/migrate"
	"github.com/curtisnewbie/miso/middleware/user-vault/common"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/vfm/internal/schema"
)

var (
	SchemaFs embed.FS
)

func PrepareServer() {
	common.LoadBuiltinPropagationKeys()
	miso.PreServerBootstrap(func(rail miso.Rail) error {
		rail.Infof("vfm version: %v", Version)
		return nil
	})
	// starting from v0.1.18, let svc manages the schema migration
	migrate.EnableSchemaMigrateOnProd(schema.SchemaFs, schema.BaseDir, schema.StartingVer)
	miso.PreServerBootstrap(PrepareEventBus)
	miso.PreServerBootstrap(RegisterHttpRoutes)
}

func BootstrapServer(args []string) {
	PrepareServer()
	miso.BootstrapServer(os.Args)
}
