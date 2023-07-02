package vfm

import (
	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
)

const (
	fileSavedEventBus = "vfm.file.saved"
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

// event-pump send binlog event when a file_info record is saved
//
// e.g., on event-pump
//
//	pipeline:
//	  - schema: 'fileserver'
//	    table: 'file_info'
//	    type: '^(INS)$'
//	    stream: 'vfm.file.saved'
//	    enabled: true
func OnFileSaved(evt StreamEvent) error {
	c := common.EmptyExecContext()
	if evt.Type != "INS" {
		return nil
	}

	var uuid string
	uuidCol, ok := evt.Columns["uuid"]
	if !ok {
		c.Log.Errorf("Event doesn't contain uuid, %+v", evt)
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

		if e := bus.SendToEventBus(c, CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcBus); e != nil {
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
