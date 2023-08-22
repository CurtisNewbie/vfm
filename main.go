package main

import (
	"os"

	"github.com/curtisnewbie/gocommon/server"
	"github.com/curtisnewbie/vfm/vfm"
)

func main() {
	server.PreServerBootstrap(vfm.PrepareServer)
	server.BootstrapServer(os.Args)
}
