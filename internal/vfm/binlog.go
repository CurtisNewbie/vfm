package vfm

import (
	"github.com/curtisnewbie/event-pump/binlog"
	"github.com/curtisnewbie/event-pump/client"
	"github.com/curtisnewbie/miso/middleware/mysql"
	"github.com/curtisnewbie/miso/miso"
)

func SubscribeBinlogChanges() {

	binlog.SubscribeBinlogEventsOnBootstrapV2(binlog.SubscribeBinlogOption{
		Pipeline: client.Pipeline{
			Schema:     miso.GetPropStr(mysql.PropMySQLSchema),
			Table:      "file_info",
			EventTypes: []client.EventType{client.EventTypeInsert},
			Stream:     FileSavedEventBus,
		},
		Concurrency:   2,
		ContinueOnErr: true,
		Listener:      OnFileSaved,
	})

	binlog.SubscribeBinlogEventsOnBootstrapV2(binlog.SubscribeBinlogOption{
		Pipeline: client.Pipeline{
			Schema:     miso.GetPropStr(mysql.PropMySQLSchema),
			Table:      "file_info",
			EventTypes: []client.EventType{client.EventTypeUpdate},
			Stream:     ThumbnailUpdatedEventBus,
			Condition: client.Condition{
				ColumnChanged: []string{"thumbnail"},
			},
		},
		Concurrency:   2,
		ContinueOnErr: true,
		Listener:      OnThumbnailUpdated,
	})

	binlog.SubscribeBinlogEventsOnBootstrapV2(binlog.SubscribeBinlogOption{
		Pipeline: client.Pipeline{
			Schema:     miso.GetPropStr(mysql.PropMySQLSchema),
			Table:      "file_info",
			EventTypes: []client.EventType{client.EventTypeUpdate},
			Stream:     FileLDeletedEventBus,
			Condition: client.Condition{
				ColumnChanged: []string{"is_logic_deleted"},
			},
		},
		Concurrency:   2,
		ContinueOnErr: true,
		Listener:      OnFileDeleted,
	})
}
