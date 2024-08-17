// auto generated by misoapi v0.1.5-beta.1 at 2024/08/17 16:42:19, please do not modify
package vfm

import (
	"github.com/curtisnewbie/miso/middleware/mysql"
	"github.com/curtisnewbie/miso/miso"
)

func init() {
	miso.IGet("/open/api/file/upload/duplication/preflight",
		func(inb *miso.Inbound, req PreflightCheckReq) (bool, error) {
			return DupPreflightCheckEp(inb, req)
		}).
		Desc("Preflight check for duplicate file uploads").
		Resource(ManageFilesResource)

	miso.IGet("/open/api/file/parent",
		func(inb *miso.Inbound, req FetchParentFileReq) (*ParentFileInfo, error) {
			return GetParentFileEp(inb, req)
		}).
		Desc("User fetch parent file info").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/move-to-dir",
		func(inb *miso.Inbound, req MoveIntoDirReq) (any, error) {
			return MoveFileToDirEp(inb, req)
		}).
		Desc("User move files into directory").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/make-dir",
		func(inb *miso.Inbound, req MakeDirReq) (string, error) {
			return MakeDirEp(inb, req)
		}).
		Desc("User make directory").
		Resource(ManageFilesResource)

	miso.Get("/open/api/file/dir/list",
		func(inb *miso.Inbound) ([]ListedDir, error) {
			return ListDirEp(inb)
		}).
		Desc("User list directories").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/list",
		func(inb *miso.Inbound, req ListFileReq) (miso.PageRes[ListedFile], error) {
			return ListFilesEp(inb, req)
		}).
		Desc("User list files").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/delete",
		func(inb *miso.Inbound, req DeleteFileReq) (any, error) {
			return DeleteFileEp(inb, req)
		}).
		Desc("User delete file").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/dir/truncate",
		func(inb *miso.Inbound, req DeleteFileReq) (any, error) {
			return TruncateDirEp(inb, req)
		}).
		Desc("User delete truncate directory recursively").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/delete/batch",
		func(inb *miso.Inbound, req BatchDeleteFileReq) (any, error) {
			return BatchDeleteFileEp(inb, req)
		}).
		Desc("User delete file in batch").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/create",
		func(inb *miso.Inbound, req CreateFileReq) (any, error) {
			return CreateFileEp(inb, req)
		}).
		Desc("User create file").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/info/update",
		func(inb *miso.Inbound, req UpdateFileReq) (any, error) {
			return UpdateFileEp(inb, req)
		}).
		Desc("User update file").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/token/generate",
		func(inb *miso.Inbound, req GenerateTempTokenReq) (string, error) {
			return GenFileTknEp(inb, req)
		}).
		Desc("User generate temporary token").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/file/unpack",
		func(inb *miso.Inbound, req UnpackZipReq) (any, error) {
			return UnpackZipEp(inb, req)
		}).
		Desc("User unpack zip").
		Resource(ManageFilesResource)

	miso.RawGet("/open/api/file/token/qrcode", GenFileTknQRCodeEp).
		Desc("User generate qrcode image for temporary token").
		Public().
		DocQueryParam("token", "Generated temporary file key")

	miso.Get("/open/api/vfolder/brief/owned",
		func(inb *miso.Inbound) ([]VFolderBrief, error) {
			return ListVFolderBriefEp(inb)
		}).
		Desc("User list virtual folder briefs").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/list",
		func(inb *miso.Inbound, req ListVFolderReq) (ListVFolderRes, error) {
			return ListVFoldersEp(inb, req)
		}).
		Desc("User list virtual folders").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/create",
		func(inb *miso.Inbound, req CreateVFolderReq) (string, error) {
			return CreateVFolderEp(inb, req)
		}).
		Desc("User create virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/file/add",
		func(inb *miso.Inbound, req AddFileToVfolderReq) (any, error) {
			return VFolderAddFileEp(inb, req)
		}).
		Desc("User add file to virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/file/remove",
		func(inb *miso.Inbound, req RemoveFileFromVfolderReq) (any, error) {
			return VFolderRemoveFileEp(inb, req)
		}).
		Desc("User remove file from virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/share",
		func(inb *miso.Inbound, req ShareVfolderReq) (any, error) {
			return ShareVFolderEp(inb, req)
		}).
		Desc("Share access to virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/access/remove",
		func(inb *miso.Inbound, req RemoveGrantedFolderAccessReq) (any, error) {
			return RemoveVFolderAccessEp(inb, req)
		}).
		Desc("Remove granted access to virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/granted/list",
		func(inb *miso.Inbound, req ListGrantedFolderAccessReq) (ListGrantedFolderAccessRes, error) {
			return ListVFolderAccessEp(inb, req)
		}).
		Desc("List granted access to virtual folder").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/vfolder/remove",
		func(inb *miso.Inbound, req RemoveVFolderReq) (any, error) {
			return RemoveVFolderEp(inb, req)
		}).
		Desc("Remove virtual folder").
		Resource(ManageFilesResource)

	miso.Get("/open/api/gallery/brief/owned",
		func(inb *miso.Inbound) ([]VGalleryBrief, error) {
			return ListGalleryBriefsEp(inb)
		}).
		Desc("List owned gallery brief info").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/new",
		func(inb *miso.Inbound, req CreateGalleryCmd) (*Gallery, error) {
			return CreateGalleryEp(inb, req)
		}).
		Desc("Create new gallery").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/update",
		func(inb *miso.Inbound, req UpdateGalleryCmd) (any, error) {
			return UpdateGalleryEp(inb, req)
		}).
		Desc("Update gallery").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/delete",
		func(inb *miso.Inbound, req DeleteGalleryCmd) (any, error) {
			return DeleteGalleryEp(inb, req)
		}).
		Desc("Delete gallery").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/list",
		func(inb *miso.Inbound, req ListGalleriesCmd) (miso.PageRes[VGallery], error) {
			return ListGalleriesEp(inb, req)
		}).
		Desc("List galleries").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/access/grant",
		func(inb *miso.Inbound, req PermitGalleryAccessCmd) (any, error) {
			return GranteGalleryAccessEp(inb, req)
		}).
		Desc("Grant access to the galleries").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/access/remove",
		func(inb *miso.Inbound, req RemoveGalleryAccessCmd) (any, error) {
			return RemoveGalleryAccessEp(inb, req)
		}).
		Desc("Remove access to the galleries").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/access/list",
		func(inb *miso.Inbound, req ListGrantedGalleryAccessCmd) (miso.PageRes[ListedGalleryAccessRes], error) {
			return ListGalleryAccessEp(inb, req)
		}).
		Desc("List granted access to the galleries").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/images",
		func(inb *miso.Inbound, req ListGalleryImagesCmd) (*ListGalleryImagesResp, error) {
			return ListGalleryImagesEp(inb, req)
		}).
		Desc("List images of gallery").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/gallery/image/transfer",
		func(inb *miso.Inbound, req TransferGalleryImageReq) (any, error) {
			return TransferGalleryImageEp(inb, req)
		}).
		Desc("Host selected images on gallery").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/list",
		func(inb *miso.Inbound, req ApiListVerFileReq) (miso.PageRes[ApiListVerFileRes], error) {
			return ApiListVersionedFile(inb, req)
		}).
		Desc("List versioned files").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/history",
		func(inb *miso.Inbound, req ApiListVerFileHistoryReq) (miso.PageRes[ApiListVerFileHistoryRes], error) {
			return ApiListVersionedFileHistory(inb, req)
		}).
		Desc("List versioned file history").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/accumulated-size",
		func(inb *miso.Inbound, req ApiQryVerFileAccuSizeReq) (ApiQryVerFileAccuSizeRes, error) {
			return ApiQryVersionedFileAccuSize(inb, req)
		}).
		Desc("Query versioned file log accumulated size").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/create",
		func(inb *miso.Inbound, req ApiCreateVerFileReq) (ApiCreateVerFileRes, error) {
			return ApiCreateVersionedFile(inb, req)
		}).
		Desc("Create versioned file").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/update",
		func(inb *miso.Inbound, req ApiUpdateVerFileReq) (any, error) {
			return ApiUpdateVersionedFile(inb, req)
		}).
		Desc("Update versioned file").
		Resource(ManageFilesResource)

	miso.IPost("/open/api/versioned-file/delete",
		func(inb *miso.Inbound, req ApiDelVerFileReq) (any, error) {
			return ApiDelVersionedFile(inb, req)
		}).
		Desc("Delete versioned file").
		Resource(ManageFilesResource)

	miso.Post("/compensate/thumbnail",
		func(inb *miso.Inbound) (any, error) {
			return CompensateThumbnailEp(inb.Rail(), mysql.GetMySQL())
		}).
		Desc("Compensate thumbnail generation")

	miso.Post("/compensate/dir/calculate-size",
		func(inb *miso.Inbound) (any, error) {
			return ImMemBatchCalcDirSizeEp(inb.Rail(), mysql.GetMySQL())
		}).
		Desc("Calculate size of all directories recursively")

	miso.Put("/bookmark/file/upload",
		func(inb *miso.Inbound) (any, error) {
			return UploadBookmarkFileEp(inb)
		}).
		Desc("Upload bookmark file").
		Resource(ResourceManageBookmark)

	miso.IPost("/bookmark/list",
		func(inb *miso.Inbound, req ListBookmarksReq) (any, error) {
			return ListBookmarksEp(inb, req)
		}).
		Desc("List bookmarks").
		Resource(ResourceManageBookmark)

	miso.IPost("/bookmark/remove",
		func(inb *miso.Inbound, req RemoveBookmarkReq) (any, error) {
			return RemoveBookmarkEp(inb, req)
		}).
		Desc("Remove bookmark").
		Resource(ResourceManageBookmark)

	miso.IPost("/bookmark/blacklist/list",
		func(inb *miso.Inbound, req ListBookmarksReq) (any, error) {
			return ListBlacklistedBookmarksEp(inb, req)
		}).
		Desc("List bookmark blacklist").
		Resource(ResourceManageBookmark)

	miso.IPost("/bookmark/blacklist/remove",
		func(inb *miso.Inbound, req RemoveBookmarkReq) (any, error) {
			return RemoveBookmarkBlacklistEp(inb, req)
		}).
		Desc("Remove bookmark blacklist").
		Resource(ResourceManageBookmark)

}
