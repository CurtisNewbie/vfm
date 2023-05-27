package main

import (
	"os"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/curtisnewbie/vfm/vfm"
)

func main() {
	server.DefaultBootstrapServer(os.Args, common.EmptyExecContext(), func() error {
		vfm.PrepareServer()
		return nil
	})
}
