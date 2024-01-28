package vfm

import "github.com/curtisnewbie/miso/miso"

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
