package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/miso/miso"
	"github.com/gin-gonic/gin"
)

func RegisterHttpRoutes(rail miso.Rail) error {
	miso.IGet("/open/api/file/upload/duplication/preflight",
		func(c *gin.Context, rail miso.Rail, req PreflightCheckReq) (any, error) {
			return FileExists(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User - preflight check for duplicate file uploads", MANAGE_FILE_CODE),
	)

	miso.IGet("/open/api/file/parent",
		func(c *gin.Context, rail miso.Rail, req FetchParentFileReq) (any, error) {
			if req.FileKey == "" {
				return nil, miso.NewErr("fileKey is required")
			}
			pf, e := FindParentFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
			if e != nil {
				return nil, e
			}
			if pf.Zero {
				return nil, nil
			}
			return pf, nil
		},
		goauth.Protected("User fetch parent file info", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/move-to-dir",
		func(c *gin.Context, rail miso.Rail, req MoveIntoDirReq) (any, error) {
			return nil, MoveFileToDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User move files into directory", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/make-dir",
		func(c *gin.Context, rail miso.Rail, req MakeDirReq) (any, error) {
			return MakeDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User make directory", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/grant-access",
		func(c *gin.Context, rail miso.Rail, req GrantAccessReq) (any, error) {
			uid, e := FindUserId(rail, req.GrantedTo)
			if e != nil {
				rail.Warnf("Unable to find user id, grantedTo: %s, %v", req.GrantedTo, e)
				return nil, miso.NewErr("Failed to find user")
			}
			return nil, GranteFileAccess(rail, miso.GetMySQL(), uid, req.FileId, common.GetUser(rail))
		},
		goauth.Protected("User grant file access", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/list-granted-access",
		func(c *gin.Context, rail miso.Rail, req ListGrantedAccessReq) (any, error) {
			return ListGrantedFileAccess(rail, miso.GetMySQL(), req)
		},
		goauth.Protected("User list granted file access", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/remove-granted-access",
		func(c *gin.Context, rail miso.Rail, req RemoveGrantedAccessReq) (any, error) {
			return nil, RemoveGrantedFileAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User remove granted file access", MANAGE_FILE_CODE),
	)

	miso.Get("/open/api/file/dir/list",
		func(c *gin.Context, rail miso.Rail) (any, error) {
			return ListDirs(rail, miso.GetMySQL(), common.GetUser(rail))
		},
		goauth.Protected("User list directories", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/list",
		func(c *gin.Context, rail miso.Rail, req ListFileReq) (any, error) {
			return ListFiles(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User list files", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/delete",
		func(c *gin.Context, rail miso.Rail, req DeleteFileReq) (any, error) {
			return nil, DeleteFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User delete file", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/create",
		func(c *gin.Context, rail miso.Rail, req CreateFileReq) (any, error) {
			return nil, CreateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User create file", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/info/update",
		func(c *gin.Context, rail miso.Rail, req UpdateFileReq) (any, error) {
			return nil, UpdateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User update file", MANAGE_FILE_CODE),
	)

	miso.Get("/open/api/file/tag/list/all",
		func(c *gin.Context, rail miso.Rail) (any, error) {
			return ListAllTags(rail, miso.GetMySQL(), common.GetUser(rail))
		},
		goauth.Protected("User list all file tags", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/tag/list-for-file",
		func(c *gin.Context, rail miso.Rail, req ListFileTagReq) (any, error) {
			return ListFileTags(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User list tags of file", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/tag",
		func(c *gin.Context, rail miso.Rail, req TagFileReq) (any, error) {
			user := common.GetUser(rail)
			return nil, TagFile(rail, miso.GetMySQL(), req, user)
		},
		goauth.Protected("User tag file", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/untag",
		func(c *gin.Context, rail miso.Rail, req UntagFileReq) (any, error) {
			user := common.GetUser(rail)
			return nil, UntagFile(rail, miso.GetMySQL(), req, user)
		},
		goauth.Protected("User untag file", MANAGE_FILE_CODE),
	)

	miso.Get("/open/api/vfolder/brief/owned",
		func(c *gin.Context, rail miso.Rail) (any, error) {
			return ListVFolderBrief(rail, miso.GetMySQL(), common.GetUser(rail))
		},
		goauth.Protected("User list virtual folder briefs", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/list",
		func(c *gin.Context, rail miso.Rail, req ListVFolderReq) (any, error) {
			return ListVFolders(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User list virtual folders", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/create",
		func(c *gin.Context, rail miso.Rail, req CreateVFolderReq) (any, error) {
			return CreateVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User create virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/file/add",
		func(c *gin.Context, rail miso.Rail, req AddFileToVfolderReq) (any, error) {
			return nil, AddFileToVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User add file to virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/file/remove",
		func(c *gin.Context, rail miso.Rail, req RemoveFileFromVfolderReq) (any, error) {
			return nil, RemoveFileFromVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User remove file from virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/share",
		func(c *gin.Context, rail miso.Rail, req ShareVfolderReq) (any, error) {
			sharedTo, e := FindUser(rail, FindUserReq{Username: &req.Username})
			if e != nil {
				rail.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
				return nil, miso.NewErr("Failed to find user")
			}
			return nil, ShareVFolder(rail, miso.GetMySQL(), sharedTo, req.FolderNo, common.GetUser(rail))
		},
		goauth.Protected("Share access to virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/access/remove",
		func(c *gin.Context, rail miso.Rail, req RemoveGrantedFolderAccessReq) (any, error) {
			return nil, RemoveVFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("Remove granted access to virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/vfolder/granted/list",
		func(c *gin.Context, rail miso.Rail, req ListGrantedFolderAccessReq) (any, error) {
			return ListGrantedFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("List granted access to virtual folder", MANAGE_FILE_CODE),
	)

	miso.IPost("/open/api/file/token/generate",
		func(c *gin.Context, rail miso.Rail, req GenerateTempTokenReq) (any, error) {
			return GenTempToken(rail, miso.GetMySQL(), req, common.GetUser(rail))
		},
		goauth.Protected("User generate temporary token", MANAGE_FILE_CODE),
	)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	miso.IGet("/remote/user/file/indir/list",
		func(c *gin.Context, rail miso.Rail, req ListFilesInDirReq) (any, error) {
			return ListFilesInDir(rail, miso.GetMySQL(), req)
		})
	miso.IGet("/remote/user/file/info",
		func(c *gin.Context, rail miso.Rail, req FetchFileInfoReq) (any, error) {
			return FetchFileInfoInternal(rail, miso.GetMySQL(), req)
		})
	miso.IGet("/remote/user/file/owner/validation",
		func(c *gin.Context, rail miso.Rail, req ValidateFileOwnerReq) (any, error) {
			return ValidateFileOwner(rail, miso.GetMySQL(), req)
		})

	// ---------------------------------- endpoints used to compensate --------------------------------------

	miso.Post("/compensate/image/compression",
		func(c *gin.Context, rail miso.Rail) (any, error) {
			return nil, CompensateImageCompression(rail, miso.GetMySQL())
		})
	return nil
}
