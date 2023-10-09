package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/miso/miso"
	"github.com/gin-gonic/gin"
)

func RegisterHttpRoutes(rail miso.Rail) error {

	miso.BaseRoute("/open/api/file").
		Group(
			miso.IGet("/upload/duplication/preflight",
				DupPreflightCheckEp,
				goauth.Protected("User - preflight check for duplicate file uploads", ManageFileResCode)),

			miso.IGet("/parent",
				GetParentFileEp,
				goauth.Protected("User fetch parent file info", ManageFileResCode)),

			miso.IPost("/move-to-dir",
				MoveFileToDirEp,
				goauth.Protected("User move files into directory", ManageFileResCode)),

			miso.IPost("/make-dir",
				MakeDirEp,
				goauth.Protected("User make directory", ManageFileResCode)),

			miso.Get("/dir/list",
				ListDirEp,
				goauth.Protected("User list directories", ManageFileResCode)),

			miso.IPost("/list",
				ListFilesEp,
				goauth.Protected("User list files", ManageFileResCode)),

			miso.IPost("/delete",
				DeleteFileEp,
				goauth.Protected("User delete file", ManageFileResCode)),

			miso.IPost("/create",
				CreateFileEp,
				goauth.Protected("User create file", ManageFileResCode)),

			miso.IPost("/info/update",
				UpdateFileEp,
				goauth.Protected("User update file", ManageFileResCode)),

			miso.Get("/tag/list/all",
				ListAllFileTagsEp,
				goauth.Protected("User list all file tags", ManageFileResCode)),

			miso.IPost("/tag/list-for-file",
				ListTagsOfFileEp,
				goauth.Protected("User list tags of file", ManageFileResCode)),

			miso.IPost("/tag",
				TagFileEp,
				goauth.Protected("User tag file", ManageFileResCode)),

			miso.IPost("/untag",
				UntagFileEp,
				goauth.Protected("User untag file", ManageFileResCode)),

			miso.IPost("/token/generate",
				GenFileTknEp,
				goauth.Protected("User generate temporary token", ManageFileResCode)),
		)

	miso.BaseRoute("/open/api/vfolder").
		Group(
			miso.Get("/brief/owned",
				ListVFolderBriefEp,
				goauth.Protected("User list virtual folder briefs", ManageFileResCode)),

			miso.IPost("/list",
				ListVFoldersEp,
				goauth.Protected("User list virtual folders", ManageFileResCode)),

			miso.IPost("/create",
				CreateVFolderEp,
				goauth.Protected("User create virtual folder", ManageFileResCode)),

			miso.IPost("/file/add",
				VFolderAddFileEp,
				goauth.Protected("User add file to virtual folder", ManageFileResCode)),

			miso.IPost("/file/remove",
				VFolderRemoveFileEp,
				goauth.Protected("User remove file from virtual folder", ManageFileResCode)),

			miso.IPost("/share",
				ShareVFolderEp,
				goauth.Protected("Share access to virtual folder", ManageFileResCode)),

			miso.IPost("/access/remove",
				RemoveVFolderAccessEp,
				goauth.Protected("Remove granted access to virtual folder", ManageFileResCode)),

			miso.IPost("/granted/list",
				ListVFolderAccessEp,
				goauth.Protected("List granted access to virtual folder", ManageFileResCode)),
		)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	miso.BaseRoute("/remote/user/file").
		Group(
			miso.IGet("/indir/list", ListFilesInDirEp),
			miso.IGet("/info", FetchFileInfoItnEp),
			miso.IGet("/owner/validation", ValidateOwnerEp),
		)

	// ---------------------------------- endpoints used to compensate --------------------------------------

	miso.BaseRoute("/compensate").
		Group(
			// Compensate image compressions, those that are images (guessed by names) are compressed to generate thumbnail
			miso.Post("/image/compression",
				func(c *gin.Context, rail miso.Rail) (any, error) {
					return nil, CompensateImageCompression(rail, miso.GetMySQL())
				}),

			// update file_info records that do not have uploader_no
			miso.Post("/file/uploaderno",
				func(c *gin.Context, rail miso.Rail) (any, error) {
					return nil, CompensateFileUploaderNo(rail, miso.GetMySQL())
				},
			),
		)

	return nil
}

func DupPreflightCheckEp(c *gin.Context, rail miso.Rail, req PreflightCheckReq) (any, error) {
	return FileExists(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func GetParentFileEp(c *gin.Context, rail miso.Rail, req FetchParentFileReq) (any, error) {
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
}

func MoveFileToDirEp(c *gin.Context, rail miso.Rail, req MoveIntoDirReq) (any, error) {
	return nil, MoveFileToDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func MakeDirEp(c *gin.Context, rail miso.Rail, req MakeDirReq) (any, error) {
	return MakeDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListDirEp(c *gin.Context, rail miso.Rail) (any, error) {
	return ListDirs(rail, miso.GetMySQL(), common.GetUser(rail))
}

func ListFilesEp(c *gin.Context, rail miso.Rail, req ListFileReq) (any, error) {
	return ListFiles(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func DeleteFileEp(c *gin.Context, rail miso.Rail, req DeleteFileReq) (any, error) {
	return nil, DeleteFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func CreateFileEp(c *gin.Context, rail miso.Rail, req CreateFileReq) (any, error) {
	return nil, CreateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func UpdateFileEp(c *gin.Context, rail miso.Rail, req UpdateFileReq) (any, error) {
	return nil, UpdateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListAllFileTagsEp(c *gin.Context, rail miso.Rail) (any, error) {
	return ListAllTags(rail, miso.GetMySQL(), common.GetUser(rail))
}

func ListTagsOfFileEp(c *gin.Context, rail miso.Rail, req ListFileTagReq) (any, error) {
	return ListFileTags(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func TagFileEp(c *gin.Context, rail miso.Rail, req TagFileReq) (any, error) {
	user := common.GetUser(rail)
	return nil, TagFile(rail, miso.GetMySQL(), req, user)
}

func UntagFileEp(c *gin.Context, rail miso.Rail, req UntagFileReq) (any, error) {
	user := common.GetUser(rail)
	return nil, UntagFile(rail, miso.GetMySQL(), req, user)
}

func GenFileTknEp(c *gin.Context, rail miso.Rail, req GenerateTempTokenReq) (any, error) {
	return GenTempToken(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListVFolderBriefEp(c *gin.Context, rail miso.Rail) (any, error) {
	return ListVFolderBrief(rail, miso.GetMySQL(), common.GetUser(rail))
}

func ListVFoldersEp(c *gin.Context, rail miso.Rail, req ListVFolderReq) (any, error) {
	return ListVFolders(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func CreateVFolderEp(c *gin.Context, rail miso.Rail, req CreateVFolderReq) (any, error) {
	return CreateVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func VFolderAddFileEp(c *gin.Context, rail miso.Rail, req AddFileToVfolderReq) (any, error) {
	return nil, AddFileToVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func VFolderRemoveFileEp(c *gin.Context, rail miso.Rail, req RemoveFileFromVfolderReq) (any, error) {
	return nil, RemoveFileFromVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ShareVFolderEp(c *gin.Context, rail miso.Rail, req ShareVfolderReq) (any, error) {
	sharedTo, e := FindUser(rail, FindUserReq{Username: &req.Username})
	if e != nil {
		rail.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
		return nil, miso.NewErr("Failed to find user")
	}
	return nil, ShareVFolder(rail, miso.GetMySQL(), sharedTo, req.FolderNo, common.GetUser(rail))
}

func RemoveVFolderAccessEp(c *gin.Context, rail miso.Rail, req RemoveGrantedFolderAccessReq) (any, error) {
	return nil, RemoveVFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListVFolderAccessEp(c *gin.Context, rail miso.Rail, req ListGrantedFolderAccessReq) (any, error) {
	return ListGrantedFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListFilesInDirEp(c *gin.Context, rail miso.Rail, req ListFilesInDirReq) (any, error) {
	return ListFilesInDir(rail, miso.GetMySQL(), req)
}

func FetchFileInfoItnEp(c *gin.Context, rail miso.Rail, req FetchFileInfoReq) (any, error) {
	return FetchFileInfoInternal(rail, miso.GetMySQL(), req)
}

func ValidateOwnerEp(c *gin.Context, rail miso.Rail, req ValidateFileOwnerReq) (any, error) {
	return ValidateFileOwner(rail, miso.GetMySQL(), req)
}
