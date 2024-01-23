package vfm

import (
	"fmt"
	"time"

	hammer "github.com/curtisnewbie/hammer/api"
	fstore "github.com/curtisnewbie/mini-fstore/api"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

const (
	VfmFileSavedEventBus         = "event.bus.vfm.file.saved"
	VfmThumbnailUpdatedEventBus  = "event.bus.vfm.file.thumbnail.updated"
	VfmFileLDeletedEventBus      = "event.bus.vfm.file.logic.deleted"
	VfmCalcDirSizeEventBus       = "event.bus.vfm.dir.size.calc"
	VfmAddFileToVFolderEventBus  = "event.bus.vfm.file.vfolder.add"
	VfmCompressImgNotifyEventBus = "vfm.image.compressed.event"
	VfmUnzipResultNotifyEventBus = "vfm.unzip.result.notify.event"

	AddFantahseaDirGalleryImgEventBus  = "event.bus.fantahsea.dir.gallery.image.add"
	NotifyFantahseaFileDeletedEventBus = "event.bus.fantahsea.notify.file.deleted"
)

func PrepareEventBus(rail miso.Rail) error {
	miso.SubEventBus(VfmCompressImgNotifyEventBus, 2, OnImageCompressed)
	miso.SubEventBus(VfmFileSavedEventBus, 2, OnFileSaved)
	miso.SubEventBus(VfmThumbnailUpdatedEventBus, 2, OnThumbnailUpdated)
	miso.SubEventBus(VfmFileLDeletedEventBus, 2, OnFileDeleted)
	miso.SubEventBus(VfmAddFileToVFolderEventBus, 2, OnAddFileToVfolderEvent)
	miso.SubEventBus(AddFantahseaDirGalleryImgEventBus, 2, OnCreateGalleryImgEvent)
	miso.SubEventBus(NotifyFantahseaFileDeletedEventBus, 2, OnNotifyFileDeletedEvent)
	miso.SubEventBus(VfmCalcDirSizeEventBus, 1, OnCalcDirSizeEvt)
	miso.SubEventBus(VfmUnzipResultNotifyEventBus, 2, OnUnzipFileReplyEvent)
	return nil
}

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
func OnFileSaved(rail miso.Rail, evt StreamEvent) error {
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
	lock := fileLock(rail, uuid)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	f, err := findFile(rail, miso.GetMySQL(), uuid)
	if err != nil {
		return err
	}
	if f.IsZero() {
		rail.Infof("file is deleted, %v", uuid)
		return nil // file already deleted
	}

	if f.FileType != FileTypeFile {
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

	event := hammer.ImageCompressTriggerEvent{Identifier: f.Uuid, FileId: f.FstoreFileId, ReplyTo: VfmCompressImgNotifyEventBus}
	if e := miso.PubEventBus(rail, event, hammer.CompressImageTriggerEventBus); e != nil {
		return miso.TraceErrf(e, "Failed to send CompressImageEvent, uuid: %v", f.Uuid)
	}
	return nil
}

// hammer sends event message when the thumbnail image is compressed and saved on mini-fstore
func OnImageCompressed(rail miso.Rail, evt hammer.ImageCompressReplyEvent) error {
	rail.Infof("Received CompressedImageEvent, %+v", evt)
	return ReactOnImageCompressed(rail, miso.GetMySQL(), evt)
}

func ReactOnImageCompressed(rail miso.Rail, tx *gorm.DB, evt hammer.ImageCompressReplyEvent) error {
	fileKey := evt.Identifier
	lock := miso.NewRLock(rail, "file:uuid:"+fileKey)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	f, e := findFile(rail, tx, fileKey)
	if e != nil {
		rail.Errorf("Unable to find file, uuid: %v, %v", fileKey, e)
		return nil
	}
	if f.IsZero() {
		rail.Errorf("File not found, uuid: %v", fileKey)
		return nil
	}

	return tx.Exec("UPDATE file_info SET thumbnail = ? WHERE uuid = ?", evt.FileId, fileKey).
		Error
}

// event-pump send binlog event when a file_info's thumbnail is updated.
// vfm receives the event and check if the file has a thumbnail,
// if so, sends events to fantahsea to create a gallery image,
// adding current image to the gallery for its directory
func OnThumbnailUpdated(rail miso.Rail, evt StreamEvent) error {
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
	lock := fileLock(rail, uuid)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	f, err := findFile(rail, miso.GetMySQL(), uuid)
	if err != nil {
		return err
	}
	if f.IsZero() || f.FileType != FileTypeFile {
		return nil
	}

	if f.Thumbnail == "" || f.ParentFile == "" {
		return nil
	}

	pf, err := findFile(rail, miso.GetMySQL(), f.ParentFile)
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

	cfi := CreateFantahseaImgEvt{
		Username:     user.Username,
		UserNo:       user.UserNo,
		DirFileKey:   pf.Uuid,
		DirName:      pf.Name,
		ImageName:    f.Name,
		ImageFileKey: f.Uuid,
	}
	return miso.PubEventBus(rail, cfi, AddFantahseaDirGalleryImgEventBus)
}

// event-pump send binlog event when a file_info is deleted (is_logic_deleted changed)
// vfm notifies fantahsea about the delete
func OnFileDeleted(rail miso.Rail, evt StreamEvent) error {
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

	if e := miso.PubEventBus(rail, NotifyFileDeletedEvent{FileKey: uuid}, NotifyFantahseaFileDeletedEventBus); e != nil {
		return miso.TraceErrf(e, "Failed to send NotifyFileDeletedEvent, uuid: %v", uuid)
	}
	return nil
}

type AddFileToVfolderEvent struct {
	UserId   int
	Username string
	UserNo   string
	FolderNo string
	FileKeys []string
}

func OnAddFileToVfolderEvent(rail miso.Rail, evt AddFileToVfolderEvent) error {
	return HandleAddFileToVFolderEvent(rail, miso.GetMySQL(), evt)
}

type CalcDirSizeEvt struct {
	FileKey string
}

func OnCalcDirSizeEvt(rail miso.Rail, evt CalcDirSizeEvt) error {
	defer miso.TimeOp(rail, time.Now(), fmt.Sprintf("Process CalcDirSizeEvt: %+v", evt))
	return CalcDirSize(rail, evt.FileKey, miso.GetMySQL())
}

func OnUnzipFileReplyEvent(rail miso.Rail, evt fstore.UnzipFileReplyEvent) error {
	rail.Infof("received UnzipFileReplyEvent: %+v", evt)
	return HandleZipUnpackResult(rail, miso.GetMySQL(), evt)
}
