package vfm

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/curtisnewbie/miso/miso"
)

const (
	FS_STATUS_NORMAL = "NORMAL" // file.status - normal
)

var (
	userIdInfoCache = miso.NewLazyORCache("vfm:user:info:userid", 1*time.Minute,
		func(rail miso.Rail, key string) (UserInfo, error) {
			userId, err := strconv.Atoi(key)
			if err != nil {
				return UserInfo{}, miso.NewErr("Invalid userId format, %v", err)
			}
			fui, errFind := FindUser(rail, FindUserReq{
				UserId: &userId,
			})
			return fui, errFind
		},
	)
)

type FstoreFile struct {
	Id         int64       `json:"id"`
	FileId     string      `json:"fileId"`
	Name       string      `json:"name"`
	Status     string      `json:"status"`
	Size       int64       `json:"size"`
	Md5        string      `json:"md5"`
	UplTime    miso.ETime  `json:"uplTime"`
	LogDelTime *miso.ETime `json:"logDelTime"`
	PhyDelTime *miso.ETime `json:"phyDelTime"`
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

func FindUserId(rail miso.Rail, username string) (int, error) {
	type respType = miso.GnResp[int]
	r, err := miso.NewDynTClient(rail, "/remote/user/id", "user-vault").
		EnableTracing().
		AddQueryParams("username", username).
		Get().
		Json(respType{})

	if err != nil {
		return 0, fmt.Errorf("failed to findUserId (user-vault), %v", err)
	}

	res := r.(respType)
	if res.Error {
		return 0, fmt.Errorf("failed to findUserId, code: %v, msg: %v", res.ErrorCode, res.Msg)
	}

	return res.Data, nil
}

func FindUser(rail miso.Rail, req FindUserReq) (UserInfo, error) {
	type respType = miso.GnResp[UserInfo]
	r, err := miso.NewDynTClient(rail, "/remote/user/info", "user-vault").
		EnableTracing().
		PostJson(req).
		Json(respType{})

	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to find user (user-vault), %v", err)
	}

	res := r.(respType)
	if res.Error {
		return UserInfo{}, fmt.Errorf("failed to findUser, req: %+v, code: %v, msg: %v", req, res.ErrorCode, res.Msg)
	}
	return res.Data, nil
}

func CachedFindUser(rail miso.Rail, userId int) (UserInfo, error) {
	return userIdInfoCache.Get(rail, strconv.FormatInt(int64(userId), 10))
}

func FetchUsernames(rail miso.Rail, req FetchUsernamesReq) (FetchUsernamesRes, error) {
	type respType = miso.GnResp[FetchUsernamesRes]
	r, err := miso.NewDynTClient(rail, "/remote/user/userno/username", "user-vault").
		EnableTracing().
		PostJson(req).
		Json(respType{})

	if err != nil {
		return FetchUsernamesRes{}, fmt.Errorf("failed to fetch usernames (user-vault), %v", err)
	}

	res := r.(respType)
	return res.Data, res.Err()
}

func FetchFstoreFileInfo(rail miso.Rail, fileId string, uploadFileId string) (FstoreFile, error) {
	type respType = miso.GnResp[FstoreFile]
	r, err := miso.NewDynTClient(rail, "/file/info", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("uploadFileId", uploadFileId).
		Get().
		Json(respType{})

	if err != nil {
		return FstoreFile{}, fmt.Errorf("failed to fetch mini-fstore fileInfo, %v", err)
	}

	res := r.(respType)
	return res.Data, res.Err()
}

func DeleteFstoreFile(rail miso.Rail, fileId string) error {
	type respType = miso.GnResp[any]
	r, err := miso.NewDynTClient(rail, "/file", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		Delete().
		Json(respType{})

	if err != nil {
		return err
	}

	res := r.(miso.GnResp[any])
	if res.Error {
		if res.ErrorCode == "FILE_DELETED" {
			rail.Infof("file already deleted, fileId: %v", fileId)
			return nil
		}
		return res.Err()
	}
	return nil
}

func GetFstoreTmpToken(rail miso.Rail, fileId string, filename string) (string, error) {
	type respType = miso.GnResp[string]
	r, err := miso.NewDynTClient(rail, "/file/key", "fstore").
		EnableTracing().
		AddQueryParams("fileId", fileId).
		AddQueryParams("filename", url.QueryEscape(filename)).
		Get().
		Json(respType{})

	if err != nil {
		return "", fmt.Errorf("failed to generate mini-fstore temp token, fileId: %v, filename: %v, %v",
			fileId, filename, err)
	}
	res := r.(respType)

	if res.Error {
		return "", res.Err()
	}
	return res.Data, nil
}
