package vfm

import (
	"fmt"
	"net/url"
	"time"

	"github.com/curtisnewbie/miso/client"
	"github.com/curtisnewbie/miso/core"
	"github.com/curtisnewbie/miso/redis"
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
	UplTime    core.ETime  `json:"uplTime"`
	LogDelTime *core.ETime `json:"logDelTime"`
	PhyDelTime *core.ETime `json:"phyDelTime"`
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

func FindUserId(rail core.Rail, username string) (int, error) {
	r := client.NewDynTClient(rail, "/remote/user/id", "auth-service").
		EnableTracing().
		AddQueryParams("username", username).
		Get()
	if r.Err != nil {
		return 0, fmt.Errorf("failed to request auth-service, %v", r.Err)
	}
	defer r.Close()

	var res core.GnResp[int]
	if e := r.ReadJson(&res); e != nil {
		return 0, e
	}

	if res.Error {
		return 0, fmt.Errorf("failed to findUserId, code: %v, msg: %v", res.ErrorCode, res.Msg)
	}

	return res.Data, nil
}

func FindUser(rail core.Rail, req FindUserReq) (UserInfo, error) {
	r := client.NewDynTClient(rail, "/remote/user/info", "auth-service").
		EnableTracing().
		PostJson(req)
	if r.Err != nil {
		return UserInfo{}, r.Err
	}
	defer r.Close()

	var res core.GnResp[UserInfo]
	r.ReadJson(&res)
	if res.Error {
		return UserInfo{}, fmt.Errorf("failed to findUser, req: %+v, code: %v, msg: %v", req, res.ErrorCode, res.Msg)
	}
	return res.Data, nil
}

func CachedFindUser(rail core.Rail, userId int) (UserInfo, error) {
	ui, _, err := userIdInfoCache.GetElse(rail, fmt.Sprintf("vfm:user:cache:%d", userId),
		func() (UserInfo, bool, error) {
			fui, errFind := FindUser(rail, FindUserReq{
				UserId: &userId,
			})
			return fui, true, errFind
		})
	return ui, err
}

func FetchUsernames(rail core.Rail, req FetchUsernamesReq) (FetchUsernamesRes, error) {
	r := client.NewDynTClient(rail, "/remote/user/userno/username", "auth-service").
		EnableTracing().
		PostJson(&req)
	if r.Err != nil {
		return FetchUsernamesRes{}, r.Err
	}
	defer r.Close()

	var res core.GnResp[FetchUsernamesRes]
	if e := r.ReadJson(&res); e != nil {
		return FetchUsernamesRes{}, e
	}
	return res.Data, res.Err()
}

func FetchFstoreFileInfo(rail core.Rail, fileId string, uploadFileId string) (FstoreFile, error) {
	r := client.NewDynTClient(rail, "/file/info", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("uploadFileId", uploadFileId).
		Get()
	if r.Err != nil {
		return FstoreFile{}, r.Err
	}
	defer r.Close()

	var res core.GnResp[FstoreFile]
	if e := r.ReadJson(&res); e != nil {
		return FstoreFile{}, e
	}
	return res.Data, res.Err()
}

func DeleteFstoreFile(rail core.Rail, fileId string) error {
	r := client.NewDynTClient(rail, "/file", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		Delete()
	if r.Err != nil {
		return r.Err
	}
	defer r.Close()

	var res core.GnResp[any]
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

func GetFstoreTmpToken(rail core.Rail, fileId string, filename string) (string, error) {
	r := client.NewDynTClient(rail, "/file/key", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("filename", url.QueryEscape(filename)).
		Get()
	if r.Err != nil {
		return "", r.Err
	}
	defer r.Close()

	var res core.GnResp[string]
	if e := r.ReadJson(&res); e != nil {
		return "", e
	}

	if res.Error {
		return "", res.Err()
	}
	return res.Data, nil
}
