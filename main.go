package main

import (
	"os"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/curtisnewbie/vfm/vfm"
)

func main() {
	server.PreServerBootstrap(func(c common.Rail) error {
		vfm.PrepareServer(c)
		return nil
	})
	server.BootstrapServer(os.Args)
}
