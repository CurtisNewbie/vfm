package vfm

import (
	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

func PrepareServer(rail common.Rail) error {
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

func RegisterHttpRoutes(rail common.Rail) error {
	server.IGet("/open/api/file/upload/duplication/preflight",
		func(c *gin.Context, rail common.Rail, req PreflightCheckReq) (any, error) {
			return FileExists(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}),
	)

	server.IGet("/open/api/file/parent",
		func(c *gin.Context, rail common.Rail, req FetchParentFileReq) (any, error) {
			if req.FileKey == "" {
				return nil, common.NewWebErr("fileKey is required")
			}
			pf, e := FindParentFile(rail, mysql.GetConn(), req, server.ExtractUser(c))
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
		func(c *gin.Context, rail common.Rail, req MoveIntoDirReq) (any, error) {
			return nil, MoveFileToDir(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/make-dir",
		func(c *gin.Context, rail common.Rail, req MakeDirReq) (any, error) {
			return MakeDir(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User make directory", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/grant-access",
		func(c *gin.Context, rail common.Rail, req GrantAccessReq) (any, error) {
			uid, e := FindUserId(rail, req.GrantedTo)
			if e != nil {
				rail.Warnf("Unable to find user id, grantedTo: %s, %v", req.GrantedTo, e)
				return nil, common.NewWebErr("Failed to find user")
			}
			return nil, GranteFileAccess(rail, mysql.GetConn(), uid, req.FileId, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list-granted-access",
		func(c *gin.Context, rail common.Rail, req ListGrantedAccessReq) (any, error) {
			return ListGrantedFileAccess(rail, mysql.GetConn(), req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/remove-granted-access",
		func(c *gin.Context, rail common.Rail, req RemoveGrantedAccessReq) (any, error) {
			return nil, RemoveGrantedFileAccess(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/dir/list",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return ListDirs(rail, mysql.GetConn(), server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list",
		func(c *gin.Context, rail common.Rail, req ListFileReq) (any, error) {
			return ListFiles(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/delete",
		func(c *gin.Context, rail common.Rail, req DeleteFileReq) (any, error) {
			return nil, DeleteFile(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/create",
		func(c *gin.Context, rail common.Rail, req CreateFileReq) (any, error) {
			return nil, CreateFile(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/info/update",
		func(c *gin.Context, rail common.Rail, req UpdateFileReq) (any, error) {
			return nil, UpdateFile(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/tag/list/all",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return ListAllTags(rail, mysql.GetConn(), server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag/list-for-file",
		func(c *gin.Context, rail common.Rail, req ListFileTagReq) (any, error) {
			return ListFileTags(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag",
		func(c *gin.Context, rail common.Rail, req TagFileReq) (any, error) {
			user := server.ExtractUser(c)
			return nil, TagFile(rail, mysql.GetConn(), req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/untag",
		func(c *gin.Context, rail common.Rail, req UntagFileReq) (any, error) {
			user := server.ExtractUser(c)
			return nil, UntagFile(rail, mysql.GetConn(), req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/vfolder/brief/owned",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return ListVFolderBrief(rail, mysql.GetConn(), server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/list",
		func(c *gin.Context, rail common.Rail, req ListVFolderReq) (any, error) {
			return ListVFolders(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/create",
		func(c *gin.Context, rail common.Rail, req CreateVFolderReq) (any, error) {
			return CreateVFolder(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/add",
		func(c *gin.Context, rail common.Rail, req AddFileToVfolderReq) (any, error) {
			return nil, AddFileToVFolder(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/remove",
		func(c *gin.Context, rail common.Rail, req RemoveFileFromVfolderReq) (any, error) {
			return nil, RemoveFileFromVFolder(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove file from virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/share",
		func(c *gin.Context, rail common.Rail, req ShareVfolderReq) (any, error) {
			sharedTo, e := FindUser(rail, FindUserReq{Username: &req.Username})
			if e != nil {
				rail.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
				return nil, common.NewWebErr("Failed to find user")
			}
			return nil, ShareVFolder(rail, mysql.GetConn(), sharedTo, req.FolderNo, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/access/remove",
		func(c *gin.Context, rail common.Rail, req RemoveGrantedFolderAccessReq) (any, error) {
			return nil, RemoveVFolderAccess(rail, mysql.GetConn(), req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/granted/list",
		func(c *gin.Context, rail common.Rail, req ListGrantedFolderAccessReq) (any, error) {
			user := server.ExtractUser(c)
			return ListGrantedFolderAccess(rail, mysql.GetConn(), req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/token/generate",
		func(c *gin.Context, rail common.Rail, req GenerateTempTokenReq) (any, error) {
			user := server.ExtractUser(c)
			return GenTempToken(rail, mysql.GetConn(), req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User generate temporary token", Code: MANAGE_FILE_CODE}),
	)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	server.IGet("/remote/user/file/indir/list",
		func(c *gin.Context, rail common.Rail, req ListFilesInDirReq) (any, error) {
			return ListFilesInDir(rail, mysql.GetConn(), req)
		})
	server.IGet("/remote/user/file/info",
		func(c *gin.Context, rail common.Rail, req FetchFileInfoReq) (any, error) {
			return FetchFileInfoInternal(rail, mysql.GetConn(), req)
		})
	server.IGet("/remote/user/file/owner/validation",
		func(c *gin.Context, rail common.Rail, req ValidateFileOwnerReq) (any, error) {
			return ValidateFileOwner(rail, mysql.GetConn(), req)
		})

	// ---------------------------------- endpoints used to compensate --------------------------------------

	server.Post("/compensate/image/compression",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return nil, CompensateImageCompression(rail, mysql.GetConn())
		})
	return nil
}

func PrepareEventBus(rail common.Rail) error {
	// declare event bus
	if err := bus.DeclareEventBus(comprImgProcEventBus); err != nil {
		return err
	}
	if err := bus.DeclareEventBus(addFantahseaDirGalleryImgEventBus); err != nil {
		return err
	}
	if err := bus.DeclareEventBus(notifyFantahseaFileDeletedEventBus); err != nil {
		return err
	}

	// subscribe to event bus
	bus.SubscribeEventBus(comprImgNotifyEventBus, 2, OnImageCompressed)
	bus.SubscribeEventBus(fileSavedEventBus, 2, OnFileSaved)
	bus.SubscribeEventBus(thumbnailUpdatedEventBus, 2, OnThumbnailUpdated)
	bus.SubscribeEventBus(fileLDeletedEventBus, 2, OnFileDeleted)
	return nil
}

func PrepareGoAuthReport(rail common.Rail) error {
	// report goauth resources and paths
	goauth.ReportResourcesOnBootstrapped(rail, []goauth.AddResourceReq{
		{Name: MANAGE_FILE_NAME, Code: MANAGE_FILE_CODE},
		{Name: ADMIN_FS_NAME, Code: ADMIN_FS_CODE},
	})
	goauth.ReportPathsOnBootstrapped(rail)
	return nil
}
