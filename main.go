package main

import (
	"os"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/curtisnewbie/vfm/vfm"
)

func main() {
	common.DefaultReadConfig(os.Args)
	server.ConfigureLogging()
	vfm.PrepareServer()
	server.BootstrapServer()
}
