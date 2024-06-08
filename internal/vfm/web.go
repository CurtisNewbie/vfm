package vfm

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/curtisnewbie/miso/middleware/user-vault/auth"
	"github.com/curtisnewbie/miso/middleware/user-vault/common"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util"
	vault "github.com/curtisnewbie/user-vault/api"
	"github.com/skip2/go-qrcode"
)

const (
	ManageFilesResource    = "manage-files"
	ResourceManageBookmark = "manage-bookmarks"
)

var (
	ErrUnknown     = miso.NewErrf("Unknown error, please try again")
	ErrUploadFiled = miso.NewErrf("Upload failed, please try again")
)

func RegisterHttpRoutes(rail miso.Rail) error {

	auth.ExposeResourceInfo([]auth.Resource{
		{Code: ManageFilesResource, Name: "Manage files"},
		{Code: ResourceManageBookmark, Name: "Manage Bookmarks"},
	})

	miso.BaseRoute("/open/api").Group(

		miso.IGet("/file/upload/duplication/preflight", DupPreflightCheckEp).
			Desc("Preflight check for duplicate file uploads").
			Resource(ManageFilesResource),

		miso.IGet("/file/parent", GetParentFileEp).
			Desc("User fetch parent file info").
			Resource(ManageFilesResource),

		miso.IPost("/file/move-to-dir", MoveFileToDirEp).
			Desc("User move files into directory").
			Resource(ManageFilesResource),

		miso.IPost("/file/make-dir", MakeDirEp).
			Desc("User make directory").
			Resource(ManageFilesResource),

		miso.Get("/file/dir/list", ListDirEp).
			Desc("User list directories").
			Resource(ManageFilesResource),

		miso.IPost("/file/list", ListFilesEp).
			Desc("User list files").
			Resource(ManageFilesResource),

		miso.IPost("/file/delete", DeleteFileEp).
			Desc("User delete file").
			Resource(ManageFilesResource),

		miso.IPost("/file/dir/truncate", TruncateDirEp).
			Desc("User delete truncate directory recursively").
			Resource(ManageFilesResource),

		miso.IPost("/file/delete/batch", BatchDeleteFileEp).
			Desc("User delete file in batch").
			Resource(ManageFilesResource),

		miso.IPost("/file/create", CreateFileEp).
			Desc("User create file").
			Resource(ManageFilesResource),

		miso.IPost("/file/info/update", UpdateFileEp).
			Desc("User update file").
			Resource(ManageFilesResource),

		miso.IPost("/file/token/generate", GenFileTknEp).
			Desc("User generate temporary token").
			Resource(ManageFilesResource),

		miso.IPost("/file/unpack", UnpackZipEp).
			Desc("User unpack zip").
			Resource(ManageFilesResource),

		miso.RawGet("/file/token/qrcode", GenFileTknQRCodeEp).
			Desc("User generate qrcode image for temporary token").
			DocQueryParam("token", "Generated temporary file key").
			Public(),

		miso.Get("/vfolder/brief/owned", ListVFolderBriefEp).
			Desc("User list virtual folder briefs").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/list", ListVFoldersEp).
			Desc("User list virtual folders").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/create", CreateVFolderEp).
			Desc("User create virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/file/add", VFolderAddFileEp).
			Desc("User add file to virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/file/remove", VFolderRemoveFileEp).
			Desc("User remove file from virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/share", ShareVFolderEp).
			Desc("Share access to virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/access/remove", RemoveVFolderAccessEp).
			Desc("Remove granted access to virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/granted/list", ListVFolderAccessEp).
			Desc("List granted access to virtual folder").
			Resource(ManageFilesResource),

		miso.IPost("/vfolder/remove", RemoveVFolderEp).
			Desc("Remove virtual folder").
			Resource(ManageFilesResource),

		miso.Get("/gallery/brief/owned", ListGalleryBriefsEp).
			Desc("List owned gallery brief info").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/new", CreateGalleryEp).
			Desc("Create new gallery").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/update", UpdateGalleryEp).
			Desc("Update gallery").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/delete", DeleteGalleryEp).
			Desc("Delete gallery").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/list", ListGalleriesEp).
			Desc("List galleries").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/access/grant", GranteGalleryAccessEp).
			Desc("Grant access to the galleries").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/access/remove", RemoveGalleryAccessEp).
			Desc("Remove access to the galleries").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/access/list", ListGalleryAccessEp).
			Desc("List granted access to the galleries").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/images", ListGalleryImagesEp).
			Desc("List images of gallery").
			Resource(ManageFilesResource),

		miso.IPost("/gallery/image/transfer", TransferGalleryImageEp).
			Desc("Host selected images on gallery").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/list", ApiListVersionedFile).
			Desc("List versioned files").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/history", ApiListVersionedFileHistory).
			Desc("List versioned file history").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/accumulated-size", ApiQryVersionedFileAccuSize).
			Desc("Query versioned file log accumulated size").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/create", ApiCreateVersionedFile).
			Desc("Create versioned file").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/update", ApiUpdateVersionedFile).
			Desc("Update versioned file").
			Resource(ManageFilesResource),

		miso.IPost("/versioned-file/delete", ApiDelVersionedFile).
			Desc("Delete versioned file").
			Resource(ManageFilesResource),
	)

	// --------- endpoints used to compensate -----------
	// Compensate thumbnail generations, those that are images/videos (guessed by names) are processed to generate thumbnails
	miso.Post("/compensate/thumbnail",
		func(inb *miso.Inbound) (any, error) {
			return nil, CompensateThumbnail(rail, miso.GetMySQL())
		}).
		Desc("Compensate thumbnail generation")

	miso.Post("/compensate/dir/calculate-size",
		func(inb *miso.Inbound) (any, error) {
			return nil, ImMemBatchCalcDirSize(rail, miso.GetMySQL())
		}).
		Desc("Calculate size of all directories recursively")

	// -------- endpoints for managing bookmarks and bookmark blacklist --------
	miso.Put("/bookmark/file/upload", UploadBookmarkFileEp).
		Desc("Upload bookmark file").
		Resource(ResourceManageBookmark)

	miso.IPost[ListBookmarksReq]("/bookmark/list", ListBookmarksEp).
		Desc("List bookmarks").
		Resource(ResourceManageBookmark)

	miso.IPost[RemoveBookmarkReq]("/bookmark/remove", RemoveBookmarkEp).
		Desc("Remove bookmark").
		Resource(ResourceManageBookmark)

	miso.IPost[ListBookmarksReq]("/bookmark/blacklist/list", ListBlacklistedBookmarksEp).
		Desc("List bookmark blacklist").
		Resource(ResourceManageBookmark)

	miso.IPost[RemoveBookmarkReq]("/bookmark/blacklist/remove", RemoveBookmarkBlacklistEp).
		Desc("Remove bookmark blacklist").
		Resource(ResourceManageBookmark)

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
	_, err := CreateFile(rail, miso.GetMySQL(), req, common.GetUser(rail))
	return nil, err
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
	if util.IsBlankStr(token) {
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

func ApiCreateVersionedFile(inb *miso.Inbound, req ApiCreateVerFileReq) (ApiCreateVerFileRes, error) {
	return CreateVerFile(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}

func ApiListVersionedFile(inb *miso.Inbound, req ApiListVerFileReq) (miso.PageRes[ApiListVerFileRes], error) {
	return ListVerFile(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}

func ApiUpdateVersionedFile(inb *miso.Inbound, req ApiUpdateVerFileReq) (any, error) {
	return nil, UpdateVerFile(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}

func ApiDelVersionedFile(inb *miso.Inbound, req ApiDelVerFileReq) (any, error) {
	return nil, DelVerFile(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}

type ListBookmarksReq struct {
	Name        *string
	Paging      miso.Paging
	Blacklisted bool `gorm:"-" json:"-"`
}

// Upload bookmark file endpoint.
func UploadBookmarkFileEp(inb *miso.Inbound) (any, error) {
	rail := inb.Rail()
	_, r := inb.Unwrap()
	user := common.GetUser(rail)
	path, err := TransferTmpFile(rail, r.Body)
	if err != nil {
		return nil, err
	}
	defer os.Remove(path)

	lock := miso.NewRLock(rail, "docindexer:bookmark:"+user.UserNo)
	if err := lock.Lock(); err != nil {
		rail.Errorf("failed to lock for bookmark upload, user: %v, %v", user.Username, err)
		return nil, miso.NewErrf("Please try again later")
	}
	defer lock.Unlock()

	if err := ProcessUploadedBookmarkFile(rail, path, user); err != nil {
		rail.Errorf("ProcessUploadedBookmarkFile failed, user: %v, path: %v, %v", user.Username, path, err)
		return nil, miso.NewErrf("Failed to parse bookmark file")
	}

	return nil, nil
}

// List bookmarks endpoint.
func ListBookmarksEp(inb *miso.Inbound, req ListBookmarksReq) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return ListBookmarks(rail, miso.GetMySQL(), req, user.UserNo)
}

type RemoveBookmarkReq struct {
	Id int64
}

// Remove bookmark endpoint.
func RemoveBookmarkEp(inb *miso.Inbound, req RemoveBookmarkReq) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return nil, RemoveBookmark(rail, miso.GetMySQL(), req.Id, user.UserNo)
}

func ListBlacklistedBookmarksEp(inb *miso.Inbound, req ListBookmarksReq) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	req.Blacklisted = true
	return ListBookmarks(rail, miso.GetMySQL(), req, user.UserNo)
}

func RemoveBookmarkBlacklistEp(inb *miso.Inbound, req RemoveBookmarkReq) (any, error) {
	rail := inb.Rail()
	user := common.GetUser(rail)
	return nil, RemoveBookmarkBlacklist(rail, miso.GetMySQL(), req.Id, user.UserNo)
}

type ApiListVerFileHistoryReq struct {
	Paging    miso.Paging `desc:"paging params"`
	VerFileId string      `desc:"versioned file id" valid:"notEmpty"`
}

type ApiListVerFileHistoryRes struct {
	Name        string     `desc:"file name"`
	FileKey     string     `desc:"file key"`
	SizeInBytes int64      `desc:"size in bytes"`
	UploadTime  util.ETime `desc:"last upload time"`
	Thumbnail   string     `desc:"thumbnail token"`
}

func ApiListVersionedFileHistory(inb *miso.Inbound, req ApiListVerFileHistoryReq) (miso.PageRes[ApiListVerFileHistoryRes], error) {
	return ListVerFileHistory(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}

type ApiQryVerFileAccuSizeReq struct {
	VerFileId string `desc:"versioned file id" valid:"notEmpty"`
}

type ApiQryVerFileAccuSizeRes struct {
	SizeInBytes int64 `desc:"total size in bytes"`
}

func ApiQryVersionedFileAccuSize(inb *miso.Inbound, req ApiQryVerFileAccuSizeReq) (ApiQryVerFileAccuSizeRes, error) {
	return CalcVerFileAccuSize(inb.Rail(), miso.GetMySQL(), req, common.GetUser(inb.Rail()))
}
