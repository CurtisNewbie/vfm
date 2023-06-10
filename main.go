package main

import (
	"os"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/curtisnewbie/vfm/vfm"
)

func main() {
	c := common.EmptyExecContext()
	server.DefaultBootstrapServer(os.Args, c, func() error {
		vfm.PrepareServer(c)
		return nil
	})
}
