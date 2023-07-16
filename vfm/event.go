package vfm

import (
	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
)

const (
	comprImgProcEventBus              = "hammer.image.compress.processing"
	comprImgNotifyEventBus            = "hammer.image.compress.notification"
	fileSavedEventBus                 = "vfm.file.saved"
	thumbnailUpdatedEventBus          = "vfm.file.thumbnail.updated"
	addFantahseaDirGalleryImgEventBus = "fantahsea.dir.gallery.image.add"
)

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
func OnFileSaved(evt StreamEvent) error {
	c := common.EmptyExecContext()
	if evt.Type != "INS" {
		return nil
	}

	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		c.Log.Errorf("Event doesn't contain uuid, %+v", evt)
		return nil
	}
	uuid = uuidCol.After

	// lock before we do anything about it
	return _lockFileExec(c, uuid, func() error {
		f, err := findFile(c, uuid)
		if err != nil {
			return err
		}
		if f.IsZero() {
			c.Log.Infof("file is deleted, %v", uuid)
			return nil // file already deleted
		}

		if f.FileType != FILE_TYPE_FILE {
			c.Log.Infof("file is dir, %v", uuid)
			return nil // a directory
		}

		if f.Thumbnail != "" {
			c.Log.Infof("file has thumbnail aleady, %v", uuid)
			return nil // already has a thumbnail
		}

		if !isImage(f.Name) {
			c.Log.Infof("file is not image, %v, %v", uuid, f.Name)
			return nil // not an image
		}

		if e := bus.SendToEventBus(c, CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcEventBus); e != nil {
			return common.TraceErrf(e, "Failed to send CompressImageEvent, uuid: %v", f.Uuid)
		}
		return nil
	})
}

// hammer sends event message when the thumbnail image is compressed and saved on mini-fstore
func OnImageCompressed(evt CompressImageEvent) error {
	cc := common.EmptyExecContext()
	cc.Log.Infof("Received CompressedImageEvent, %+v", evt)
	return ReactOnImageCompressed(cc, evt)
}

// event-pump send binlog event when a file_info's thumbnail is updated.
// vfm receives the event and check if the file has a thumbnail,
// if so, sends events to fantahsea to create a gallery image,
// adding current image to the gallery for its directory
func OnThumbnailUpdated(evt StreamEvent) error {
	if evt.Type != "UPD" {
		return nil
	}

	c := common.EmptyExecContext()
	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		c.Log.Errorf("Event doesn't contain uuid column, %+v", evt)
		return nil
	}
	uuid = uuidCol.After

	thumbnailCol, ok := evt.Columns["thumbnail"]
	if !ok || thumbnailCol.After == "" {
		return nil
	}

	// lock before we do anything about it
	return _lockFileExec(c, uuid, func() error {
		f, err := findFile(c, uuid)
		if err != nil {
			return err
		}
		if f.IsZero() || f.FileType != FILE_TYPE_FILE {
			return nil
		}

		if f.Thumbnail == "" || f.ParentFile == "" {
			return nil
		}

		pf, err := findFile(c, f.ParentFile)
		if err != nil {
			return err
		}
		if pf.IsZero() {
			c.Log.Infof("parent file not found, %v", pf)
			return nil
		}

		user, err := CachedFindUser(c, f.UploaderId)
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
		return bus.SendToEventBus(c, evt, addFantahseaDirGalleryImgEventBus)
	})
}
