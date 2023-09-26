package vfm

import (
	"fmt"
	"strings"
	"time"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

const (
	FILE_TYPE_FILE = "FILE" // file
	FILE_TYPE_DIR  = "DIR"  // directory

	FILE_LDEL_N = 0 // normal file
	FILE_LDEL_Y = 1 // file marked deleted

	FILE_PDEL_N = 0 // file marked deleted, the actual deletion is not yet processed
	FILE_PDEL_Y = 1 // file finally deleted, may be removed from disk or move to somewhere else

	VFOWNERSHIP_OWNER   = "OWNER"   // owner of the vfolder
	VFOWNERSHIP_GRANTED = "GRANTED" // granted access to the vfolder
)

var (
	_imageSuffix = miso.NewSet[string]()
)

func init() {
	_imageSuffix.AddAll([]string{"jpeg", "jpg", "gif", "png", "svg", "bmp", "webp", "apng", "avif"})
}

type CompressImageEvent struct {
	FileKey string // file key from vfm
	FileId  string // file id from mini-fstore
}

type FileTag struct {
	Id         int
	FileId     int
	TagId      int
	UserId     int
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

func (t FileTag) IsZero() bool {
	return t.Id < 1
}

type Tag struct {
	Id         int
	Name       string
	UserId     int
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

func (t Tag) IsZero() bool {
	return t.Id < 1
}

type FileVFolder struct {
	FolderNo   string
	Uuid       string
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

type VFolderBrief struct {
	FolderNo string `json:"folderNo"`
	Name     string `json:"name"`
}

type ListedDir struct {
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type ListedFile struct {
	Id             int        `json:"id"`
	Uuid           string     `json:"uuid"`
	Name           string     `json:"name"`
	UploadTime     miso.ETime `json:"uploadTime"`
	UploaderName   string     `json:"uploaderName"`
	SizeInBytes    int64      `json:"sizeInBytes"`
	IsOwner        bool       `json:"isOwner"`
	FileType       string     `json:"fileType"`
	UpdateTime     miso.ETime `json:"updateTime"`
	ParentFileName string     `json:"parentFileName"`
	ParentFile     string     `json:"-"`
	UploaderId     int        `json:"-"`
}

type GrantAccessReq struct {
	FileId    int    `json:"fileId" validation:"positive"`
	GrantedTo string `json:"grantedTo" validation:"notEmpty"`
}

type ListedFileTag struct {
	Id         int        `json:"id"`
	Name       string     `json:"name"`
	CreateTime miso.ETime `json:"createTime"`
	CreateBy   string     `json:"createBy"`
}

type ListedVFolder struct {
	Id         int        `json:"id"`
	FolderNo   string     `json:"folderNo"`
	Name       string     `json:"name"`
	CreateTime miso.ETime `json:"createTime"`
	CreateBy   string     `json:"createBy"`
	UpdateTime miso.ETime `json:"updateTime"`
	UpdateBy   string     `json:"updateBy"`
	Ownership  string     `json:"ownership"`
}

type ListVFolderRes struct {
	Page    miso.Paging     `json:"pagingVo"`
	Payload []ListedVFolder `json:"payload"`
}

type ShareVfolderReq struct {
	FolderNo string `json:"folderNo"`
	Username string `json:"username"`
}

type ListFilesRes struct {
	Page    miso.Paging  `json:"pagingVo"`
	Payload []ListedFile `json:"payload"`
}

type ParentFileInfo struct {
	Zero     bool   `json:"-"`
	FileKey  string `json:"fileKey"`
	Filename string `json:"fileName"`
}

type FileDownloadInfo struct {
	FileId         int
	Name           string
	UploaderId     int
	IsLogicDeleted int
	FileType       string
	FstoreFileId   string
}

type FileInfo struct {
	Id               int
	Name             string
	Uuid             string
	FstoreFileId     string
	Thumbnail        string // thumbnail is also a fstore's file_id
	IsLogicDeleted   int
	IsPhysicDeleted  int
	SizeInBytes      int64
	UploaderId       int
	UploaderName     string
	UploadTime       miso.ETime
	LogicDeleteTime  miso.ETime
	PhysicDeleteTime miso.ETime
	UserGroup        int
	FsGroupId        int
	FileType         string
	ParentFile       string
	CreateTime       miso.ETime
	CreateBy         string
	UpdateTime       miso.ETime
	UpdateBy         string
	IsDel            int
}

func (f FileInfo) IsZero() bool {
	return f.Id < 1
}

type VFolderWithOwnership struct {
	Id         int
	FolderNo   string
	Name       string
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
	Ownership  string
}

func (f *VFolderWithOwnership) IsOwner() bool {
	return f.Ownership == VFOWNERSHIP_OWNER
}

type VFolder struct {
	Id         int
	FolderNo   string
	Name       string
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
}

type UserVFolder struct {
	Id         int
	UserNo     string
	Username   string
	FolderNo   string
	Ownership  string
	GrantedBy  string // grantedBy (user_no)
	CreateTime miso.ETime
	CreateBy   string
	UpdateTime miso.ETime
	UpdateBy   string
}

func newListFilesInVFolderQuery(rail miso.Rail, tx *gorm.DB, req ListFileReq, userNo string) *gorm.DB {
	return tx.Table("file_info fi").
		Joins("LEFT JOIN file_vfolder fv ON (fi.uuid = fv.uuid AND fv.is_del = 0)").
		Joins("LEFT JOIN user_vfolder uv ON (fv.folder_no = uv.folder_no AND uv.is_del = 0)").
		Where("uv.user_no = ? AND uv.folder_no = ?", userNo, req.FolderNo)
}

func listFilesInVFolder(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (ListFilesRes, error) {
	offset := req.Page.GetOffset()
	limit := req.Page.GetLimit()

	var files []ListedFile

	t := newListFilesInVFolderQuery(rail, tx, req, user.UserNo).
		Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time`).
		Offset(offset).
		Limit(limit).
		Scan(&files)

	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files in vfolder, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == user.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesInVFolderQuery(rail, tx, req, user.UserNo).
		Select("COUNT(fi.id)").
		Scan(&total)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files in vfolder, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: miso.RespPage(req.Page, total)}, nil
}

type FileKeyName struct {
	Name string
	Uuid string
}

func queryFilenames(tx *gorm.DB, fileKeys []string) (map[string]string, error) {
	var rec []FileKeyName
	e := tx.Select("uuid, name").
		Table("file_info").
		Where("uuid IN ?", fileKeys).
		Scan(&rec).Error
	if e != nil {
		return nil, e
	}
	keyName := map[string]string{}
	for _, r := range rec {
		keyName[r.Uuid] = r.Name
	}
	return keyName, nil
}

type ListFileReq struct {
	Page       miso.Paging `json:"pagingVo"`
	Filename   *string     `json:"filename"`
	TagName    *string     `json:"tagName"`
	FolderNo   *string     `json:"folderNo"`
	FileType   *string     `json:"fileType"`
	ParentFile *string     `json:"parentFile"`
}

func ListFiles(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (ListFilesRes, error) {
	var res ListFilesRes
	var e error

	if req.FolderNo != nil && *req.FolderNo != "" {
		res, e = listFilesInVFolder(rail, tx, req, user)
	} else if req.TagName != nil && *req.TagName != "" {
		res, e = listFilesForTags(rail, tx, req, user)
	} else {
		res, e = listFilesSelective(rail, tx, req, user)
	}
	if e != nil {
		return res, e
	}

	parentFileKeys := miso.NewSet[string]()
	for _, f := range res.Payload {
		if f.ParentFile != "" {
			parentFileKeys.Add(f.ParentFile)
		}
	}

	if !parentFileKeys.IsEmpty() {
		keyName, e := queryFilenames(tx, parentFileKeys.CopyKeys())
		if e != nil {
			return res, e
		}
		for i, f := range res.Payload {
			if name, ok := keyName[f.ParentFile]; ok {
				res.Payload[i].ParentFileName = name
			}
		}
	}
	return res, e
}

func listFilesForTags(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (ListFilesRes, error) {
	var files []ListedFile
	t := newListFilesForTagsQuery(rail, tx, req, user.UserId).
		Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time`).
		Order("fi.id desc").
		Offset(req.Page.GetOffset()).
		Limit(req.Page.GetLimit()).
		Scan(&files)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == user.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesForTagsQuery(rail, tx, req, user.UserId).
		Select("count(*)").
		Scan(&total)

	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: miso.RespPage(req.Page, total)}, nil
}

func listFilesSelective(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (ListFilesRes, error) {
	/*
	   If parentFile is empty, and filename are not searched, then we only return the top level file or dir.
	   The query for tags will ignore parent_file param, so it's working fine
	*/
	if (req.ParentFile == nil || *req.ParentFile == "") && (req.Filename == nil || *req.Filename == "") {
		req.Filename = nil
		var pf = ""
		req.ParentFile = &pf // top-level file/dir
	}

	var files []ListedFile
	t := newListFilesSelectiveQuery(rail, tx, req, user.UserId).
		Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time`).
		Order("fi.file_type asc, fi.id desc").
		Offset(req.Page.GetOffset()).
		Limit(req.Page.GetLimit()).
		Scan(&files)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == user.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesSelectiveQuery(rail, tx, req, user.UserId).
		Select("count(*)").
		Scan(&total)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: miso.RespPage(req.Page, total)}, nil
}

func newListFilesSelectiveQuery(rail miso.Rail, tx *gorm.DB, req ListFileReq, userId int) *gorm.DB {

	tx = tx.Table("file_info fi").
		Where("fi.uploader_id = ?", userId).
		Where("fi.is_logic_deleted = 0 AND fi.is_del = 0")

	if req.ParentFile != nil {
		tx = tx.Where("fi.parent_file = ?", *req.ParentFile)
	}

	if req.Filename != nil && *req.Filename != "" {
		tx = tx.Where("fi.name LIKE ?", "%"+*req.Filename+"%")
	}

	if req.FileType != nil && *req.FileType != "" {
		tx = tx.Where("fi.file_type = ?", *req.FileType)
	}

	return tx
}

func newListFilesForTagsQuery(rail miso.Rail, t *gorm.DB, req ListFileReq, userId int) *gorm.DB {

	t = t.Table("file_info fi").
		Joins("LEFT JOIN file_tag ft ON (ft.user_id = ? AND fi.id = ft.file_id)", userId).
		Joins("LEFT JOIN tag t ON (ft.tag_id = t.id)")

	if req.Filename != nil && *req.Filename != "" {
		t = t.Where("fi.name LIKE ?", "%"+*req.Filename+"%")
	}

	t = t.Where("fi.uploader_id = ?", userId).
		Where("fi.file_type = 'FILE'").
		Where("fi.is_del = 0").
		Where("fi.is_logic_deleted = 0").
		Where("ft.is_del = 0").
		Where("t.is_del = 0").
		Where("t.name = ?", *req.TagName)

	return t
}

type PreflightCheckReq struct {
	Filename      string `form:"fileName"`
	ParentFileKey string `form:"parentFileKey"`
}

func FileExists(c miso.Rail, tx *gorm.DB, req PreflightCheckReq, user common.User) (any, error) {
	var id int
	t := tx.Table("file_info").
		Select("id").
		Where("parent_file = ?", req.ParentFileKey).
		Where("name = ?", req.Filename).
		Where("uploader_id = ?", user.UserId).
		Where("file_type = ?", FILE_TYPE_FILE).
		Where("is_logic_deleted = ?", FILE_LDEL_N).
		Where("is_del = ?", common.IS_DEL_N).
		Limit(1).
		Scan(&id)

	if t.Error != nil {
		return false, fmt.Errorf("failed to match file, %v", t.Error)
	}

	return id > 0, nil
}

type ListFileTagReq struct {
	Page   miso.Paging `json:"pagingVo"`
	FileId int         `json:"fileId" validation:"positive"`
}

type ListFileTagRes struct {
	Page    miso.Paging     `json:"pagingVo"`
	Payload []ListedFileTag `json:"payload"`
}

func ListFileTags(rail miso.Rail, tx *gorm.DB, req ListFileTagReq, user common.User) (ListFileTagRes, error) {
	var ftags []ListedFileTag

	t := newListFileTagsQuery(rail, tx, req, user.UserId).
		Select("*").
		Scan(&ftags)
	if t.Error != nil {
		return ListFileTagRes{}, fmt.Errorf("failed to list file tags for req: %v, %v", req, t.Error)
	}

	var total int
	t = newListFileTagsQuery(rail, tx, req, user.UserId).
		Select("count(*)").
		Scan(&total)
	if t.Error != nil {
		return ListFileTagRes{}, fmt.Errorf("failed to count file tags for req: %v, %v", req, t.Error)
	}

	return ListFileTagRes{Payload: ftags}, nil
}

func newListFileTagsQuery(c miso.Rail, tx *gorm.DB, r ListFileTagReq, userId int) *gorm.DB {
	return tx.Table("file_tag ft").
		Joins("LEFT JOIN tag t ON ft.tag_id = t.id").
		Where("t.user_id = ? AND ft.file_id = ? AND ft.is_del = 0 AND t.is_del = 0", userId, r.FileId)
}

func findFile(c miso.Rail, tx *gorm.DB, fileKey string) (FileInfo, error) {
	var f FileInfo
	t := tx.Raw("SELECT * FROM file_info WHERE uuid = ? AND is_del = 0", fileKey).
		Scan(&f)
	return f, t.Error
}

func findFileById(rail miso.Rail, tx *gorm.DB, id int) (FileInfo, error) {
	var f FileInfo

	t := tx.Raw("SELECT * FROM file_info WHERE id = ? AND is_del = 0", id).
		Scan(&f)
	if t.Error != nil {
		return f, t.Error
	}
	return f, nil
}

type FetchParentFileReq struct {
	FileKey string `form:"fileKey"`
}

func FindParentFile(c miso.Rail, tx *gorm.DB, req FetchParentFileReq, user common.User) (ParentFileInfo, error) {
	userId := user.UserId

	var f FileInfo
	f, e := findFile(c, tx, req.FileKey)
	if e != nil {
		return ParentFileInfo{}, e
	}
	if f.IsZero() {
		return ParentFileInfo{}, miso.NewErr("File not found")
	}

	// dir is only visible to the uploader for now
	if f.UploaderId != userId {
		return ParentFileInfo{}, miso.NewErr("Not permitted")
	}

	if f.ParentFile == "" {
		return ParentFileInfo{Zero: true}, nil
	}

	pf, e := findFile(c, tx, f.ParentFile)
	if e != nil {
		return ParentFileInfo{}, e
	}
	if pf.IsZero() {
		return ParentFileInfo{}, miso.NewErr("File not found", fmt.Sprintf("ParentFile %v not found", f.ParentFile))
	}

	return ParentFileInfo{FileKey: pf.Uuid, Filename: pf.Name, Zero: false}, nil
}

type MakeDirReq struct {
	ParentFile string `json:"parentFile"`                 // Key of parent file
	Name       string `json:"name" validation:"notEmpty"` // name of the directory
}

func MakeDir(rail miso.Rail, tx *gorm.DB, req MakeDirReq, user common.User) (string, error) {

	var dir FileInfo
	dir.Name = req.Name
	dir.Uuid = miso.GenIdP("ZZZ")
	dir.SizeInBytes = 0
	dir.FileType = FILE_TYPE_DIR

	if e := _saveFile(rail, tx, dir, user); e != nil {
		return "", e
	}

	if req.ParentFile != "" {
		if e := MoveFileToDir(rail, tx, MoveIntoDirReq{Uuid: dir.Uuid, ParentFileUuid: req.ParentFile}, user); e != nil {
			return dir.Uuid, e
		}
	}

	return dir.Uuid, nil
}

type MoveIntoDirReq struct {
	Uuid           string `json:"uuid" validation:"notEmpty"`
	ParentFileUuid string `json:"parentFileUuid"`
}

func MoveFileToDir(rail miso.Rail, tx *gorm.DB, req MoveIntoDirReq, user common.User) error {
	if req.Uuid == "" || req.ParentFileUuid == "" || req.Uuid == req.ParentFileUuid {
		return nil
	}

	// lock the file
	return _lockFileExec(rail, req.Uuid, func() error {

		// lock directory
		return _lockFileExec(rail, req.ParentFileUuid, func() error {

			pr, e := findFile(rail, tx, req.ParentFileUuid)
			if e != nil {
				return fmt.Errorf("failed to find parentFile, %v", e)
			}
			rail.Debugf("parentFile: %+v", pr)

			if pr.IsZero() {
				return fmt.Errorf("perentFile not found, parentFileKey: %v", req.ParentFileUuid)
			}

			if pr.UploaderId != user.UserId {
				return miso.NewErr("You are not the owner of this directory")
			}

			if pr.FileType != FILE_TYPE_DIR {
				return miso.NewErr("Target file is not a directory")
			}

			if pr.IsLogicDeleted != FILE_LDEL_N {
				return miso.NewErr("Target file deleted")
			}

			return tx.Exec("UPDATE file_info SET parent_file = ?, update_by = ?, update_time = ? WHERE uuid = ?",
				req.ParentFileUuid, user.Username, time.Now(), req.Uuid).
				Error
		})
	})
}

func _saveFile(rail miso.Rail, tx *gorm.DB, f FileInfo, user common.User) error {
	userId := user.UserId
	uname := user.Username
	now := miso.Now()

	f.IsLogicDeleted = FILE_LDEL_N
	f.IsPhysicDeleted = FILE_PDEL_N
	f.UploaderId = userId
	f.UploaderName = uname
	f.CreateBy = uname
	f.UploadTime = now
	f.CreateTime = now

	return tx.Table("file_info").
		Omit("id", "update_time", "update_by").
		Create(&f).Error
}

func _lockFileExec(rail miso.Rail, fileKey string, r miso.Runnable) error {
	return miso.RLockExec(rail, "file:uuid:"+fileKey, r)
}

func _lockFileGet[T any](rail miso.Rail, fileKey string, r miso.LRunnable[T]) (any, error) {
	return miso.RLockRun(rail, "file:uuid:"+fileKey, r)
}

type CreateVFolderReq struct {
	Name string `json:"name"`
}

func CreateVFolder(rail miso.Rail, tx *gorm.DB, r CreateVFolderReq, user common.User) (string, error) {
	userNo := user.UserNo

	v, e := miso.RLockRun(rail, "vfolder:user:"+userNo, func() (any, error) {

		var id int
		t := tx.Table("vfolder vf").
			Select("vf.id").
			Joins("LEFT JOIN user_vfolder uv ON (vf.folder_no = uv.folder_no)").
			Where("uv.user_no = ? AND uv.ownership = 'OWNER'", userNo).
			Where("vf.name = ?", r.Name).
			Where("vf.is_del = 0 AND uv.is_del = 0").
			Limit(1).
			Scan(&id)
		if t.Error != nil {
			return "", t.Error
		}
		if id > 0 {
			return "", miso.NewErr(fmt.Sprintf("Found folder with same name ('%s')", r.Name))
		}

		folderNo := miso.GenIdP("VFLD")

		e := tx.Transaction(func(tx *gorm.DB) error {

			ctime := miso.Now()

			// for the vfolder
			vf := VFolder{Name: r.Name, FolderNo: folderNo, CreateTime: ctime, CreateBy: user.Username}
			if e := tx.Omit("id", "update_by", "update_time").Table("vfolder").Create(&vf).Error; e != nil {
				return fmt.Errorf("failed to save VFolder, %v", e)
			}

			// for the user - vfolder relation
			uv := UserVFolder{
				FolderNo:   folderNo,
				UserNo:     userNo,
				Username:   user.Username,
				Ownership:  VFOWNERSHIP_OWNER,
				GrantedBy:  userNo,
				CreateTime: ctime,
				CreateBy:   user.Username}
			if e := tx.Omit("id", "update_by", "update_time").Table("user_vfolder").Create(&uv).Error; e != nil {
				return fmt.Errorf("failed to save UserVFolder, %v", e)
			}
			return nil
		})
		if e != nil {
			return "", e
		}

		return folderNo, nil
	})
	if e != nil {
		return "", e
	}
	return v.(string), e
}

func ListDirs(c miso.Rail, tx *gorm.DB, user common.User) ([]ListedDir, error) {
	var dirs []ListedDir
	e := tx.Table("file_info").
		Select("id, uuid, name").
		Where("uploader_id = ?", user.UserId).
		Where("file_type = 'DIR'").
		Where("is_logic_deleted = 0").
		Where("is_del = 0").
		Scan(&dirs).Error
	return dirs, e
}

func findVFolder(rail miso.Rail, tx *gorm.DB, folderNo string, userNo string) (VFolderWithOwnership, error) {
	var vfo VFolderWithOwnership
	t := tx.Table("vfolder vf").
		Select("vf.*, uv.ownership").
		Joins("LEFT JOIN user_vfolder uv ON (vf.folder_no = uv.folder_no AND uv.is_del = 0)").
		Where("vf.is_del = 0").
		Where("uv.user_no = ?", userNo).
		Where("uv.folder_no = ?", folderNo).
		Limit(1).
		Scan(&vfo)
	if t.Error != nil {
		return vfo, fmt.Errorf("failed to fetch vfolder info for current user, userNo: %v, folderNo: %v, %v", userNo, folderNo, t.Error)
	}
	if t.RowsAffected < 1 {
		return vfo, fmt.Errorf("vfolder not found, userNo: %v, folderNo: %v", userNo, folderNo)
	}
	return vfo, nil
}

func _lockFolderExec(c miso.Rail, folderNo string, r miso.Runnable) error {
	return miso.RLockExec(c, "vfolder:"+folderNo, r)
}

func _lockFolderGet[T any](c miso.Rail, folderNo string, r miso.LRunnable[T]) (any, error) {
	return miso.RLockRun(c, "vfolder:"+folderNo, r)
}

func ShareVFolder(rail miso.Rail, tx *gorm.DB, sharedTo UserInfo, folderNo string, user common.User) error {
	if user.UserNo == sharedTo.UserNo {
		return nil
	}
	return _lockFolderExec(rail, folderNo, func() error {
		vfo, e := findVFolder(rail, tx, folderNo, user.UserNo)
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return miso.NewErr("Operation not permitted")
		}

		var id int
		e = tx.Table("user_vfolder").
			Select("id").
			Where("folder_no = ?", folderNo).
			Where("user_no = ?", sharedTo.UserNo).
			Where("is_del = 0").
			Limit(1).
			Scan(&id).Error
		if e != nil {
			return fmt.Errorf("error occurred while querying user_vfolder, %v", e)
		}
		if id > 0 {
			rail.Infof("VFolder is shared already, folderNo: %s, sharedTo: %s", folderNo, sharedTo.Username)
			return nil
		}

		uv := UserVFolder{
			FolderNo:   folderNo,
			UserNo:     sharedTo.UserNo,
			Username:   sharedTo.Username,
			Ownership:  VFOWNERSHIP_GRANTED,
			GrantedBy:  user.Username,
			CreateTime: miso.Now(),
			CreateBy:   user.Username,
		}
		if e := tx.Omit("id", "update_by", "update_time").Table("user_vfolder").Create(&uv).Error; e != nil {
			return fmt.Errorf("failed to save UserVFolder, %v", e)
		}
		rail.Infof("VFolder %s shared to %s by %s", folderNo, sharedTo.Username, user.Username)
		return nil
	})
}

type RemoveGrantedFolderAccessReq struct {
	FolderNo string `json:"folderNo"`
	UserNo   string `json:"userNo"`
}

func RemoveVFolderAccess(rail miso.Rail, tx *gorm.DB, req RemoveGrantedFolderAccessReq, user common.User) error {
	if user.UserNo == req.UserNo {
		return nil
	}
	return _lockFolderExec(rail, req.FolderNo, func() error {
		vfo, e := findVFolder(rail, tx, req.FolderNo, user.UserNo)
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return miso.NewErr("Operation not permitted")
		}
		return tx.
			Exec("DELETE FROM user_vfolder WHERE folder_no = ? AND user_no = ? AND ownership = 'GRANTED'", req.FolderNo, req.UserNo).
			Error
	})
}

func ListVFolderBrief(rail miso.Rail, tx *gorm.DB, user common.User) ([]VFolderBrief, error) {
	var vfb []VFolderBrief
	e := tx.Select("f.folder_no, f.name").
		Table("vfolder f").
		Joins("LEFT JOIN user_vfolder uv ON (f.folder_no = uv.folder_no AND uv.is_del = 0)").
		Where("f.is_del = 0 AND uv.user_no = ? AND uv.ownership = 'OWNER'", user.UserNo).
		Scan(&vfb).Error
	return vfb, e
}

type AddFileToVfolderReq struct {
	FolderNo string   `json:"folderNo"`
	FileKeys []string `json:"fileKeys"`
}

func AddFileToVFolder(rail miso.Rail, tx *gorm.DB, req AddFileToVfolderReq, user common.User) error {

	if len(req.FileKeys) < 1 {
		return nil
	}

	return _lockFolderExec(rail, req.FolderNo, func() error {

		vfo, e := findVFolder(rail, tx, req.FolderNo, user.UserNo)
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return miso.NewErr("Operation not permitted")
		}

		filtered := miso.Distinct(req.FileKeys)
		if len(filtered) < 1 {
			return nil
		}
		userId := user.UserId
		now := miso.Now()
		username := user.Username
		for _, fk := range filtered {
			f, e := findFile(rail, tx, fk)
			if e != nil {
				return e
			}
			if f.IsZero() {
				continue // file not found
			}
			if f.UploaderId != userId {
				continue // not the uploader of the file
			}

			if f.FileType != FILE_TYPE_FILE {
				continue // not a file type, may be a dir
			}

			var id int
			e = tx.Select("id").
				Table("file_vfolder").
				Where("folder_no = ? AND uuid = ?", req.FolderNo, fk).
				Scan(&id).
				Error
			if e != nil {
				return fmt.Errorf("failed to query file_vfolder record, %v", e)
			}
			if id > 0 {
				continue // file already in vfolder
			}

			fvf := FileVFolder{FolderNo: req.FolderNo, Uuid: fk, CreateTime: now, CreateBy: username}
			e = tx.Table("file_vfolder").Omit("id", "update_by", "update_time").Create(&fvf).Error
			if e != nil {
				return fmt.Errorf("failed to save file_vfolder record, %v", e)
			}
		}
		return nil
	})
}

type RemoveFileFromVfolderReq struct {
	FolderNo string   `json:"folderNo"`
	FileKeys []string `json:"fileKeys"`
}

func RemoveFileFromVFolder(rail miso.Rail, tx *gorm.DB, req RemoveFileFromVfolderReq, user common.User) error {
	if len(req.FileKeys) < 1 {
		return nil
	}

	return _lockFolderExec(rail, req.FolderNo, func() error {

		vfo, e := findVFolder(rail, tx, req.FolderNo, user.UserNo)
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return miso.NewErr("Operation not permitted")
		}

		filtered := miso.Distinct(req.FileKeys)
		if len(filtered) < 1 {
			return nil
		}

		for _, fk := range filtered {
			f, e := findFile(rail, tx, fk)
			if e != nil {
				return e
			}
			if f.IsZero() {
				continue // file not found
			}
			// TODO: get rid of the userId
			if f.UploaderId != user.UserId {
				continue // not the uploader of the file
			}
			if f.FileType != FILE_TYPE_FILE {
				continue // not a file type, may be a dir
			}

			e = tx.Exec("DELETE FROM file_vfolder WHERE folder_no = ? AND uuid = ?", req.FolderNo, fk).Error
			if e != nil {
				return fmt.Errorf("failed to delete file_vfolder record, %v", e)
			}
		}

		return nil
	})
}

type ListVFolderReq struct {
	Page miso.Paging `json:"pagingVo"`
	Name string      `json:"name"`
}

func ListVFolders(rail miso.Rail, tx *gorm.DB, req ListVFolderReq, user common.User) (ListVFolderRes, error) {
	t := newListVFoldersQuery(rail, tx, req, user.UserNo).
		Select("f.id, f.create_time, f.create_by, f.update_time, f.update_by, f.folder_no, f.name, uv.ownership").
		Order("f.id DESC").
		Offset(req.Page.GetOffset()).
		Limit(req.Page.GetLimit())

	var lvf []ListedVFolder
	if e := t.Scan(&lvf).Error; e != nil {
		return ListVFolderRes{}, fmt.Errorf("failed to query vfolder, req: %+v, %v", req, e)
	}

	var total int
	e := newListVFoldersQuery(rail, tx, req, user.UserNo).
		Select("COUNT(*)").
		Scan(&total).Error
	if e != nil {
		return ListVFolderRes{}, fmt.Errorf("failed to count vfolder, req: %+v, %v", req, e)
	}

	return ListVFolderRes{Page: miso.RespPage(req.Page, total), Payload: lvf}, nil
}

func newListVFoldersQuery(rail miso.Rail, tx *gorm.DB, req ListVFolderReq, userNo string) *gorm.DB {
	t := tx.Table("vfolder f").
		Joins("LEFT JOIN user_vfolder uv ON (f.folder_no = uv.folder_no AND uv.is_del = 0)").
		Where("f.is_del = 0 AND uv.user_no = ?", userNo)

	if req.Name != "" {
		t = t.Where("f.name like ?", "%"+req.Name+"%")
	}
	return t
}

type RemoveGrantedAccessReq struct {
	FileId int `json:"fileId" validation:"positive"`
	UserId int `json:"userId" validation:"positive"`
}

type ListGrantedFolderAccessReq struct {
	Page     miso.Paging `json:"pagingVo"`
	FolderNo string      `json:"folderNo"`
}

type ListGrantedFolderAccessRes struct {
	Page    miso.Paging          `json:"pagingVo"`
	Payload []ListedFolderAccess `json:"payload"`
}

type ListedFolderAccess struct {
	UserNo     string     `json:"userNo"`
	Username   string     `json:"username"`
	CreateTime miso.ETime `json:"createTime"`
}

func ListGrantedFolderAccess(rail miso.Rail, tx *gorm.DB, req ListGrantedFolderAccessReq, user common.User) (ListGrantedFolderAccessRes, error) {
	folderNo := req.FolderNo
	vfo, e := findVFolder(rail, tx, folderNo, user.UserNo)
	if e != nil {
		return ListGrantedFolderAccessRes{}, e
	}
	if !vfo.IsOwner() {
		return ListGrantedFolderAccessRes{}, miso.NewErr("Operation not permitted")
	}

	var l []ListedFolderAccess
	e = newListGrantedFolderAccessQuery(rail, tx, req).
		Select("user_no", "create_time", "username").
		Offset(req.Page.GetOffset()).
		Limit(req.Page.GetLimit()).
		Scan(&l).Error
	if e != nil {
		return ListGrantedFolderAccessRes{}, fmt.Errorf("failed to list granted folder access, req: %+v, %v", req, e)
	}

	var total int
	e = newListGrantedFolderAccessQuery(rail, tx, req).
		Select("count(*)").
		Scan(&total).Error
	if e != nil {
		return ListGrantedFolderAccessRes{}, fmt.Errorf("failed to count granted folder access, req: %+v, %v", req, e)
	}

	userNos := []string{}
	for _, p := range l {
		if p.Username == "" {
			userNos = append(userNos, p.UserNo)
		}
	}

	if len(userNos) > 0 { // since v0.0.4 this is not needed anymore, but we keep it here for backward compatibility
		unr, e := FetchUsernames(rail, FetchUsernamesReq{UserNos: userNos})
		if e != nil {
			rail.Errorf("Failed to fetch usernames, %v", e)
		} else {
			for i, p := range l {
				if name, ok := unr.UserNoToUsername[p.UserNo]; ok {
					p.Username = name
					l[i] = p
				}
			}
		}
	}

	return ListGrantedFolderAccessRes{Payload: l, Page: miso.RespPage(req.Page, total)}, nil
}

func newListGrantedFolderAccessQuery(rail miso.Rail, tx *gorm.DB, r ListGrantedFolderAccessReq) *gorm.DB {
	return tx.Table("user_vfolder").
		Where("folder_no = ? AND ownership = 'GRANTED' AND is_del = 0", r.FolderNo)
}

type UpdateFileReq struct {
	Id   int    `json:"id" validation:"positive"`
	Name string `json:"name"`
}

func UpdateFile(rail miso.Rail, tx *gorm.DB, r UpdateFileReq, user common.User) error {
	f, e := findFileById(rail, tx, r.Id)
	if e != nil {
		return e
	}
	if f.IsZero() {
		return miso.NewErr("File not found")
	}

	// dir is only visible to the uploader for now
	if f.UploaderId != user.UserId {
		return miso.NewErr("Not permitted")
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return miso.NewErr("Name can't be empty")
	}

	return tx.
		Exec("UPDATE file_info SET name = ? WHERE id = ? AND is_logic_deleted = 0 AND is_del = 0", r.Name, r.Id).
		Error
}

func ListAllTags(rail miso.Rail, tx *gorm.DB, user common.User) ([]string, error) {
	var l []string
	e := tx.Raw("SELECT t.name FROM tag t WHERE t.user_id = ? AND t.is_del = 0", user.UserId).
		Scan(&l).
		Error

	return l, e
}

type TagFileReq struct {
	FileId  int    `json:"fileId" validation:"positive"`
	TagName string `json:"tagName" validation:"notEmpty"`
}

func TagFile(rail miso.Rail, tx *gorm.DB, req TagFileReq, user common.User) error {
	req.TagName = strings.TrimSpace(req.TagName)
	return _lockFileTagExec(rail, user.UserId, req.TagName, func() error {

		// find the tag first, and create one for current user if necessary
		tagId, e := tryCreateTag(rail, tx, user.UserId, req.TagName, user.Username)
		if e != nil {
			return e
		}
		if tagId < 1 {
			return fmt.Errorf("tagId illegal, shouldn't be less than 1")
		}

		// check if it's already tagged
		ft, e := findFileTag(rail, tx, req.FileId, tagId)
		if e != nil {
			return e
		}

		if ft.IsZero() {
			ft = FileTag{UserId: user.UserId,
				FileId:     req.FileId,
				TagId:      tagId,
				CreateTime: miso.Now(),
				CreateBy:   user.Username,
			}
			return tx.Table("file_tag").
				Omit("id", "update_time", "update_by").
				Create(&ft).Error
		}

		if ft.IsDel == common.IS_DEL_Y {
			return tx.Exec("UPDATE file_tag SET is_del = 0, update_time = ?, update_by = ? WHERE id = ?", time.Now(), user.Username, ft.Id).
				Error
		}

		return nil
	})
}

func tryCreateTag(rail miso.Rail, tx *gorm.DB, userId int, tagName string, username string) (int, error) {
	t, e := findTag(rail, tx, userId, tagName)
	if e != nil {
		return 0, fmt.Errorf("failed to find tag, userId: %v, tagName: %v, %e", userId, tagName, e)
	}

	if t.IsZero() {
		t = Tag{Name: tagName, UserId: userId, CreateBy: username, CreateTime: miso.Now()}
		e := tx.Table("tag").Omit("id", "update_time", "update_by").Create(&t).Error
		if e != nil {
			return 0, fmt.Errorf("failed to create tag, userId: %v, tagName: %v, %e", userId, tagName, e)
		}
		return t.Id, nil
	}

	if t.IsDel == common.IS_DEL_Y {
		e := tx.Exec("UPDATE tag SET is_del = 0, update_time = ?, update_by = ? WHERE id = ?", time.Now(), username, t.Id).Error
		if e != nil {
			return 0, fmt.Errorf("failed to update tag, id: %v, %e", t.Id, e)
		}
	}

	return t.Id, nil
}

func findTag(rail miso.Rail, tx *gorm.DB, userId int, tagName string) (Tag, error) {
	var t Tag
	e := tx.Raw("SELECT * FROM tag WHERE user_id = ? AND name = ?", userId, tagName).
		Scan(&t).Error
	return t, e
}

func findFileTag(rail miso.Rail, tx *gorm.DB, fileId int, tagId int) (FileTag, error) {
	var ft FileTag
	e := tx.Raw("SELECT * FROM file_tag WHERE file_id = ? AND tag_id = ?", fileId, tagId).
		Scan(&ft).Error
	return ft, e
}

type UntagFileReq struct {
	FileId  int    `json:"fileId" validation:"positive"`
	TagName string `json:"tagName" validation:"notEmpty"`
}

func UntagFile(rail miso.Rail, tx *gorm.DB, req UntagFileReq, user common.User) error {
	req.TagName = strings.TrimSpace(req.TagName)
	return _lockFileTagExec(rail, user.UserId, req.TagName, func() error {
		// each tag is bound to a specific user
		tag, e := findTag(rail, tx, user.UserId, req.TagName)
		if e != nil {
			return e
		}
		if tag.IsZero() {
			rail.Infof("Tag for '%v' doesn't exist, unable to untag file", req.TagName)
			return nil // tag doesn't exist
		}

		fileTag, e := findFileTag(rail, tx, req.FileId, tag.Id)
		if e != nil {
			return e
		}

		if fileTag.IsZero() || fileTag.IsDel == common.IS_DEL_Y {
			rail.Infof("FileTag for file_id: %d, tag_id: %d, doesn't exist", req.FileId, tag.Id)
			return nil
		}

		return tx.Transaction(func(txx *gorm.DB) error {
			// it was a logic delete in file-server, it now becomes a physical delete
			e = txx.Exec("DELETE FROM file_tag WHERE id = ?", fileTag.Id).Error
			if e != nil {
				return fmt.Errorf("failed to update file_tag, %v", e)
			}

			rail.Infof("Untagged file, file_id: %d, tag_name: %s", req.FileId, req.TagName)

			/*
			   check if the tag is still associated with other files, if not, we remove it
			   remember, the tag is bound for a specific user only, so this doesn't affect
			   other users
			*/
			var anyFileTagId int
			e = txx.Table("file_tag").
				Where("tag_id = ? AND is_del = 0", tag.Id).
				Limit(1).
				Scan(&anyFileTagId).
				Error
			if e != nil {
				return e
			}

			if anyFileTagId < 1 {
				// it was a logic delete in file-server, it now becomes a physical delete
				return txx.Exec("delete from tag where id = ?", tag.Id).Error
			}
			return nil
		})
	})
}

func _lockFileTagExec(rail miso.Rail, userId int, tagName string, run miso.Runnable) error {
	return miso.RLockExec(rail, fmt.Sprintf("file:tag:uid:%d:name:%s", userId, tagName), run)
}

type CreateFileReq struct {
	Filename         string   `json:"filename"`
	FakeFstoreFileId string   `json:"fstoreFileId"`
	Tags             []string `json:"tags"`
	ParentFile       string   `json:"parentFile"`
}

func CreateFile(rail miso.Rail, tx *gorm.DB, r CreateFileReq, user common.User) error {
	fsf, e := FetchFstoreFileInfo(rail, "", r.FakeFstoreFileId)
	if e != nil {
		return fmt.Errorf("failed to fetch file info from fstore, %v", e)
	}
	if fsf.IsZero() || fsf.Status != FS_STATUS_NORMAL {
		return miso.NewErr("File not found or deleted")
	}

	var f FileInfo
	f.Name = r.Filename
	f.Uuid = miso.GenIdP("ZZZ")
	f.FstoreFileId = fsf.FileId
	f.SizeInBytes = fsf.Size
	f.FileType = FILE_TYPE_FILE

	if e := _saveFile(rail, tx, f, user); e != nil {
		return e
	}

	if r.ParentFile != "" {
		if e := MoveFileToDir(rail, tx, MoveIntoDirReq{Uuid: f.Uuid, ParentFileUuid: r.ParentFile}, user); e != nil {
			return e
		}
	}

	// TODO: Since v0.0.4, this is based on event-pump binlog event
	// if isImage(f.Name) {
	// 	if e := bus.SendToEventBus(CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcBus); e != nil {
	// 		c.Errorf("Failed to send CompressImageEvent, uuid: %v, %v", f.Uuid, e)
	// 	}
	// }

	return nil
}

func isImage(name string) bool {
	i := strings.LastIndex(name, ".")
	if i < 0 || i == len([]rune(name))-1 {
		return false
	}

	suf := string(name[i+1:])
	return _imageSuffix.Has(strings.ToLower(suf))
}

type DeleteFileReq struct {
	Uuid string `json:"uuid"`
}

func DeleteFile(rail miso.Rail, tx *gorm.DB, req DeleteFileReq, user common.User) error {
	return _lockFileExec(rail, req.Uuid, func() error {
		f, e := findFile(rail, tx, req.Uuid)
		if e != nil {
			return fmt.Errorf("unable to find file, uuid: %v, %v", req.Uuid, e)
		}
		if f.IsZero() {
			return miso.NewErr("File not found")
		}

		if f.UploaderId != user.UserId {
			return miso.NewErr("Not permitted")
		}

		if f.IsLogicDeleted == FILE_LDEL_Y {
			return nil // deleted already
		}

		if f.FileType == FILE_TYPE_DIR { // if it's dir make sure it's empty
			var anyId int
			e := tx.Select("id").
				Table("file_info").
				Where("parent_file = ? AND is_logic_deleted = 0 AND is_del = 0", req.Uuid).
				Limit(1).
				Scan(&anyId).Error
			if e != nil {
				return fmt.Errorf("failed to count files in dir, uuid: %v, %v", req.Uuid, e)
			}
			if anyId > 0 {
				return miso.NewErr("Directory is not empty, unable to delete it")
			}
		}

		if f.FstoreFileId != "" {
			if e := DeleteFstoreFile(rail, f.FstoreFileId); e != nil {
				return fmt.Errorf("failed to delete fstore file, fileId: %v, %v", f.FstoreFileId, e)
			}
		}

		if f.Thumbnail != "" {
			if e := DeleteFstoreFile(rail, f.Thumbnail); e != nil {
				return fmt.Errorf("failed to delete fstore file (thumbnail), fileId: %v, %v", f.Thumbnail, e)
			}
		}

		return tx.
			Exec("UPDATE file_info SET is_logic_deleted = 1, logic_delete_time = NOW() WHERE id = ? AND is_logic_deleted = 0", f.Id).
			Error
	})
}

func validateFileAccess(rail miso.Rail, tx *gorm.DB, fileKey string, userId int, userNo string) (bool, FileDownloadInfo, error) {
	var f FileDownloadInfo
	e := tx.
		Select("fi.id 'file_id', fi.fstore_file_id, fi.name, fi.uploader_id, fi.is_logic_deleted, fi.file_type").
		Table("file_info fi").
		Where("fi.uuid = ? AND fi.is_del = 0", fileKey).
		Limit(1).
		Scan(&f).Error
	if e != nil {
		return false, f, e
	}

	if f.FileId < 1 {
		return false, f, miso.NewErr("File not found")
	}
	if f.IsLogicDeleted == FILE_LDEL_Y {
		return false, f, miso.NewErr("File deleted")
	}
	if f.FileType != FILE_TYPE_FILE {
		return false, f, miso.NewErr("Downloading a directory is not supported")
	}

	permitted := f.UploaderId == userId // owner of the file

	// user may have access to the vfolder, which contains thefile
	if !permitted {
		var uvid int
		e := tx.
			Select("uv.id").
			Table("file_info fi").
			Joins("LEFT JOIN file_vfolder fv ON (fi.uuid = fv.uuid AND fv.is_del = 0)").
			Joins("LEFT JOIN user_vfolder uv ON (uv.user_no = ? AND uv.folder_no = fv.folder_no AND uv.is_del = 0)", userNo).
			Where("fi.id = ?", f.FileId).
			Limit(1).
			Scan(&uvid).Error
		if e != nil {
			return false, f, fmt.Errorf("failed to query user folder relation for file, id: %v, %v", f.FileId, e)
		}
		permitted = uvid > 0 // granted access to a folder that contains this file
	}
	return permitted, f, nil
}

type GenerateTempTokenReq struct {
	FileKey string `json:"fileKey"`
}

func GenTempToken(rail miso.Rail, tx *gorm.DB, r GenerateTempTokenReq, user common.User) (string, error) {
	// TODO: get rid of userId
	ok, f, e := validateFileAccess(rail, tx, r.FileKey, user.UserId, user.UserNo)
	if e != nil {
		return "", e
	}
	if !ok {
		return "", miso.NewErr("Not permitted")
	}
	if f.FstoreFileId == "" {
		return "", miso.NewErr("File cannot be downloaded, please contact system administrator")
	}

	t, e := GetFstoreTmpToken(rail, f.FstoreFileId, f.Name)
	if e != nil {
		return "", e
	}
	return t, nil
}

type ListFilesInDirReq struct {
	FileKey string `form:"fileKey"`
	Limit   int    `form:"limit"`
	Page    int    `form:"page"`
}

func ListFilesInDir(rail miso.Rail, tx *gorm.DB, req ListFilesInDirReq) ([]string, error) {
	if req.Limit < 0 || req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page < 1 {
		req.Page = 1
	}

	var fileKeys []string
	e := tx.Table("file_info").
		Select("uuid").
		Where("parent_file = ?", req.FileKey).
		Where("file_type = 'FILE'").
		Where("is_del = 0").
		Offset((req.Page - 1) * req.Limit).
		Limit(req.Limit).
		Scan(&fileKeys).Error
	return fileKeys, e
}

type FetchFileInfoReq struct {
	FileKey string `form:"fileKey"`
}

type FileInfoResp struct {
	Name         string `json:"name"`
	Uuid         string `json:"uuid"`
	SizeInBytes  int64  `json:"sizeInBytes"`
	UploaderId   int    `json:"uploaderId"`
	UploaderName string `json:"uploaderName"`
	IsDeleted    bool   `json:"isDeleted"`
	FileType     string `json:"fileType"`
	ParentFile   string `json:"parentFile"`
	LocalPath    string `json:"localPath"`
	FstoreFileId string `json:"fstoreFileId"`
	Thumbnail    string `json:"thumbnail"`
}

func FetchFileInfoInternal(rail miso.Rail, tx *gorm.DB, req FetchFileInfoReq) (FileInfoResp, error) {
	var fir FileInfoResp
	f, e := findFile(rail, tx, req.FileKey)
	if e != nil {
		return fir, e
	}
	if f.IsZero() {
		return fir, miso.NewErr("File not found")
	}

	fir.Name = f.Name
	fir.Uuid = f.Uuid
	fir.SizeInBytes = f.SizeInBytes
	fir.UploaderId = f.UploaderId
	fir.UploaderName = f.UploaderName
	fir.IsDeleted = f.IsLogicDeleted == FILE_LDEL_Y
	fir.FileType = f.FileType
	fir.ParentFile = f.ParentFile
	fir.LocalPath = "" // files are managed by the mini-fstore, this field will no longer contain any value in it
	fir.FstoreFileId = f.FstoreFileId
	fir.Thumbnail = f.Thumbnail
	return fir, nil
}

type ValidateFileOwnerReq struct {
	FileKey string `form:"fileKey"`
	UserId  int    `form:"userId"`
}

func ValidateFileOwner(rail miso.Rail, tx *gorm.DB, q ValidateFileOwnerReq) (bool, error) {
	var id int
	e := tx.Select("id").
		Table("file_info").
		Where("uuid = ?", q.FileKey).
		Where("uploader_id = ?", q.UserId).
		Where("is_logic_deleted = 0").
		Limit(1).
		Scan(&id).Error
	return id > 0, e
}

func ReactOnImageCompressed(rail miso.Rail, tx *gorm.DB, evt CompressImageEvent) error {
	return _lockFileExec(rail, evt.FileKey, func() error {
		f, e := findFile(rail, tx, evt.FileKey)
		if e != nil {
			rail.Errorf("Unable to find file, uuid: %v, %v", evt.FileKey, e)
			return nil
		}
		if f.IsZero() {
			rail.Errorf("File not found, uuid: %v", evt.FileKey)
			return nil
		}

		return tx.Exec("UPDATE file_info SET thumbnail = ? WHERE uuid = ?", evt.FileId, evt.FileKey).
			Error
	})
}

type FileCompressInfo struct {
	Id           int
	Name         string
	Uuid         string
	FstoreFileId string
}

func CompensateImageCompression(rail miso.Rail, tx *gorm.DB) error {

	limit := 500
	minId := 0

	for {
		var files []FileCompressInfo
		t := tx.
			Raw(`SELECT id, name, uuid, fstore_file_id
			FROM file_info
			WHERE id > ?
			AND file_type = 'file'
			AND is_logic_deleted = 0
			AND thumbnail = ''
			ORDER BY id ASC
			LIMIT ?`, minId, limit).
			Scan(&files)
		if t.Error != nil {
			return t.Error
		}
		if t.RowsAffected < 1 || len(files) < 1 {
			return nil // the end
		}

		for _, f := range files {
			if isImage(f.Name) {
				if e := miso.PubEventBus(rail, CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcEventBus); e != nil {
					rail.Errorf("Failed to send CompressImageEvent, uuid: %v, %v", f.Uuid, e)
				}
			}
		}

		minId = files[len(files)-1].Id
		rail.Infof("CompensateImageCompression, minId: %v", minId)
	}
}
