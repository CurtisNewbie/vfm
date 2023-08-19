package vfm

import (
	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
)

const (
	comprImgProcEventBus               = "hammer.image.compress.processing"
	comprImgNotifyEventBus             = "hammer.image.compress.notification"
	fileSavedEventBus                  = "vfm.file.saved"
	thumbnailUpdatedEventBus           = "vfm.file.thumbnail.updated"
	fileLDeletedEventBus               = "vfm.file.logic.deleted"
	addFantahseaDirGalleryImgEventBus  = "fantahsea.dir.gallery.image.add"
	notifyFantahseaFileDeletedEventBus = "fantahsea.notify.file.deleted"
)

type NotifyFileDeletedEvent struct {
	FileKey string `json:"fileKey"`
}

type StreamEvent struct {
	Timestamp uint32                       `json:"timestamp"` // epoch time second
	Schema    string                       `json:"schema"`
	Table     string                       `json:"table"`
	Type      string                       `json:"type"`    // INS-INSERT, UPD-UPDATE, DEL-DELETE
	Columns   map[string]StreamEventColumn `json:"columns"` // key is the column name
}

type StreamEventColumn struct {
	DataType string `json:"dataType"`
	Before   string `json:"before"`
	After    string `json:"after"`
}

type CreateFantahseaImgEvt struct {
	Username     string `json:"username"`
	UserNo       string `json:"userNo"`
	DirFileKey   string `json:"dirFileKey"`
	DirName      string `json:"dirName"`
	ImageName    string `json:"imageName"`
	ImageFileKey string `json:"imageFileKey"`
}

// event-pump send binlog event when a file_info record is saved.
// vfm guesses if the file is an image by file name,
// if so, vfm sends events to hammer to compress the image as a thumbnail
func OnFileSaved(rail common.Rail, evt StreamEvent) error {
	if evt.Type != "INS" {
		return nil
	}

	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		rail.Errorf("Event doesn't contain uuid, %+v", evt)
		return nil
	}
	uuid = uuidCol.After

	// lock before we do anything about it
	return _lockFileExec(rail, uuid, func() error {
		f, err := findFile(rail, uuid)
		if err != nil {
			return err
		}
		if f.IsZero() {
			rail.Infof("file is deleted, %v", uuid)
			return nil // file already deleted
		}

		if f.FileType != FILE_TYPE_FILE {
			rail.Infof("file is dir, %v", uuid)
			return nil // a directory
		}

		if f.Thumbnail != "" {
			rail.Infof("file has thumbnail aleady, %v", uuid)
			return nil // already has a thumbnail
		}

		if !isImage(f.Name) {
			rail.Infof("file is not image, %v, %v", uuid, f.Name)
			return nil // not an image
		}

		if e := bus.SendToEventBus(rail, CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcEventBus); e != nil {
			return common.TraceErrf(e, "Failed to send CompressImageEvent, uuid: %v", f.Uuid)
		}
		return nil
	})
}

// hammer sends event message when the thumbnail image is compressed and saved on mini-fstore
func OnImageCompressed(rail common.Rail, evt CompressImageEvent) error {
	rail.Infof("Received CompressedImageEvent, %+v", evt)
	return ReactOnImageCompressed(rail, evt)
}

// event-pump send binlog event when a file_info's thumbnail is updated.
// vfm receives the event and check if the file has a thumbnail,
// if so, sends events to fantahsea to create a gallery image,
// adding current image to the gallery for its directory
func OnThumbnailUpdated(rail common.Rail, evt StreamEvent) error {
	if evt.Type != "UPD" {
		return nil
	}

	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		rail.Errorf("Event doesn't contain uuid column, %+v", evt)
		return nil
	}
	uuid = uuidCol.After

	thumbnailCol, ok := evt.Columns["thumbnail"]
	if !ok || thumbnailCol.After == "" {
		return nil
	}

	// lock before we do anything about it
	return _lockFileExec(rail, uuid, func() error {
		f, err := findFile(rail, uuid)
		if err != nil {
			return err
		}
		if f.IsZero() || f.FileType != FILE_TYPE_FILE {
			return nil
		}

		if f.Thumbnail == "" || f.ParentFile == "" {
			return nil
		}

		pf, err := findFile(rail, f.ParentFile)
		if err != nil {
			return err
		}
		if pf.IsZero() {
			rail.Infof("parent file not found, %v", pf)
			return nil
		}

		user, err := CachedFindUser(rail, f.UploaderId)
		if err != nil {
			return err
		}

		evt := CreateFantahseaImgEvt{
			Username:     user.Username,
			UserNo:       user.UserNo,
			DirFileKey:   pf.Uuid,
			DirName:      pf.Name,
			ImageName:    f.Name,
			ImageFileKey: f.Uuid,
		}
		return bus.SendToEventBus(rail, evt, addFantahseaDirGalleryImgEventBus)
	})
}

// event-pump send binlog event when a file_info is deleted (is_logic_deleted changed)
// vfm notifies fantahsea about the delete
func OnFileDeleted(rail common.Rail, evt StreamEvent) error {
	if evt.Type != "UPD" {
		return nil
	}

	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		rail.Errorf("Event doesn't contain uuid, %+v", evt)
		return nil
	}
	uuid = uuidCol.After

	isLogicDeletedCol, ok := evt.Columns["is_logic_deleted"]
	if !ok {
		rail.Errorf("Event doesn't contain is_logic_deleted, %+v", evt)
		return nil
	}

	if isLogicDeletedCol.After != "1" { // FILE_LDEL_Y
		return nil
	}

	rail.Infof("File logically deleted, %v", uuid)

	if e := bus.SendToEventBus(rail, NotifyFileDeletedEvent{FileKey: uuid}, notifyFantahseaFileDeletedEventBus); e != nil {
		return common.TraceErrf(e, "Failed to send NotifyFileDeletedEvent, uuid: %v", uuid)
	}
	return nil
}
