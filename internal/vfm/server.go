package vfm

import (
	"github.com/curtisnewbie/miso/middleware/user-vault/common"
	"github.com/curtisnewbie/miso/miso"
)

func PrepareServer(rail miso.Rail) error {
	common.LoadBuiltinPropagationKeys()

	rail.Infof("vfm version: %v", Version)

	if err := PrepareEventBus(rail); err != nil {
		return err
	}

	if err := RegisterHttpRoutes(rail); err != nil {
		return err
	}
	return nil
}
