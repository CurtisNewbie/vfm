package vfm

import (
	"fmt"

	"github.com/curtisnewbie/gocommon/client"
	"github.com/curtisnewbie/gocommon/common"
)

type FindUserIdRes struct {
	common.Resp
	Data int `json:"data"`
}

func FindUserId(c common.ExecContext, username string) (int, error) {
	r := client.NewDynTClient(c, "/remote/user/id", "auth-service").
		EnableTracing().
		EnableRequestLog().
		Get(map[string][]string{"username": {username}})
	if r.Err != nil {
		return 0, fmt.Errorf("failed to request auth-service, %v", r.Err)
	}

	var res FindUserIdRes
	if e := r.ReadJson(&res); e != nil {
		return 0, e
	}

	return res.Data, nil
}
