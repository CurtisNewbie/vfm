package vfm

import (
	"fmt"
	"time"

	hammer "github.com/curtisnewbie/hammer/api"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

type FileCompressInfo struct {
	Id           int
	Name         string
	Uuid         string
	FstoreFileId string
}

func CompensateImageCompression(rail miso.Rail, tx *gorm.DB) error {
	rail.Info("CompensateImageCompression start")
	defer miso.TimeOp(rail, time.Now(), "CompensateImageCompression")

	limit := 500
	minId := 0

	for {
		var files []FileCompressInfo
		t := tx.
			Raw(`SELECT id, name, uuid, fstore_file_id
			FROM file_info
			WHERE id > ?
			AND file_type = 'file'
			AND is_logic_deleted = 0
			AND thumbnail = ''
			ORDER BY id ASC
			LIMIT ?`, minId, limit).
			Scan(&files)
		if t.Error != nil {
			return t.Error
		}
		if t.RowsAffected < 1 || len(files) < 1 {
			return nil // the end
		}

		for _, f := range files {
			if !isImage(f.Name) {
				continue
			}
			event := hammer.ImageCompressTriggerEvent{Identifier: f.Uuid, FileId: f.FstoreFileId, ReplyTo: VfmCompressImgNotifyEventBus}
			if e := miso.PubEventBus(rail, event, hammer.CompressImageTriggerEventBus); e != nil {
				rail.Errorf("Failed to send CompressImageEvent, minId: %v, uuid: %v, %v", minId, f.Uuid, e)
				return e
			}
		}

		minId = files[len(files)-1].Id
		rail.Infof("CompensateImageCompression, minId: %v", minId)
	}
}

type FileUplNoInf struct {
	Id         int
	UploaderId int
}

func CompensateFileUploaderNo(rail miso.Rail, tx *gorm.DB) error {
	rail.Info("CompensateFileUploaderNo start")
	defer miso.TimeOp(rail, time.Now(), "CompensateFileUploaderNo")

	limit := 500
	minId := 0

	for {
		var files []FileUplNoInf
		t := tx.
			Raw(`SELECT id, uploader_id
			FROM file_info
			WHERE id > ?
			AND uploader_no = ''
			AND uploader_id > 0
			ORDER BY id ASC
			LIMIT ?`, minId, limit).
			Scan(&files)
		if t.Error != nil {
			return t.Error
		}
		if t.RowsAffected < 1 || len(files) < 1 {
			return nil // the end
		}

		for i := range files {
			f := files[i]

			if err := UpdateUploaderNo(rail, tx, f); err != nil {
				rail.Errorf("failed to UpdateUploaderNo, f.id: %v, f.uploader_id: %v, %v", f.Id, f.UploaderId, err)
				return err
			}
		}

		minId = files[len(files)-1].Id
		rail.Infof("CompensateFileUploaderNo, minId: %v", minId)
	}
}

func UpdateUploaderNo(rail miso.Rail, tx *gorm.DB, f FileUplNoInf) error {
	u, err := CachedFindUser(rail, f.UploaderId)
	if err != nil {
		return fmt.Errorf("failed to find user, %v", err)
	}
	if u.UserNo == "" {
		rail.Warnf("User doesn't have userNo, uploader_id: %v", f.UploaderId)
		return nil
	}

	return tx.Exec(`UPDATE file_info SET uploader_no = ? WHERE id = ?`, u.UserNo, f.Id).Error
}
