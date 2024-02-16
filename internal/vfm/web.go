package vfm

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/curtisnewbie/gocommon/auth"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

const (
	ManageFilesResource = "manage-files"
)

func RegisterHttpRoutes(rail miso.Rail) error {

	auth.ExposeResourceInfo([]auth.Resource{
		{Code: ManageFilesResource, Name: "Manage files"},
	})

	miso.GroupRoute("/open/api",
		miso.GroupRoute("/file",
			miso.IGet("/upload/duplication/preflight", DupPreflightCheckEp).
				Desc("Preflight check for duplicate file uploads").
				Resource(ManageFilesResource),

			miso.IGet("/parent", GetParentFileEp).
				Desc("User fetch parent file info").
				Resource(ManageFilesResource),

			miso.IPost("/move-to-dir", MoveFileToDirEp).
				Desc("User move files into directory").
				Resource(ManageFilesResource),

			miso.IPost("/make-dir", MakeDirEp).
				Desc("User make directory").
				Resource(ManageFilesResource),

			miso.Get("/dir/list", ListDirEp).
				Desc("User list directories").
				Resource(ManageFilesResource),

			miso.IPost("/list", ListFilesEp).
				Desc("User list files").
				Resource(ManageFilesResource),

			miso.IPost("/delete", DeleteFileEp).
				Desc("User delete file").
				Resource(ManageFilesResource),

			miso.IPost("/dir/truncate", TruncateDirEp).
				Desc("User delete truncate directory recursively").
				Resource(ManageFilesResource),

			miso.IPost("/delete/batch", BatchDeleteFileEp).
				Desc("User delete file in batch").
				Resource(ManageFilesResource),

			miso.IPost("/create", CreateFileEp).
				Desc("User create file").
				Resource(ManageFilesResource),

			miso.IPost("/info/update", UpdateFileEp).
				Desc("User update file").
				Resource(ManageFilesResource),

			miso.Get("/tag/list/all", ListAllFileTagsEp).
				Desc("User list all file tags").
				Resource(ManageFilesResource),

			miso.IPost("/tag/list-for-file", ListTagsOfFileEp).
				Desc("User list tags of file").
				Resource(ManageFilesResource),

			miso.IPost("/tag", TagFileEp).
				Desc("User tag file").
				Resource(ManageFilesResource),

			miso.IPost("/untag", UntagFileEp).
				Desc("User untag file").
				Resource(ManageFilesResource),

			miso.IPost("/token/generate", GenFileTknEp).
				Desc("User generate temporary token").
				Resource(ManageFilesResource),

			miso.IPost("/unpack", UnpackZipEp).
				Desc("User unpack zip").
				Resource(ManageFilesResource),

			miso.RawGet("/token/qrcode", GenFileTknQRCodeEp).
				Desc("User generate qrcode image for temporary token").
				DocQueryParam("token", "Generated temporary file key").
				Public(),
		),

		miso.GroupRoute("/vfolder",

			miso.Get("/brief/owned", ListVFolderBriefEp).
				Desc("User list virtual folder briefs").
				Resource(ManageFilesResource),

			miso.IPost("/list", ListVFoldersEp).
				Desc("User list virtual folders").
				Resource(ManageFilesResource),

			miso.IPost("/create", CreateVFolderEp).
				Desc("User create virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/file/add", VFolderAddFileEp).
				Desc("User add file to virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/file/remove", VFolderRemoveFileEp).
				Desc("User remove file from virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/share", ShareVFolderEp).
				Desc("Share access to virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/access/remove", RemoveVFolderAccessEp).
				Desc("Remove granted access to virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/granted/list", ListVFolderAccessEp).
				Desc("List granted access to virtual folder").
				Resource(ManageFilesResource),

			miso.IPost("/remove", RemoveVFolderEp).
				Desc("Remove virtual folder").
				Resource(ManageFilesResource),
		),

		miso.GroupRoute("/gallery",

			miso.Get("/brief/owned", ListGalleryBriefsEp).
				Desc("List owned gallery brief info").
				Resource(ManageFilesResource),

			miso.IPost("/new", CreateGalleryEp).
				Desc("Create new gallery").
				Resource(ManageFilesResource),

			miso.IPost("/update", UpdateGalleryEp).
				Desc("Update gallery").
				Resource(ManageFilesResource),

			miso.IPost("/delete", DeleteGalleryEp).
				Desc("Delete gallery").
				Resource(ManageFilesResource),

			miso.IPost("/list", ListGalleriesEp).
				Desc("List galleries").
				Resource(ManageFilesResource),

			miso.IPost("/access/grant", GranteGalleryAccessEp).
				Desc("Grant access to the galleries").
				Resource(ManageFilesResource),

			miso.IPost("/access/remove", RemoveGalleryAccessEp).
				Desc("Remove access to the galleries").
				Resource(ManageFilesResource),

			miso.IPost("/access/list", ListGalleryAccessEp).
				Desc("List granted access to the galleries").
				Resource(ManageFilesResource),

			miso.IPost("/images", ListGalleryImagesEp).
				Desc("List images of gallery").
				Resource(ManageFilesResource),

			miso.IPost("/image/transfer", TransferGalleryImageEp).
				Desc("Host selected images on gallery").
				Resource(ManageFilesResource),
		),
	)

	// ---------------------------------------------- internal endpoints ------------------------------------------

	miso.GroupRoute("/remote/user/file",
		miso.IGet("/indir/list", ListFilesInDirEp),
		miso.IGet("/info", FetchFileInfoItnEp),
		miso.IGet("/owner/validation", ValidateOwnerEp),
	)

	// ---------------------------------- endpoints used to compensate --------------------------------------

	miso.GroupRoute("/compensate",

		// Compensate thumbnail generations, those that are images/videos (guessed by names) are processed to generate thumbnails
		// curl -X POST "http://localhost:8086/compensate/thumbnail"
		miso.Post("/thumbnail",
			func(c *gin.Context, rail miso.Rail) (any, error) {
				return nil, CompensateThumbnail(rail, miso.GetMySQL())
			}).
			Desc("Compensate thumbnail generation"),

		// update file_info records that do not have uploader_no
		// curl -X POST "http://localhost:8086/compensate/file/uploaderno"
		miso.Post("/file/uploaderno",
			func(c *gin.Context, rail miso.Rail) (any, error) {
				return nil, CompensateFileUploaderNo(rail, miso.GetMySQL())
			}).
			Desc("Update file_info records that don't have uploader_no"),

		// curl -X POST "http://localhost:8086/compensate/dir/calculate-size"
		miso.Post("/dir/calculate-size",
			func(c *gin.Context, rail miso.Rail) (any, error) {
				return nil, ImMemBatchCalcDirSize(rail, miso.GetMySQL())
			}).
			Desc("Calculate size of all directories recursively"),
	)

	return nil
}

func DupPreflightCheckEp(c *gin.Context, rail miso.Rail, req PreflightCheckReq) (any, error) {
	return FileExists(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func GetParentFileEp(c *gin.Context, rail miso.Rail, req FetchParentFileReq) (any, error) {
	if req.FileKey == "" {
		return nil, miso.NewErrf("fileKey is required")
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
	return nil, DeleteFile(rail, miso.GetMySQL(), req, common.GetUser(rail), nil)
}

func TruncateDirEp(c *gin.Context, rail miso.Rail, req DeleteFileReq) (any, error) {
	return nil, TruncateDir(rail, miso.GetMySQL(), req, common.GetUser(rail), true)
}

type BatchDeleteFileReq struct {
	FileKeys []string
}

func BatchDeleteFileEp(c *gin.Context, rail miso.Rail, req BatchDeleteFileReq) (any, error) {
	user := common.GetUser(rail)
	if len(req.FileKeys) < 31 {
		for i := range req.FileKeys {
			fk := req.FileKeys[i]
			if err := DeleteFile(rail, miso.GetMySQL(), DeleteFileReq{fk}, user, nil); err != nil {
				rail.Errorf("failed to delete file, fileKey: %v, %v", fk, err)
				return nil, err
			}
		}
		return nil, nil
	}

	// too many file keys, delete files asynchronously
	for i := range req.FileKeys {
		fk := req.FileKeys[i]
		vfmPool.Go(func() {
			rrail := rail.NextSpan()
			if err := DeleteFile(rrail, miso.GetMySQL(), DeleteFileReq{fk}, user, nil); err != nil {
				rrail.Errorf("failed to delete file, fileKey: %v, %v", fk, err)
			}
		})
	}
	return nil, nil
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
		return nil, miso.NewErrf("Failed to find user")
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

func ListGalleryBriefsEp(c *gin.Context, rail miso.Rail) (any, error) {
	user := common.GetUser(rail)
	return ListOwnedGalleryBriefs(rail, user, miso.GetMySQL())
}

func CreateGalleryEp(c *gin.Context, rail miso.Rail, cmd CreateGalleryCmd) (any, error) {
	user := common.GetUser(rail)
	return CreateGallery(rail, cmd, user, miso.GetMySQL())
}

func UpdateGalleryEp(c *gin.Context, rail miso.Rail, cmd UpdateGalleryCmd) (any, error) {
	user := common.GetUser(rail)
	e := UpdateGallery(rail, cmd, user, miso.GetMySQL())
	return nil, e
}

func DeleteGalleryEp(c *gin.Context, rail miso.Rail, cmd DeleteGalleryCmd) (any, error) {
	user := common.GetUser(rail)
	e := DeleteGallery(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func ListGalleriesEp(c *gin.Context, rail miso.Rail, cmd ListGalleriesCmd) (any, error) {
	user := common.GetUser(rail)
	return ListGalleries(rail, cmd, user, miso.GetMySQL())
}

func GranteGalleryAccessEp(c *gin.Context, rail miso.Rail, cmd PermitGalleryAccessCmd) (any, error) {
	user := common.GetUser(rail)
	e := GrantGalleryAccessToUser(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func RemoveGalleryAccessEp(c *gin.Context, rail miso.Rail, cmd RemoveGalleryAccessCmd) (any, error) {
	user := common.GetUser(rail)
	e := RemoveGalleryAccess(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func ListGalleryAccessEp(c *gin.Context, rail miso.Rail, cmd ListGrantedGalleryAccessCmd) (any, error) {
	user := common.GetUser(rail)
	return ListedGrantedGalleryAccess(rail, miso.GetMySQL(), cmd, user)
}

func ListGalleryImagesEp(c *gin.Context, rail miso.Rail, cmd ListGalleryImagesCmd) (any, error) {
	return ListGalleryImages(rail, miso.GetMySQL(), cmd, common.GetUser(rail))
}

func TransferGalleryImageEp(c *gin.Context, rail miso.Rail, cmd TransferGalleryImageReq) (any, error) {
	user := common.GetUser(rail)
	return BatchTransferAsync(rail, cmd, user, miso.GetMySQL())
}

func GenFileTknQRCodeEp(c *gin.Context, rail miso.Rail) {
	token := c.Query("token")
	if miso.IsBlankStr(token) {
		c.String(http.StatusBadRequest, "token is required")
		return
	}

	url := miso.GetPropStr(PropVfmSiteHost) + "/fstore/file/raw?key=" + url.QueryEscape(token)
	png, err := qrcode.Encode(url, qrcode.Medium, 512)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		rail.Errorf("Failed to generate qrcode image, fileKey: %v, %v", token, err)
		return
	}

	reader := bytes.NewReader(png)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		rail.Errorf("Failed to tranfer qrcode image, fileKey: %v, %v", token, err)
		return
	}

	c.Status(http.StatusOK)
}
