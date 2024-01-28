package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
)

func PrepareServer(rail miso.Rail) error {
	common.LoadBuiltinPropagationKeys()

	if err := PrepareEventBus(rail); err != nil {
		return err
	}

	if err := RegisterHttpRoutes(rail); err != nil {
		return err
	}

	if err := ScheduleJobs(rail); err != nil {
		return err
	}
	return nil
}
