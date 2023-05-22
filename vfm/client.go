package vfm

import (
	"fmt"

	"github.com/curtisnewbie/gocommon/client"
	"github.com/curtisnewbie/gocommon/common"
)

type FindUserReq struct {
	UserId   *int    `json:"userId"`
	UserNo   *string `json:"userNo"`
	Username *string `json:"username"`
}

type UserInfo struct {
	Id       int
	Username string
	UserNo   string
}

type FetchUsernamesReq struct {
	UserNos []string `json:"userNos"`
}

type FetchUsernamesRes struct {
	UserNoToUsername map[string]string `json:"userNoToUsername"`
}

func FindUserId(c common.ExecContext, username string) (int, error) {
	r := client.NewDynTClient(c, "/remote/user/id", "auth-service").
		EnableTracing().
		Get(map[string][]string{"username": {username}})
	if r.Err != nil {
		return 0, fmt.Errorf("failed to request auth-service, %v", r.Err)
	}
	defer r.Close()

	var res common.GnResp[int]
	if e := r.ReadJson(&res); e != nil {
		return 0, e
	}

	if res.Error {
		return 0, fmt.Errorf("failed to findUserId, code: %v, msg: %v", res.ErrorCode, res.Msg)
	}

	return res.Data, nil
}

func FindUser(c common.ExecContext, req FindUserReq) (UserInfo, error) {
	r := client.NewDynTClient(c, "/remote/user/info", "auth-service").
		EnableTracing().
		PostJson(req)
	if r.Err != nil {
		return UserInfo{}, r.Err
	}
	defer r.Close()

	var res common.GnResp[UserInfo]
	r.ReadJson(&res)
	if res.Error {
		return UserInfo{}, fmt.Errorf("failed to findUser, req: %+v, code: %v, msg: %v", req, res.ErrorCode, res.Msg)
	}
	return res.Data, nil
}

func FetchUsernames(c common.ExecContext, req FetchUsernamesReq) (FetchUsernamesRes, error) {
	r := client.NewDynTClient(c, "/remote/user/userno/username", "auth-service").
		EnableTracing().
		PostJson(&req)
	if r.Err != nil {
		return FetchUsernamesRes{}, r.Err
	}
	defer r.Close()

	var res common.GnResp[FetchUsernamesRes]
	if e := r.ReadJson(&res); e != nil {
		return FetchUsernamesRes{}, fmt.Errorf("failed to unmarshel to FetchUsernamesRes, %v", e)
	}
	return res.Data, nil
}
