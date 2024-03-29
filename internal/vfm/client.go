package vfm

import (
	"fmt"
	"net/url"
	"time"

	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
)

const (
	FileStatusNormal = "NORMAL" // file.status - normal
)

var (
	userIdInfoCache = miso.NewRCache[vault.UserInfo]("vfm:user:info:userno", miso.RCacheConfig{Exp: 5 * time.Minute, NoSync: true})
)

func CachedFindUser(rail miso.Rail, userNo string) (vault.UserInfo, error) {
	return userIdInfoCache.Get(rail, userNo, func() (vault.UserInfo, error) {
		fui, errFind := vault.FindUser(rail, vault.FindUserReq{
			UserNo: &userNo,
		})
		return fui, errFind
	})
}

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
