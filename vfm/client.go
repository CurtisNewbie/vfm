package vfm

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/curtisnewbie/miso/miso"
)

const (
	FileStatusNormal = "NORMAL" // file.status - normal
)

var (
	userIdInfoCache = miso.NewLazyORCache("vfm:user:info:userid",
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
		miso.RCacheConfig{
			Exp:    1 * time.Minute,
			NoSync: true,
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
	var r miso.GnResp[int]
	err := miso.NewDynTClient(rail, "/remote/user/id", "user-vault").
		Require2xx().
		AddQueryParams("username", username).
		Get().
		Json(&r)
	if err != nil {
		return 0, fmt.Errorf("failed to findUserId (user-vault), %v", err)
	}
	return r.Res()
}

func FindUser(rail miso.Rail, req FindUserReq) (UserInfo, error) {
	var r miso.GnResp[UserInfo]
	err := miso.NewDynTClient(rail, "/remote/user/info", "user-vault").
		Require2xx().
		PostJson(req).
		Json(&r)
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to find user (user-vault), %v", err)
	}
	return r.Res()
}

func CachedFindUser(rail miso.Rail, userId int) (UserInfo, error) {
	userIdInt := strconv.FormatInt(int64(userId), 10)
	return userIdInfoCache.Get(rail, userIdInt)
}

func FetchUsernames(rail miso.Rail, req FetchUsernamesReq) (FetchUsernamesRes, error) {
	var r miso.GnResp[FetchUsernamesRes]
	err := miso.NewDynTClient(rail, "/remote/user/userno/username", "user-vault").
		Require2xx().
		PostJson(req).
		Json(&r)
	if err != nil {
		return FetchUsernamesRes{}, fmt.Errorf("failed to fetch usernames (user-vault), %v", err)
	}
	return r.Res()
}

func FetchFstoreFileInfo(rail miso.Rail, fileId string, uploadFileId string) (FstoreFile, error) {
	var r miso.GnResp[FstoreFile]
	err := miso.NewDynTClient(rail, "/file/info", "fstore").
		Require2xx().
		AddQueryParams("fileId", fileId).
		AddQueryParams("uploadFileId", uploadFileId).
		Get().
		Json(&r)
	if err != nil {
		return FstoreFile{}, fmt.Errorf("failed to fetch mini-fstore fileInfo, %v", err)
	}
	return r.Res()
}

func DeleteFstoreFile(rail miso.Rail, fileId string) error {
	var r miso.GnResp[any]
	err := miso.NewDynTClient(rail, "/file", "fstore").
		Require2xx().
		AddQueryParams("fileId", fileId).
		Delete().
		Json(&r)
	if err != nil {
		return fmt.Errorf("failed to delete mini-fstore file, fileId: %v, %v", fileId, err)
	}

	if r.Error {
		if r.ErrorCode == "FILE_DELETED" {
			rail.Infof("file already deleted, fileId: %v", fileId)
			return nil
		}
		return r.Err()
	}
	return nil
}

func GetFstoreTmpToken(rail miso.Rail, fileId string, filename string) (string, error) {
	var r miso.GnResp[string]
	err := miso.NewDynTClient(rail, "/file/key", "fstore").
		Require2xx().
		AddQueryParams("fileId", fileId).
		AddQueryParams("filename", url.QueryEscape(filename)).
		Get().
		Json(&r)

	if err != nil {
		return "", fmt.Errorf("failed to generate mini-fstore temp token, fileId: %v, filename: %v, %v",
			fileId, filename, err)
	}
	return r.Res()
}
