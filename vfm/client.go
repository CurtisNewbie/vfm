package vfm

import (
	"fmt"
	"net/url"
	"time"

	"github.com/curtisnewbie/gocommon/client"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/redis"
)

const (
	FS_STATUS_NORMAL = "NORMAL" // file.status - normal
)

var (
	userIdInfoCache = redis.NewLazyObjectRCache[UserInfo](5 * time.Minute)
)

type FstoreFile struct {
	Id         int64         `json:"id"`
	FileId     string        `json:"fileId"`
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	Size       int64         `json:"size"`
	Md5        string        `json:"md5"`
	UplTime    common.ETime  `json:"uplTime"`
	LogDelTime *common.ETime `json:"logDelTime"`
	PhyDelTime *common.ETime `json:"phyDelTime"`
}

func (f FstoreFile) IsZero() bool {
	return f.Id < 1
}

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

func FindUserId(c common.Rail, username string) (int, error) {
	r := client.NewDynTClient(c, "/remote/user/id", "auth-service").
		EnableTracing().
		AddQueryParams("username", username).
		Get()
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

func FindUser(c common.Rail, req FindUserReq) (UserInfo, error) {
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

func CachedFindUser(c common.Rail, userId int) (UserInfo, error) {
	ui, _, err := userIdInfoCache.GetElse(c, fmt.Sprintf("vfm:user:cache:%d", userId),
		func() (UserInfo, bool, error) {
			fui, errFind := FindUser(c, FindUserReq{
				UserId: &userId,
			})
			return fui, true, errFind
		})
	return ui, err
}

func FetchUsernames(c common.Rail, req FetchUsernamesReq) (FetchUsernamesRes, error) {
	r := client.NewDynTClient(c, "/remote/user/userno/username", "auth-service").
		EnableTracing().
		PostJson(&req)
	if r.Err != nil {
		return FetchUsernamesRes{}, r.Err
	}
	defer r.Close()

	var res common.GnResp[FetchUsernamesRes]
	if e := r.ReadJson(&res); e != nil {
		return FetchUsernamesRes{}, e
	}
	return res.Data, res.Err()
}

func FetchFstoreFileInfo(c common.Rail, fileId string, uploadFileId string) (FstoreFile, error) {
	r := client.NewDynTClient(c, "/file/info", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("uploadFileId", uploadFileId).
		Get()
	if r.Err != nil {
		return FstoreFile{}, r.Err
	}
	defer r.Close()

	var res common.GnResp[FstoreFile]
	if e := r.ReadJson(&res); e != nil {
		return FstoreFile{}, e
	}
	return res.Data, res.Err()
}

func DeleteFstoreFile(c common.Rail, fileId string) error {
	r := client.NewDynTClient(c, "/file", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		Delete()
	if r.Err != nil {
		return r.Err
	}
	defer r.Close()

	var res common.GnResp[any]
	if e := r.ReadJson(&res); e != nil {
		return e
	}
	if res.Error {
		if res.ErrorCode == "FILE_DELETED" {
			return nil
		}
		return res.Err()
	}
	return nil
}

func GetFstoreTmpToken(c common.Rail, fileId string, filename string) (string, error) {
	r := client.NewDynTClient(c, "/file/key", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("filename", url.QueryEscape(filename)).
		Get()
	if r.Err != nil {
		return "", r.Err
	}
	defer r.Close()

	var res common.GnResp[string]
	if e := r.ReadJson(&res); e != nil {
		return "", e
	}

	if res.Error {
		return "", res.Err()
	}
	return res.Data, nil
}
