package vfm

import (
	"fmt"
	"time"

	fstore "github.com/curtisnewbie/mini-fstore/api"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

const (
	FileSavedEventBus               = "event.bus.vfm.file.saved"
	ThumbnailUpdatedEventBus        = "event.bus.vfm.file.thumbnail.updated"
	FileLDeletedEventBus            = "event.bus.vfm.file.logic.deleted"
	CalcDirSizeEventBus             = "event.bus.vfm.dir.size.calc"
	AddFileToVFolderEventBus        = "event.bus.vfm.file.vfolder.add"
	CompressImgNotifyEventBus       = "vfm.image.compressed.event"
	GenVideoThumbnailNotifyEventBus = "vfm.video.thumbnail.generate"
	UnzipResultNotifyEventBus       = "vfm.unzip.result.notify.event"
	AddDirGalleryImgEventBus        = "event.bus.fantahsea.dir.gallery.image.add"
	SyncGalleryFileDeletedEventBus  = "event.bus.fantahsea.notify.file.deleted"
)

var (
	UnzipResultNotifyPipeline       = miso.NewEventPipeline[fstore.UnzipFileReplyEvent](UnzipResultNotifyEventBus)
	GenVideoThumbnailNotifyPipeline = miso.NewEventPipeline[fstore.GenVideoThumbnailReplyEvent](GenVideoThumbnailNotifyEventBus)
	CompressImgNotifyPipeline       = miso.NewEventPipeline[fstore.ImageCompressReplyEvent](CompressImgNotifyEventBus)
	AddFileToVFolderPipeline        = miso.NewEventPipeline[AddFileToVfolderEvent](AddFileToVFolderEventBus)
	CalcDirSizePipeline             = miso.NewEventPipeline[CalcDirSizeEvt](CalcDirSizeEventBus)
	FileLDeletedPipeline            = miso.NewEventPipeline[StreamEvent](FileLDeletedEventBus)
	ThumbnailUpdatedPipeline        = miso.NewEventPipeline[StreamEvent](ThumbnailUpdatedEventBus)
	FileSavedPipeline               = miso.NewEventPipeline[StreamEvent](FileSavedEventBus)
	AddDirGalleryImgPipeline        = miso.NewEventPipeline[CreateGalleryImgEvent](AddDirGalleryImgEventBus)
	SyncGalleryFileDeletedPipeline  = miso.NewEventPipeline[NotifyFileDeletedEvent](SyncGalleryFileDeletedEventBus)
)

func PrepareEventBus(rail miso.Rail) error {
	UnzipResultNotifyPipeline.Listen(2, OnUnzipFileReplyEvent)
	GenVideoThumbnailNotifyPipeline.Listen(2, OnVidoeThumbnailGenerated)
	CompressImgNotifyPipeline.Listen(2, OnImageCompressed)
	AddFileToVFolderPipeline.Listen(2, OnAddFileToVfolderEvent)
	CalcDirSizePipeline.Listen(1, OnCalcDirSizeEvt)
	FileLDeletedPipeline.Listen(2, OnFileDeleted)
	ThumbnailUpdatedPipeline.Listen(2, OnThumbnailUpdated)
	FileSavedPipeline.Listen(2, OnFileSaved)
	AddDirGalleryImgPipeline.Listen(2, OnCreateGalleryImgEvent)
	SyncGalleryFileDeletedPipeline.Listen(2, OnNotifyFileDeletedEvent)
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
	if f == nil {
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

	if isImage(f.Name) {
		evt := fstore.ImgThumbnailTriggerEvent{Identifier: f.Uuid, FileId: f.FstoreFileId, ReplyTo: CompressImgNotifyEventBus}
		if err := fstore.GenImgThumbnailPipeline.Send(rail, evt); err != nil {
			return fmt.Errorf("failed to send %#v, uuid: %v, %v", evt, f.Uuid, err)
		}
		return nil
	}

	if isVideo(f.Name) {
		evt := fstore.VidThumbnailTriggerEvent{
			Identifier: f.Uuid,
			FileId:     f.FstoreFileId,
			ReplyTo:    GenVideoThumbnailNotifyEventBus,
		}
		if err := fstore.GenVidThumbnailPipeline.Send(rail, evt); err != nil {
			return fmt.Errorf("failed to send %#v, uuid: %v, %v", evt, f.Uuid, err)
		}
		return nil
	}

	return nil
}

// hammer sends event message when the thumbnail image is compressed and saved on mini-fstore
func OnImageCompressed(rail miso.Rail, evt fstore.ImageCompressReplyEvent) error {
	rail.Infof("Receive %#v", evt)
	return OnThumbnailGenerated(rail, miso.GetMySQL(), evt.Identifier, evt.FileId)
}

func OnVidoeThumbnailGenerated(rail miso.Rail, evt fstore.GenVideoThumbnailReplyEvent) error {
	rail.Infof("Receive %#v", evt)
	return OnThumbnailGenerated(rail, miso.GetMySQL(), evt.Identifier, evt.FileId)
}

func OnThumbnailGenerated(rail miso.Rail, tx *gorm.DB, identifier string, fileId string) error {
	fileKey := identifier
	lock := fileLock(rail, fileKey)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	f, e := findFile(rail, tx, fileKey)
	if e != nil {
		rail.Errorf("Unable to find file, uuid: %v, %v", fileKey, e)
		return nil
	}
	if f == nil {
		rail.Errorf("File not found, uuid: %v", fileKey)
		return nil
	}

	return tx.Exec("UPDATE file_info SET thumbnail = ? WHERE uuid = ?", fileId, fileKey).
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
	if f == nil || f.FileType != FileTypeFile {
		return nil
	}

	if f.Thumbnail == "" || f.ParentFile == "" {
		return nil
	}
	if !isImage(f.Name) {
		return nil
	}

	pf, err := findFile(rail, miso.GetMySQL(), f.ParentFile)
	if err != nil {
		return err
	}
	if pf == nil {
		rail.Infof("parent file not found, %v", f.ParentFile)
		return nil
	}

	user, err := CachedFindUser(rail, f.UploaderNo)
	if err != nil {
		return err
	}

	cfi := CreateGalleryImgEvent{
		Username:     user.Username,
		UserNo:       user.UserNo,
		DirFileKey:   pf.Uuid,
		DirName:      pf.Name,
		ImageName:    f.Name,
		ImageFileKey: f.Uuid,
	}
	return AddDirGalleryImgPipeline.Send(rail, cfi)
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

	if e := SyncGalleryFileDeletedPipeline.Send(rail, NotifyFileDeletedEvent{FileKey: uuid}); e != nil {
		return fmt.Errorf("failed to send NotifyFileDeletedEvent, uuid: %v, %v", uuid, e)
	}
	return nil
}

type AddFileToVfolderEvent struct {
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
