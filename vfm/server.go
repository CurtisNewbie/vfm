package vfm

import (
	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/gocommon/server"
	"github.com/gin-gonic/gin"
)

func PrepareServer(rail common.Rail) error {
	if goauth.IsEnabled() {
		server.PostServerBootstrapped(func(sc common.Rail) error {
			c := common.EmptyRail()
			if e := goauth.AddResourceAsync(rail, goauth.AddResourceReq{Name: MANAGE_FILE_NAME, Code: MANAGE_FILE_CODE}); e != nil {
				c.Errorf("Failed to create goauth resource, %v", e)
			}

			if e := goauth.AddResourceAsync(rail, goauth.AddResourceReq{Name: ADMIN_FS_NAME, Code: ADMIN_FS_CODE}); e != nil {
				c.Errorf("Failed to create goauth resource, %v", e)
			}
			return nil
		})
		goauth.ReportPathsOnBootstrapped()
	}

	bus.DeclareEventBus(comprImgProcEventBus)
	bus.DeclareEventBus(addFantahseaDirGalleryImgEventBus)
	bus.DeclareEventBus(notifyFantahseaFileDeletedEventBus)

	bus.SubscribeEventBus(comprImgNotifyEventBus, 2, OnImageCompressed)
	bus.SubscribeEventBus(fileSavedEventBus, 2, OnFileSaved)
	bus.SubscribeEventBus(thumbnailUpdatedEventBus, 2, OnThumbnailUpdated)
	bus.SubscribeEventBus(fileLDeletedEventBus, 2, OnFileDeleted)

	server.Get("/open/api/file/upload/duplication/preflight",
		func(c *gin.Context, ec common.Rail) (any, error) {
			filename := c.Query("fileName")
			parentFileKey := c.Query("parentFileKey")
			ec.Debugf("filename: %v, parentFileKey: %v", filename, parentFileKey)
			return FileExists(ec, filename, parentFileKey, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User - preflight check for duplicate file uploads", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/parent",
		func(c *gin.Context, rail common.Rail) (any, error) {
			fk := c.Query("fileKey")
			if fk == "" {
				return nil, common.NewWebErr("fileKey is required")
			}
			rail.Debugf("fileKey: %v", fk)
			pf, e := FindParentFile(rail, fk, server.ExtractUser(c))
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
			return nil, MoveFileToDir(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User move files into directory", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/make-dir",
		func(c *gin.Context, rail common.Rail, req MakeDirReq) (any, error) {
			return MakeDir(rail, req, server.ExtractUser(c))
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
			return nil, GranteFileAccess(rail, uid, req.FileId, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User grant file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list-granted-access",
		func(c *gin.Context, rail common.Rail, req ListGrantedAccessReq) (any, error) {
			return ListGrantedFileAccess(rail, req)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/remove-granted-access",
		func(c *gin.Context, rail common.Rail, req RemoveGrantedAccessReq) (any, error) {
			return nil, RemoveGrantedFileAccess(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User remove granted file access", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/dir/list",
		func(c *gin.Context, ec common.Rail) (any, error) {
			return ListDirs(ec, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list directories", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/list",
		func(c *gin.Context, rail common.Rail, req ListFileReq) (any, error) {
			return ListFiles(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list files", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/delete",
		func(c *gin.Context, rail common.Rail, req DeleteFileReq) (any, error) {
			return nil, DeleteFile(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User delete file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/create",
		func(c *gin.Context, rail common.Rail, req CreateFileReq) (any, error) {
			return nil, CreateFile(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/info/update",
		func(c *gin.Context, rail common.Rail, req UpdateFileReq) (any, error) {
			return nil, UpdateFile(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User update file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/file/tag/list/all",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return ListAllTags(rail, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list all file tags", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag/list-for-file",
		func(c *gin.Context, rail common.Rail, req ListFileTagReq) (any, error) {
			return ListFileTags(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list tags of file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/tag",
		func(c *gin.Context, rail common.Rail, req TagFileReq) (any, error) {
			user := server.ExtractUser(c)
			return nil, TagFile(rail, req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User tag file", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/untag",
		func(c *gin.Context, rail common.Rail, req UntagFileReq) (any, error) {
			user := server.ExtractUser(c)
			return nil, UntagFile(rail, req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User untag file", Code: MANAGE_FILE_CODE}),
	)

	server.Get("/open/api/vfolder/brief/owned",
		func(c *gin.Context, rail common.Rail) (any, error) {
			return ListVFolderBrief(rail, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folder briefs", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/list",
		func(c *gin.Context, rail common.Rail, req ListVFolderReq) (any, error) {
			return ListVFolders(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User list virtual folders", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/create",
		func(c *gin.Context, rail common.Rail, req CreateVFolderReq) (any, error) {
			return CreateVFolder(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User create virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/add",
		func(c *gin.Context, rail common.Rail, req AddFileToVfolderReq) (any, error) {
			return nil, AddFileToVFolder(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User add file to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/file/remove",
		func(c *gin.Context, rail common.Rail, req RemoveFileFromVfolderReq) (any, error) {
			return nil, RemoveFileFromVFolder(rail, req, server.ExtractUser(c))
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
			return nil, ShareVFolder(rail, sharedTo, req.FolderNo, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Share access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/access/remove",
		func(c *gin.Context, rail common.Rail, req RemoveGrantedFolderAccessReq) (any, error) {
			return nil, RemoveVFolderAccess(rail, req, server.ExtractUser(c))
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "Remove granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/vfolder/granted/list",
		func(c *gin.Context, rail common.Rail, req ListGrantedFolderAccessReq) (any, error) {
			user := server.ExtractUser(c)
			return ListGrantedFolderAccess(rail, req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "List granted access to virtual folder", Code: MANAGE_FILE_CODE}),
	)

	server.IPost("/open/api/file/token/generate",
		func(c *gin.Context, rail common.Rail, req GenerateTempTokenReq) (any, error) {
			user := server.ExtractUser(c)
			return GenTempToken(rail, req, user)
		},
		goauth.PathDocExtra(goauth.PathDoc{Desc: "User generate temporary token", Code: MANAGE_FILE_CODE}),
	)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	server.IGet("/remote/user/file/indir/list", func(c *gin.Context, rail common.Rail, q ListFilesInDirReq) (any, error) {
		return ListFilesInDir(rail, q)
	})
	server.Get("/remote/user/file/info", func(c *gin.Context, rail common.Rail) (any, error) {
		return FetchFileInfoInternal(rail, c.Query("fileKey"))
	})
	server.IGet("/remote/user/file/owner/validation", func(c *gin.Context, rail common.Rail, q ValidateFileOwnerReq) (any, error) {
		return ValidateFileOwner(rail, q)
	})

	// ---------------------------------- endpoints used to compensate --------------------------------------

	server.Post("/compensate/image/compression", func(c *gin.Context, rail common.Rail) (any, error) {
		return nil, CompensateImageCompression(rail)
	})

	return nil
}
