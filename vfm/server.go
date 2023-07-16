package vfm

import (
	"context"

	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

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

	bus.SubscribeEventBus(comprImgNotifyEventBus, 2, OnImageCompressed)
	bus.SubscribeEventBus(fileSavedEventBus, 2, OnFileSaved)
	bus.SubscribeEventBus(thumbnailUpdatedEventBus, 2, OnThumbnailUpdated)

	server.Get("/open/api/file/upload/duplication/preflight",
		func(c *gin.Context, ec common.ExecContext) (any, error) {
			filename := c.Query("fileName")
			parentFileKey := c.Query("parentFileKey")
			ec.Log.Debugf("filename: %v, parentFileKey: %v", filename, parentFileKey)
			return FileExists(ec, filename, parentFileKey)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/parent",
		func(c *gin.Context, ec common.ExecContext) (any, error) {
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
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User fetch parent file info", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/move-to-dir",
		func(c *gin.Context, ec common.ExecContext, req MoveIntoDirReq) (any, error) {
			return nil, MoveFileToDir(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/make-dir",
		func(c *gin.Context, ec common.ExecContext, req MakeDirReq) (any, error) {
			return MakeDir(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/grant-access",
		func(c *gin.Context, ec common.ExecContext, req GrantAccessReq) (any, error) {
			uid, e := FindUserId(ec, req.GrantedTo)
			if e != nil {
				ec.Log.Warnf("Unable to find user id, grantedTo: %s, %v", req.GrantedTo, e)
				return nil, common.NewWebErr("Failed to find user")
			}
			return nil, GranteFileAccess(ec, uid, req.FileId)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list-granted-access",
		func(c *gin.Context, ec common.ExecContext, req ListGrantedAccessReq) (any, error) {
			return ListGrantedFileAccess(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/remove-granted-access",
		func(c *gin.Context, ec common.ExecContext, req RemoveGrantedAccessReq) (any, error) {
			return nil, RemoveGrantedFileAccess(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/dir/list",
		func(c *gin.Context, ec common.ExecContext) (any, error) {
			return ListDirs(ec)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list",
		func(c *gin.Context, ec common.ExecContext, req ListFileReq) (any, error) {
			return ListFiles(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/delete",
		func(c *gin.Context, ec common.ExecContext, req DeleteFileReq) (any, error) {
			return nil, DeleteFile(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/create",
		func(c *gin.Context, ec common.ExecContext, req CreateFileReq) (any, error) {
			return nil, CreateFile(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/info/update",
		func(c *gin.Context, ec common.ExecContext, req UpdateFileReq) (any, error) {
			return nil, UpdateFile(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/tag/list/all",
		func(c *gin.Context, ec common.ExecContext) (any, error) {
			return ListAllTags(ec)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag/list-for-file",
		func(c *gin.Context, ec common.ExecContext, req ListFileTagReq) (any, error) {
			return ListFileTags(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag",
		func(c *gin.Context, ec common.ExecContext, req TagFileReq) (any, error) {
			return nil, TagFile(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/untag",
		func(c *gin.Context, ec common.ExecContext, req UntagFileReq) (any, error) {
			return nil, UntagFile(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/vfolder/brief/owned",
		func(c *gin.Context, ec common.ExecContext) (any, error) {
			return ListVFolderBrief(ec)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/list",
		func(c *gin.Context, ec common.ExecContext, req ListVFolderReq) (any, error) {
			ec.Log.Debugf("req: %+v", req)
			return ListVFolders(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/create",
		func(c *gin.Context, ec common.ExecContext, req CreateVFolderReq) (any, error) {
			return CreateVFolder(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/add",
		func(c *gin.Context, ec common.ExecContext, req AddFileToVfolderReq) (any, error) {
			return nil, AddFileToVFolder(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/remove",
		func(c *gin.Context, ec common.ExecContext, req RemoveFileFromVfolderReq) (any, error) {
			return nil, RemoveFileFromVFolder(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove file from virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/share",
		func(c *gin.Context, ec common.ExecContext, req ShareVfolderReq) (any, error) {
			sharedTo, e := FindUser(ec, FindUserReq{Username: &req.Username})
			if e != nil {
				ec.Log.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
				return nil, common.NewWebErr("Failed to find user")
			}
			return nil, ShareVFolder(ec, sharedTo, req.FolderNo)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/access/remove",
		func(c *gin.Context, ec common.ExecContext, req RemoveGrantedFolderAccessReq) (any, error) {
			return nil, RemoveVFolderAccess(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/granted/list",
		func(c *gin.Context, ec common.ExecContext, req ListGrantedFolderAccessReq) (any, error) {
			return ListGrantedFolderAccess(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/token/generate",
		func(c *gin.Context, ec common.ExecContext, req GenerateTempTokenReq) (any, error) {
			return GenTempToken(ec, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User generate temporary token", Code: MANAGE_FILE_CODE}),
	)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	server.IGet("/remote/user/file/indir/list", func(c *gin.Context, ec common.ExecContext, q ListFilesInDirReq) (any, error) {
		return ListFilesInDir(ec, q)
	})
	server.Get("/remote/user/file/info", func(c *gin.Context, ec common.ExecContext) (any, error) {
		return FetchFileInfoInternal(ec, c.Query("fileKey"))
	})
	server.IGet("/remote/user/file/owner/validation", func(c *gin.Context, ec common.ExecContext, q ValidateFileOwnerReq) (any, error) {
		return ValidateFileOwner(ec, q)
	})

	// ---------------------------------- endpoints used to compensate --------------------------------------

	server.Post("/compensate/image/compression", func(c *gin.Context, ec common.ExecContext) (any, error) {
		return nil, CompensateImageCompression(ec)
	})

	return nil
}
