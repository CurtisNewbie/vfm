package vfm

import (
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/goauth"
	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
	"github.com/gin-gonic/gin"
)

func RegisterHttpRoutes(rail miso.Rail) error {

	miso.BaseRoute("/open/api").With(
		miso.SubPath("/file").Group(
			miso.IGet("/upload/duplication/preflight", DupPreflightCheckEp).
				Extra(goauth.Protected("Preflight check for duplicate file uploads", ManageFileResCode)),

			miso.IGet("/parent", GetParentFileEp).
				Extra(goauth.Protected("User fetch parent file info", ManageFileResCode)),

			miso.IPost("/move-to-dir", MoveFileToDirEp).
				Extra(goauth.Protected("User move files into directory", ManageFileResCode)),

			miso.IPost("/make-dir", MakeDirEp).
				Extra(goauth.Protected("User make directory", ManageFileResCode)),

			miso.Get("/dir/list", ListDirEp).
				Extra(goauth.Protected("User list directories", ManageFileResCode)),

			miso.IPost("/list", ListFilesEp).
				Extra(goauth.Protected("User list files", ManageFileResCode)),

			miso.IPost("/delete", DeleteFileEp).
				Extra(goauth.Protected("User delete file", ManageFileResCode)),

			miso.IPost("/create", CreateFileEp).
				Extra(goauth.Protected("User create file", ManageFileResCode)),

			miso.IPost("/info/update", UpdateFileEp).
				Extra(goauth.Protected("User update file", ManageFileResCode)),

			miso.Get("/tag/list/all", ListAllFileTagsEp).
				Extra(goauth.Protected("User list all file tags", ManageFileResCode)),

			miso.IPost("/tag/list-for-file", ListTagsOfFileEp).
				Extra(goauth.Protected("User list tags of file", ManageFileResCode)),

			miso.IPost("/tag", TagFileEp).
				Extra(goauth.Protected("User tag file", ManageFileResCode)),

			miso.IPost("/untag", UntagFileEp).
				Extra(goauth.Protected("User untag file", ManageFileResCode)),

			miso.IPost("/token/generate", GenFileTknEp).
				Extra(goauth.Protected("User generate temporary token", ManageFileResCode)),

			miso.IPost("/unpack", UnpackZipEp).
				Extra(goauth.Protected("User unpack zip", ManageFileResCode)),
		),

		miso.SubPath("/vfolder").Group(
			miso.Get("/brief/owned", ListVFolderBriefEp).
				Extra(goauth.Protected("User list virtual folder briefs", ManageFileResCode)),

			miso.IPost("/list", ListVFoldersEp).
				Extra(goauth.Protected("User list virtual folders", ManageFileResCode)),

			miso.IPost("/create", CreateVFolderEp).
				Extra(goauth.Protected("User create virtual folder", ManageFileResCode)),

			miso.IPost("/file/add", VFolderAddFileEp).
				Extra(goauth.Protected("User add file to virtual folder", ManageFileResCode)),

			miso.IPost("/file/remove", VFolderRemoveFileEp).
				Extra(goauth.Protected("User remove file from virtual folder", ManageFileResCode)),

			miso.IPost("/share", ShareVFolderEp).
				Extra(goauth.Protected("Share access to virtual folder", ManageFileResCode)),

			miso.IPost("/access/remove", RemoveVFolderAccessEp).
				Extra(goauth.Protected("Remove granted access to virtual folder", ManageFileResCode)),

			miso.IPost("/granted/list", ListVFolderAccessEp).
				Extra(goauth.Protected("List granted access to virtual folder", ManageFileResCode)),

			miso.IPost("/remove", RemoveVFolderEp).
				Extra(goauth.Protected("Remove virtual folder", ManageFileResCode)),
		),
		miso.SubPath("/gallery").Group(
			miso.Get("/brief/owned",
				func(c *gin.Context, rail miso.Rail) (any, error) {
					user := common.GetUser(rail)
					return ListOwnedGalleryBriefs(rail, user, miso.GetMySQL())
				}).
				Extra(goauth.Protected("List owned gallery brief info", ManageFileResCode)),

			miso.IPost("/new",
				func(c *gin.Context, rail miso.Rail, cmd CreateGalleryCmd) (any, error) {
					user := common.GetUser(rail)
					return CreateGallery(rail, cmd, user, miso.GetMySQL())
				}).
				Extra(goauth.Protected("Create new gallery", ManageFileResCode)),

			miso.IPost("/update",
				func(c *gin.Context, rail miso.Rail, cmd UpdateGalleryCmd) (any, error) {
					user := common.GetUser(rail)
					e := UpdateGallery(rail, cmd, user, miso.GetMySQL())
					return nil, e
				}).
				Extra(goauth.Protected("Update gallery", ManageFileResCode)),

			miso.IPost("/delete",
				func(c *gin.Context, rail miso.Rail, cmd DeleteGalleryCmd) (any, error) {
					user := common.GetUser(rail)
					e := DeleteGallery(rail, miso.GetMySQL(), cmd, user)
					return nil, e
				}).
				Extra(goauth.Protected("Delete gallery", ManageFileResCode)),

			miso.IPost("/list",
				func(c *gin.Context, rail miso.Rail, cmd ListGalleriesCmd) (any, error) {
					user := common.GetUser(rail)
					return ListGalleries(rail, cmd, user, miso.GetMySQL())
				}).
				Extra(goauth.Protected("List galleries", ManageFileResCode)),

			miso.IPost("/access/grant",
				func(c *gin.Context, rail miso.Rail, cmd PermitGalleryAccessCmd) (any, error) {
					user := common.GetUser(rail)
					e := GrantGalleryAccessToUser(rail, miso.GetMySQL(), cmd, user)
					return nil, e
				}).
				Extra(goauth.Protected("Grant access to the galleries", ManageFileResCode)),

			miso.IPost("/access/remove",
				func(c *gin.Context, rail miso.Rail, cmd RemoveGalleryAccessCmd) (any, error) {
					user := common.GetUser(rail)
					e := RemoveGalleryAccess(rail, miso.GetMySQL(), cmd, user)
					return nil, e
				}).
				Extra(goauth.Protected("Grant access to the galleries", ManageFileResCode)),

			miso.IPost("/access/list",
				func(c *gin.Context, rail miso.Rail, cmd ListGrantedGalleryAccessCmd) (any, error) {
					user := common.GetUser(rail)
					return ListedGrantedGalleryAccess(rail, miso.GetMySQL(), cmd, user)
				}).
				Extra(goauth.Protected("List granted access to the galleries", ManageFileResCode)),

			miso.IPost("/images",
				func(c *gin.Context, rail miso.Rail, cmd ListGalleryImagesCmd) (any, error) {
					return ListGalleryImages(rail, miso.GetMySQL(), cmd, common.GetUser(rail))
				}).
				Extra(goauth.Protected("List images of gallery", ManageFileResCode)),

			miso.IPost("/image/transfer",
				func(c *gin.Context, rail miso.Rail, cmd TransferGalleryImageReq) (any, error) {
					user := common.GetUser(rail)
					return BatchTransferAsync(rail, cmd, user, miso.GetMySQL())
				}).
				Extra(goauth.Protected("Host selected images on gallery", ManageFileResCode)),
		),
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
	sharedTo, e := vault.FindUser(rail, vault.FindUserReq{Username: &req.Username})
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

func RemoveVFolderEp(c *gin.Context, rail miso.Rail, req RemoveVFolderReq) (any, error) {
	return nil, RemoveVFolder(rail, miso.GetMySQL(), common.GetUser(rail), req)
}

func UnpackZipEp(c *gin.Context, rail miso.Rail, req UnpackZipReq) (any, error) {
	err := UnpackZip(rail, miso.GetMySQL(), common.GetUser(rail), req)
	return nil, err
}
