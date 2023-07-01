package vfm

import (
	"context"

	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

func compensateImageCompressionEp(c *gin.Context, ec common.ExecContext) (any, error) {
	return nil, CompensateImageCompression(ec)
}

func listGrantedFolderAccessEp(c *gin.Context, ec common.ExecContext, req ListGrantedFolderAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListGrantedFolderAccess(ec, req)
}

func removeGrantedFolderAccessEp(c *gin.Context, ec common.ExecContext, req RemoveGrantedFolderAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, RemoveVFolderAccess(ec, req)
}

func shareVFolderEp(c *gin.Context, ec common.ExecContext, req ShareVfolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	sharedTo, e := FindUser(ec, FindUserReq{Username: &req.Username})
	if e != nil {
		ec.Log.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
		return nil, common.NewWebErr("Failed to find user")
	}
	return nil, ShareVFolder(ec, sharedTo, req.FolderNo)
}

func removeFileFromVfolderEp(c *gin.Context, ec common.ExecContext, req RemoveFileFromVfolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, RemoveFileFromVFolder(ec, req)
}

func addFileToVFolderEp(c *gin.Context, ec common.ExecContext, req AddFileToVfolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, AddFileToVFolder(ec, req)
}

func createVFolderEp(c *gin.Context, ec common.ExecContext, req CreateVFolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return CreateVFolder(ec, req)
}

func listVfoldersEp(c *gin.Context, ec common.ExecContext, req ListVFolderReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListVFolders(ec, req)
}

func listVfolderBriefEp(c *gin.Context, ec common.ExecContext) (any, error) {
	return ListVFolderBrief(ec)
}

func untagFileEp(c *gin.Context, ec common.ExecContext, req UntagFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, UntagFile(ec, req)
}

func tagFileEp(c *gin.Context, ec common.ExecContext, req TagFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, TagFile(ec, req)
}

func listFileTagsEp(c *gin.Context, ec common.ExecContext, req ListFileTagReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListFileTags(ec, req)
}

func listAllTagsEp(c *gin.Context, ec common.ExecContext) (any, error) {
	return ListAllTags(ec)
}

func updateFileEp(c *gin.Context, ec common.ExecContext, req UpdateFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, UpdateFile(ec, req)
}

func uploadPreflightCheckEp(c *gin.Context, ec common.ExecContext) (any, error) {
	filename := c.Query("fileName")
	parentFileKey := c.Query("parentFileKey")
	ec.Log.Debugf("filename: %v, parentFileKey: %v", filename, parentFileKey)
	return FileExists(ec, filename, parentFileKey)
}

func fetchParentFileInfoEp(c *gin.Context, ec common.ExecContext) (any, error) {
	fk := c.Query("fileKey")
	if fk == "" {
		return nil, common.NewWebErr("fileKey is required")
	}
	ec.Log.Debugf("fileKey: %v", fk)
	pf, e := FindParentFile(ec, fk)
	if e != nil {
		return nil, e
	}
	if pf.Zero {
		return nil, nil
	}
	return pf, nil
}

func moveFileIntoDirEp(c *gin.Context, ec common.ExecContext, req MoveIntoDirReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, MoveFileToDir(ec, req)
}

func makeDirEp(c *gin.Context, ec common.ExecContext, req MakeDirReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return MakeDir(ec, req)
}

func grantAccessEp(c *gin.Context, ec common.ExecContext, req GrantAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	uid, e := FindUserId(ec, req.GrantedTo)
	if e != nil {
		ec.Log.Warnf("Unable to find user id, grantedTo: %s, %v", req.GrantedTo, e)
		return nil, common.NewWebErr("Failed to find user")
	}
	return nil, GranteFileAccess(ec, uid, req.FileId)
}

func listGrantedAccessEp(c *gin.Context, ec common.ExecContext, req ListGrantedAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListGrantedFileAccess(ec, req)
}

func removeGrantedAccessEp(c *gin.Context, ec common.ExecContext, req RemoveGrantedAccessReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, RemoveGrantedFileAccess(ec, req)
}

func listDirsEp(c *gin.Context, ec common.ExecContext) (any, error) {
	return ListDirs(ec)
}

func listFilesEp(c *gin.Context, ec common.ExecContext, req ListFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return ListFiles(ec, req)
}

func deleteFileEp(c *gin.Context, ec common.ExecContext, req DeleteFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, DeleteFile(ec, req)
}

func createFileEp(c *gin.Context, ec common.ExecContext, req CreateFileReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return nil, CreateFile(ec, req)
}

func generateTempTokenEp(c *gin.Context, ec common.ExecContext, req GenerateTempTokenReq) (any, error) {
	ec.Log.Debugf("req: %+v", req)
	return GenTempToken(ec, req)
}

func listFilesInDirInternalEp(c *gin.Context, ec common.ExecContext, q ListFilesInDirReq) (any, error) {
	return ListFilesInDir(ec, q)
}

func fetchFileInfoInternalEp(c *gin.Context, ec common.ExecContext) (any, error) {
	return FetchFileInfoInternal(ec, c.Query("fileKey"))
}

func validateFileOwnerEp(c *gin.Context, ec common.ExecContext, q ValidateFileOwnerReq) (any, error) {
	return ValidateFileOwner(ec, q)
}

func PrepareServer(c common.ExecContext) error {
	if goauth.IsEnabled() {
		server.PostServerBootstrapped(func(sc common.ExecContext) error {
			c := common.EmptyExecContext()
			if e := goauth.AddResource(context.Background(), goauth.AddResourceReq{Name: MANAGE_FILE_NAME, Code: MANAGE_FILE_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource, %v", e)
			}

			if e := goauth.AddResource(context.Background(), goauth.AddResourceReq{Name: ADMIN_FS_NAME, Code: ADMIN_FS_CODE}); e != nil {
				c.Log.Errorf("Failed to create goauth resource, %v", e)
			}
			return nil
		})
		goauth.ReportPathsOnBootstrapped()
	}

	if e := bus.DeclareEventBus(comprImgNotifyBus); e != nil {
		return e
	}

	if e := bus.DeclareEventBus(comprImgProcBus); e != nil {
		return e
	}

	if err := bus.SubscribeEventBus(comprImgNotifyBus, 2, func(evt CompressImageEvent) error {
		cc := common.EmptyExecContext()
		cc.Log.Infof("Received CompressedImageEvent, %+v", evt)
		return ReactOnImageCompressed(cc, evt)
	}); err != nil {
		return err
	}

	server.Get("/open/api/file/upload/duplication/preflight", uploadPreflightCheckEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/parent", fetchParentFileInfoEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User fetch parent file info", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/move-to-dir", moveFileIntoDirEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/make-dir", makeDirEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/grant-access", grantAccessEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/list-granted-access", listGrantedAccessEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/remove-granted-access", removeGrantedAccessEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/dir/list", listDirsEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/list", listFilesEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/delete", deleteFileEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/create", createFileEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create file", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/info/update", updateFileEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/file/tag/list/all", listAllTagsEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/tag/list-for-file", listFileTagsEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/tag", tagFileEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/untag", untagFileEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}))

	server.Get("/open/api/vfolder/brief/owned", listVfolderBriefEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/list", listVfoldersEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/create", createVFolderEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/file/add", addFileToVFolderEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/file/remove", removeFileFromVfolderEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove file from virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/share", shareVFolderEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/access/remove", removeGrantedFolderAccessEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/vfolder/granted/list", listGrantedFolderAccessEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}))

	server.IPost("/open/api/file/token/generate", generateTempTokenEp,
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User generate temporary token", Code: MANAGE_FILE_CODE}))

	// ---------------------------------------------- internal endpoints ------------------------------------------

	server.IGet("/remote/user/file/indir/list", listFilesInDirInternalEp)
	server.Get("/remote/user/file/info", fetchFileInfoInternalEp)
	server.IGet("/remote/user/file/owner/validation", validateFileOwnerEp)

	// ---------------------------------- endpoints used to compensate --------------------------------------

	server.Post("/compensate/image/compression", compensateImageCompressionEp)

	return nil
}
