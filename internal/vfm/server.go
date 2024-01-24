package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/miso/miso"
)

func PrepareServer(rail miso.Rail) error {
	common.LoadBuiltinPropagationKeys()

	if err := PrepareGoAuthReport(rail); err != nil {
		return err
	}

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

func PrepareGoAuthReport(rail miso.Rail) error {
	goauth.ReportResourcesOnBootstrapped(rail, []goauth.AddResourceReq{
		{Name: ManageFileResName, Code: ManageFileResCode},
		{Name: AdminFsResName, Code: AdminFsResCode},
	})
	goauth.ReportPathsOnBootstrapped(rail)
	return nil
}

func ScheduleJobs(rail miso.Rail) error {
	err := miso.ScheduleDistributedTask(miso.Job{
		Name:            "CalcDirSizeJob",
		Cron:            "0/30 * * * ?",
		CronWithSeconds: false,
		LogJobExec:      true,
		Run: func(r miso.Rail) error {
			return BatchCalcDirSize(r, miso.GetMySQL())
		},
	})
	if err != nil {
		return err
	}
	return nil
}