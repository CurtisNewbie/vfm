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
			func(inb *miso.Inbound) (any, error) {
				return nil, CompensateThumbnail(rail, miso.GetMySQL())
			}).
			Desc("Compensate thumbnail generation"),

		// curl -X POST "http://localhost:8086/compensate/dir/calculate-size"
		miso.Post("/dir/calculate-size",
			func(inb *miso.Inbound) (any, error) {
				return nil, ImMemBatchCalcDirSize(rail, miso.GetMySQL())
			}).
			Desc("Calculate size of all directories recursively"),
	)

	return nil
}

func DupPreflightCheckEp(inb *miso.Inbound, req PreflightCheckReq) (bool, error) {
	rail := inb.Rail()
	return FileExists(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func GetParentFileEp(inb *miso.Inbound, req FetchParentFileReq) (*ParentFileInfo, error) {
	rail := inb.Rail()
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
	return &pf, nil
}

func MoveFileToDirEp(inb *miso.Inbound, req MoveIntoDirReq) (any, error) {
	rail := inb.Rail()
	return nil, MoveFileToDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func MakeDirEp(inb *miso.Inbound, req MakeDirReq) (string, error) {
	rail := inb.Rail()
	return MakeDir(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListDirEp(inb *miso.Inbound) ([]ListedDir, error) {
	rail := inb.Rail()
	return ListDirs(rail, miso.GetMySQL(), common.GetUser(rail))
}

func ListFilesEp(inb *miso.Inbound, req ListFileReq) (miso.PageRes[ListedFile], error) {
	rail := inb.Rail()
	return ListFiles(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func DeleteFileEp(inb *miso.Inbound, req DeleteFileReq) (any, error) {
	rail := inb.Rail()
	return nil, DeleteFile(rail, miso.GetMySQL(), req, common.GetUser(rail), nil)
}

func TruncateDirEp(inb *miso.Inbound, req DeleteFileReq) (any, error) {
	rail := inb.Rail()
	return nil, TruncateDir(rail, miso.GetMySQL(), req, common.GetUser(rail), true)
}

type BatchDeleteFileReq struct {
	FileKeys []string
}

func BatchDeleteFileEp(inb *miso.Inbound, req BatchDeleteFileReq) (any, error) {
	rail := inb.Rail()
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

func CreateFileEp(inb *miso.Inbound, req CreateFileReq) (any, error) {
	rail := inb.Rail()
	return nil, CreateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func UpdateFileEp(inb *miso.Inbound, req UpdateFileReq) (any, error) {
	rail := inb.Rail()
	return nil, UpdateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func GenFileTknEp(inb *miso.Inbound, req GenerateTempTokenReq) (string, error) {
	rail := inb.Rail()
	return GenTempToken(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListVFolderBriefEp(inb *miso.Inbound) ([]VFolderBrief, error) {
	rail := inb.Rail()
	return ListVFolderBrief(rail, miso.GetMySQL(), common.GetUser(rail))
}

func ListVFoldersEp(inb *miso.Inbound, req ListVFolderReq) (ListVFolderRes, error) {
	rail := inb.Rail()
	return ListVFolders(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func CreateVFolderEp(inb *miso.Inbound, req CreateVFolderReq) (string, error) {
	rail := inb.Rail()
	return CreateVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func VFolderAddFileEp(inb *miso.Inbound, req AddFileToVfolderReq) (any, error) {
	rail := inb.Rail()
	return nil, AddFileToVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func VFolderRemoveFileEp(inb *miso.Inbound, req RemoveFileFromVfolderReq) (any, error) {
	rail := inb.Rail()
	return nil, RemoveFileFromVFolder(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ShareVFolderEp(inb *miso.Inbound, req ShareVfolderReq) (any, error) {
	rail := inb.Rail()
	sharedTo, e := vault.FindUser(rail, vault.FindUserReq{Username: &req.Username})
	if e != nil {
		rail.Warnf("Unable to find user, sharedTo: %s, %v", req.Username, e)
		return nil, miso.NewErrf("Failed to find user")
	}
	return nil, ShareVFolder(rail, miso.GetMySQL(), sharedTo, req.FolderNo, common.GetUser(rail))
}

func RemoveVFolderAccessEp(inb *miso.Inbound, req RemoveGrantedFolderAccessReq) (any, error) {
	rail := inb.Rail()
	return nil, RemoveVFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListVFolderAccessEp(inb *miso.Inbound, req ListGrantedFolderAccessReq) (ListGrantedFolderAccessRes, error) {
	rail := inb.Rail()
	return ListGrantedFolderAccess(rail, miso.GetMySQL(), req, common.GetUser(rail))
}

func ListFilesInDirEp(inb *miso.Inbound, req ListFilesInDirReq) ([]string, error) {
	rail := inb.Rail()
	return ListFilesInDir(rail, miso.GetMySQL(), req)
}

func FetchFileInfoItnEp(inb *miso.Inbound, req FetchFileInfoReq) (FileInfoResp, error) {
	rail := inb.Rail()
	return FetchFileInfoInternal(rail, miso.GetMySQL(), req)
}

func ValidateOwnerEp(inb *miso.Inbound, req ValidateFileOwnerReq) (bool, error) {
	rail := inb.Rail()
	return ValidateFileOwner(rail, miso.GetMySQL(), req)
}

func RemoveVFolderEp(inb *miso.Inbound, req RemoveVFolderReq) (any, error) {
	rail := inb.Rail()
	return nil, RemoveVFolder(rail, miso.GetMySQL(), common.GetUser(rail), req)
}

func UnpackZipEp(inb *miso.Inbound, req UnpackZipReq) (any, error) {
	rail := inb.Rail()
	err := UnpackZip(rail, miso.GetMySQL(), common.GetUser(rail), req)
	return nil, err
}

func ListGalleryBriefsEp(inb *miso.Inbound) ([]VGalleryBrief, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return ListOwnedGalleryBriefs(rail, user, miso.GetMySQL())
}

func CreateGalleryEp(inb *miso.Inbound, cmd CreateGalleryCmd) (*Gallery, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return CreateGallery(rail, cmd, user, miso.GetMySQL())
}

func UpdateGalleryEp(inb *miso.Inbound, cmd UpdateGalleryCmd) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	e := UpdateGallery(rail, cmd, user, miso.GetMySQL())
	return nil, e
}

func DeleteGalleryEp(inb *miso.Inbound, cmd DeleteGalleryCmd) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	e := DeleteGallery(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func ListGalleriesEp(inb *miso.Inbound, cmd ListGalleriesCmd) (miso.PageRes[VGallery], error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return ListGalleries(rail, cmd, user, miso.GetMySQL())
}

func GranteGalleryAccessEp(inb *miso.Inbound, cmd PermitGalleryAccessCmd) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	e := GrantGalleryAccessToUser(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func RemoveGalleryAccessEp(inb *miso.Inbound, cmd RemoveGalleryAccessCmd) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	e := RemoveGalleryAccess(rail, miso.GetMySQL(), cmd, user)
	return nil, e
}

func ListGalleryAccessEp(inb *miso.Inbound, cmd ListGrantedGalleryAccessCmd) (miso.PageRes[ListedGalleryAccessRes], error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return ListedGrantedGalleryAccess(rail, miso.GetMySQL(), cmd, user)
}

func ListGalleryImagesEp(inb *miso.Inbound, cmd ListGalleryImagesCmd) (*ListGalleryImagesResp, error) {
	rail := inb.Rail()
	return ListGalleryImages(rail, miso.GetMySQL(), cmd, common.GetUser(rail))
}

func TransferGalleryImageEp(inb *miso.Inbound, cmd TransferGalleryImageReq) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return BatchTransferAsync(rail, cmd, user, miso.GetMySQL())
}

func GenFileTknQRCodeEp(inb *miso.Inbound) {
	w, r := inb.Unwrap()
	rail := inb.Rail()
	token := r.URL.Query().Get("token")
	if miso.IsBlankStr(token) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := miso.GetPropStr(PropVfmSiteHost) + "/fstore/file/raw?key=" + url.QueryEscape(token)
	png, err := qrcode.Encode(url, qrcode.Medium, 512)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rail.Errorf("Failed to generate qrcode image, fileKey: %v, %v", token, err)
		return
	}

	reader := bytes.NewReader(png)
	_, err = io.Copy(w, reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rail.Errorf("Failed to tranfer qrcode image, fileKey: %v, %v", token, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
