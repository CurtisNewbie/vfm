package main

import (
	"os"

	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/vfm/internal/vfm"
)

func main() {
	miso.PreServerBootstrap(vfm.PrepareServer)
	miso.BootstrapServer(os.Args)
}
