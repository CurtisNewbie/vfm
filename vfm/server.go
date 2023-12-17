package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/miso/miso"
)

func PrepareServer(rail miso.Rail) error {
	common.LoadBuiltinPropagationKeys()

	if err := PrepareGoAuthReport(rail); err != nil {
		return err
	}

	if err := PrepareEventBus(rail); err != nil {
		return err
	}

	if err := RegisterHttpRoutes(rail); err != nil {
		return err
	}
	return nil
}

func PrepareEventBus(rail miso.Rail) error {
	if err := miso.NewEventBus(comprImgProcEventBus); err != nil {
		return err
	}
	if err := miso.NewEventBus(addFantahseaDirGalleryImgEventBus); err != nil {
		return err
	}
	if err := miso.NewEventBus(notifyFantahseaFileDeletedEventBus); err != nil {
		return err
	}
	if err := miso.NewEventBus(addFileToVFolderEventBus); err != nil {
		return err
	}

	miso.SubEventBus(comprImgNotifyEventBus, 2, OnImageCompressed)
	miso.SubEventBus(fileSavedEventBus, 2, OnFileSaved)
	miso.SubEventBus(thumbnailUpdatedEventBus, 2, OnThumbnailUpdated)
	miso.SubEventBus(fileLDeletedEventBus, 2, OnFileDeleted)
	miso.SubEventBus(addFileToVFolderEventBus, 2, OnAddFileToVfolderEvent)
	miso.SubEventBus(addFantahseaDirGalleryImgEventBus, 2, OnCreateGalleryImgEvent)
	miso.SubEventBus(notifyFantahseaFileDeletedEventBus, 2, OnNotifyFileDeletedEvent)
	return nil
}

func PrepareGoAuthReport(rail miso.Rail) error {
	goauth.ReportResourcesOnBootstrapped(rail, []goauth.AddResourceReq{
		{Name: ManageFileResName, Code: ManageFileResCode},
		{Name: AdminFsResName, Code: AdminFsResCode},
	})
	goauth.ReportPathsOnBootstrapped(rail)
	return nil
}
