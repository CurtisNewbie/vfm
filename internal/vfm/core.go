package vfm

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/curtisnewbie/gocommon/common"
	fstore "github.com/curtisnewbie/mini-fstore/api"
	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
	"gorm.io/gorm"
)

const (
	FileTypeFile = "FILE" // file
	FileTypeDir  = "DIR"  // directory

	LDelN = 0 // normal file
	LDelY = 1 // file marked deleted

	PDelN = 0 // file marked deleted, the actual deletion is not yet processed
	PDelY = 1 // file finally deleted, may be removed from disk or move to somewhere else

	VfolderOwner   = "OWNER"   // owner of the vfolder
	VfolderGranted = "GRANTED" // granted access to the vfolder
)

var (
	_imageSuffix = miso.NewSet[string]()
)

func init() {
	_imageSuffix.AddAll([]string{"jpeg", "jpg", "gif", "png", "svg", "bmp", "webp", "apng", "avif"})
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
	FileType       string     `json:"fileType"`
	UpdateTime     miso.ETime `json:"updateTime"`
	ParentFileName string     `json:"parentFileName"`
	SensitiveMode  string     `json:"sensitiveMode"`
	ThumbnailToken string     `json:"thumbnailToken"`
	Thumbnail      string     `json:"-"`
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

type ParentFileInfo struct {
	Zero     bool   `json:"-"`
	FileKey  string `json:"fileKey"`
	Filename string `json:"fileName"`
}

type FileDownloadInfo struct {
	FileId         int
	Name           string
	IsLogicDeleted int
	FileType       string
	FstoreFileId   string
	UploaderNo     string
}

func (f *FileDownloadInfo) Deleted() bool {
	return f.IsLogicDeleted == LDelY
}

func (f *FileDownloadInfo) IsFile() bool {
	return f.FileType == FileTypeFile
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
	UploaderNo       string // uploader's user_no
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
	return f.Ownership == VfolderOwner
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

func listFilesInVFolder(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (miso.PageRes[ListedFile], error) {
	qpp := miso.QueryPageParam[ListedFile]{
		ReqPage: req.Page,
		AddSelectQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time, fi.thumbnail`)
		},
		GetBaseQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Table("file_info fi").
				Joins("LEFT JOIN file_vfolder fv ON (fi.uuid = fv.uuid AND fv.is_del = 0)").
				Joins("LEFT JOIN user_vfolder uv ON (fv.folder_no = uv.folder_no AND uv.is_del = 0)").
				Where("uv.user_no = ? AND uv.folder_no = ?", user.UserNo, req.FolderNo)
		},
	}
	return qpp.ExecPageQuery(rail, tx)
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
	Sensitive  *bool       `json:"sensitive"`
}

func ListFiles(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (miso.PageRes[ListedFile], error) {
	var res miso.PageRes[ListedFile]
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

	for i, f := range res.Payload {
		if f.Thumbnail != "" {
			tkn, err := GetFstoreTmpToken(rail, f.Thumbnail, "")
			if err != nil {
				rail.Errorf("failed to generate file token for thumbnail: %v, %v", f.Thumbnail, err)
			} else {
				res.Payload[i].ThumbnailToken = tkn
			}
		}

	}

	return res, e
}

func listFilesForTags(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (miso.PageRes[ListedFile], error) {
	qpp := miso.QueryPageParam[ListedFile]{
		AddSelectQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time, fi.sensitive_mode, fi.thumbnail`)
		},
		GetBaseQuery: func(t *gorm.DB) *gorm.DB {
			t = t.Table("file_info fi").
				Joins("LEFT JOIN file_tag ft ON (ft.user_id = ? AND fi.id = ft.file_id)", user.UserId).
				Joins("LEFT JOIN tag t ON (ft.tag_id = t.id)").
				Order("fi.id desc")

			if req.Filename != nil && *req.Filename != "" {
				t = t.Where("fi.name LIKE ?", "%"+*req.Filename+"%")
			}

			if req.Sensitive != nil && *req.Sensitive {
				t = t.Where("fi.sensitive_mode = 'N'")
			}

			t = t.Where("fi.uploader_id = ?", user.UserId).
				Where("fi.file_type = 'FILE'").
				Where("fi.is_del = 0").
				Where("fi.is_logic_deleted = 0").
				Where("ft.is_del = 0").
				Where("t.is_del = 0").
				Where("t.name = ?", *req.TagName)
			return t
		},
	}
	return qpp.ExecPageQuery(rail, tx)
}

func listFilesSelective(rail miso.Rail, tx *gorm.DB, req ListFileReq, user common.User) (miso.PageRes[ListedFile], error) {
	/*
	   If parentFile is empty, and filename are not searched, then we only return the top level file or dir.
	   The query for tags will ignore parent_file param, so it's working fine
	*/
	if (req.ParentFile == nil || *req.ParentFile == "") && (req.Filename == nil || *req.Filename == "") {
		req.Filename = nil
		var pf = ""
		req.ParentFile = &pf // top-level file/dir
	}

	qpq := miso.QueryPageParam[ListedFile]{
		ReqPage: req.Page,
		AddSelectQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Select(`fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.uploader_id,
			fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time, fi.sensitive_mode, fi.thumbnail`)
		},
		GetBaseQuery: func(tx *gorm.DB) *gorm.DB {
			tx = tx.Table("file_info fi").
				Where("fi.uploader_id = ?", user.UserId).
				Where("fi.is_logic_deleted = 0 AND fi.is_del = 0").
				Order("fi.file_type asc, fi.id desc")

			if req.ParentFile != nil {
				tx = tx.Where("fi.parent_file = ?", *req.ParentFile)
			}

			if req.Filename != nil && *req.Filename != "" {
				tx = tx.Where("fi.name LIKE ?", "%"+*req.Filename+"%")
			}

			if req.FileType != nil && *req.FileType != "" {
				tx = tx.Where("fi.file_type = ?", *req.FileType)
			}

			if req.Sensitive != nil && *req.Sensitive {
				tx = tx.Where("fi.sensitive_mode = 'N'")
			}

			return tx
		},
	}

	return qpq.ExecPageQuery(rail, tx)
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
		Where("file_type = ?", FileTypeFile).
		Where("is_logic_deleted = ?", LDelN).
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

func findFile(rail miso.Rail, tx *gorm.DB, fileKey string) (FileInfo, error) {
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
	rail.Infof("Making dir, req: %+v", req)

	var dir FileInfo
	dir.Name = req.Name
	dir.Uuid = miso.GenIdP("ZZZ")
	dir.SizeInBytes = 0
	dir.FileType = FileTypeDir

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

func MoveFileToDir(rail miso.Rail, db *gorm.DB, req MoveIntoDirReq, user common.User) error {
	if req.Uuid == "" || req.Uuid == req.ParentFileUuid {
		return nil
	}

	// lock the file
	flock := fileLock(rail, req.Uuid)
	if err := flock.Lock(); err != nil {
		return err
	}
	defer flock.Unlock()

	fi, err := findFile(rail, db, req.Uuid)
	if err != nil {
		return miso.NewErr("File not found", "failed to find file, uuid: %v, %v", req.Uuid, err)
	}
	if fi.ParentFile == req.ParentFileUuid {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {

		// lock directory if necessary, if parentFileUuid is empty, the file is moved out of a directory
		if req.ParentFileUuid != "" {
			pflock := fileLock(rail, req.ParentFileUuid)
			if err := pflock.Lock(); err != nil {
				return err
			}
			defer pflock.Unlock()

			pr, e := findFile(rail, tx, req.ParentFileUuid)
			if e != nil {
				return fmt.Errorf("failed to find parentFile, %v", e)
			}
			rail.Debugf("parentFile: %+v", pr)

			if pr.IsZero() {
				return fmt.Errorf("perentFile not found, parentFileKey: %v", req.ParentFileUuid)
			}

			if pr.UploaderNo != user.UserNo {
				return miso.NewErr("You are not the owner of this directory")
			}

			if pr.FileType != FileTypeDir {
				return miso.NewErr("Target file is not a directory")
			}

			if pr.IsLogicDeleted != LDelN {
				return miso.NewErr("Target file deleted")
			}

			newSize := pr.SizeInBytes + fi.SizeInBytes
			err := tx.Exec("UPDATE file_info SET size_in_bytes = ?, update_by = ?, update_time = ? WHERE uuid = ?",
				newSize, user.Username, time.Now(), req.ParentFileUuid).Error
			if err != nil {
				return fmt.Errorf("failed to updated dir's size, dir: %v, %v", req.ParentFileUuid, err)
			}
			rail.Infof("updated dir %v size to %v", req.ParentFileUuid, newSize)

		}

		if !miso.IsBlankStr(fi.ParentFile) {

			// calculate the dir size asynchronously
			if err := miso.PubEventBus(rail, CalcDirSizeEvt{
				FileKey: fi.ParentFile,
			}, VfmCalcDirSizeEventBus); err != nil {
				rail.Errorf("failed to send CalcDirSizeEvt, fileKey: %v, %v", fi.ParentFile, err)
			}
		}

		return tx.Exec("UPDATE file_info SET parent_file = ?, update_by = ?, update_time = ? WHERE uuid = ?",
			req.ParentFileUuid, user.Username, time.Now(), req.Uuid).
			Error
	})
}

func _saveFile(rail miso.Rail, tx *gorm.DB, f FileInfo, user common.User) error {
	uname := user.Username
	now := miso.Now()

	f.IsLogicDeleted = LDelN
	f.IsPhysicDeleted = PDelN
	f.UploaderId = user.UserId
	f.UploaderName = uname
	f.CreateBy = uname
	f.UploadTime = now
	f.CreateTime = now
	f.UploaderNo = user.UserNo

	err := tx.Table("file_info").
		Omit("id", "update_time", "update_by").
		Create(&f).Error
	if err == nil {
		rail.Infof("Saved file %+v", f)
	}
	return err
}

func fileLock(rail miso.Rail, fileKey string) *miso.RLock {
	return miso.NewRLock(rail, "file:uuid:"+fileKey)
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
				Ownership:  VfolderOwner,
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

func ShareVFolder(rail miso.Rail, tx *gorm.DB, sharedTo vault.UserInfo, folderNo string, user common.User) error {
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
			Ownership:  VfolderGranted,
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
			Exec("UPDATE user_vfolder SET is_del = 1, update_by = ? WHERE folder_no = ? AND user_no = ? AND ownership = 'GRANTED'",
				user.Username, req.FolderNo, req.UserNo).
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
	Sync     bool     `json:"-"`
}

func NewVFolderLock(rail miso.Rail, folderNo string) *miso.RLock {
	return miso.NewRLock(rail, "vfolder:"+folderNo)
}

func HandleAddFileToVFolderEvent(rail miso.Rail, tx *gorm.DB, evt AddFileToVfolderEvent) error {
	lock := NewVFolderLock(rail, evt.FolderNo)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	var vfo VFolderWithOwnership
	var e error
	if vfo, e = findVFolder(rail, tx, evt.FolderNo, evt.UserNo); e != nil {
		return fmt.Errorf("failed to findVFolder, folderNo: %v, userNo: %v, %v", evt.FolderNo, evt.UserNo, e)
	}
	if !vfo.IsOwner() {
		return miso.NewErr("Operation not permitted")
	}

	distinct := miso.NewSet[string]()
	for _, fk := range evt.FileKeys {
		distinct.Add(fk)
	}

	filtered := miso.Distinct(evt.FileKeys)
	if len(filtered) < 1 {
		return nil
	}

	now := miso.Now()
	username := evt.Username
	doAddFileToVfolder := func(rail miso.Rail, folderNo string, fk string) error {
		var id int
		var err error
		err = tx.Select("id").
			Table("file_vfolder").
			Where("folder_no = ? AND uuid = ?", folderNo, fk).
			Where("is_del = 0").
			Scan(&id).
			Error
		if err != nil {
			return fmt.Errorf("failed to query file_vfolder record, %v", err)
		}
		if id > 0 {
			return nil
		}

		fvf := FileVFolder{FolderNo: folderNo, Uuid: fk, CreateTime: now, CreateBy: username}
		if err = tx.Table("file_vfolder").Omit("id", "update_by", "update_time").Create(&fvf).Error; err != nil {
			return fmt.Errorf("failed to save file_vfolder record, %v", err)
		}
		rail.Infof("added file.uuid: %v to vfolder: %v by %v", fk, folderNo, username)
		return nil
	}

	// add files to vfolder
	dirs := []FileInfo{}
	for fk := range distinct.Keys {
		var f FileInfo
		var e error

		f, e = findFile(rail, tx, fk)
		if e != nil {
			return e
		}
		if f.IsZero() || f.UploaderId != evt.UserId {
			continue
		}
		if f.FileType != FileTypeFile {
			dirs = append(dirs, f)
			continue
		}
		if e = doAddFileToVfolder(rail, evt.FolderNo, fk); e != nil {
			return fmt.Errorf("failed to doAddFileToVfolder, file.uuid: %v, %v", fk, e)
		}
	}

	// add files in dir to vfolder, but we only go one layer deep
	for _, dir := range dirs {
		var filesInDir []string
		var err error
		var page int = 1

		for {
			if filesInDir, err = ListFilesInDir(rail, tx, ListFilesInDirReq{
				FileKey: dir.Uuid,
				Limit:   500,
				Page:    page,
			}); err != nil {
				return fmt.Errorf("failed to list files in dir, dir.uuid: %v, %v", dir.Uuid, err)
			}

			if len(filesInDir) < 1 {
				break
			}

			for _, fk := range filesInDir {
				if !distinct.Add(fk) {
					continue
				}
				if err = doAddFileToVfolder(rail, evt.FolderNo, fk); err != nil {
					return fmt.Errorf("failed to doAddFileToVfolder, file.uuid: %v, %v", fk, e)
				}
			}
			page += 1
		}
	}
	return nil
}

func AddFileToVFolder(rail miso.Rail, tx *gorm.DB, req AddFileToVfolderReq, user common.User) error {

	if len(req.FileKeys) < 1 {
		return nil
	}

	vfo, e := findVFolder(rail, tx, req.FolderNo, user.UserNo)
	if e != nil {
		return e
	}
	if !vfo.IsOwner() {
		return miso.NewErr("Operation not permitted")
	}

	evt := AddFileToVfolderEvent{
		UserId:   user.UserId,
		Username: user.Username,
		UserNo:   user.UserNo,
		FolderNo: req.FolderNo,
		FileKeys: req.FileKeys,
	}
	err := miso.PubEventBus(rail, evt, VfmAddFileToVFolderEventBus)
	if err != nil {
		return fmt.Errorf("failed to publish AddFileToVfolderEvent, %+v, %v", evt, err)
	}
	return nil
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
			if f.FileType != FileTypeFile {
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
		unr, e := vault.FetchUsernames(rail, vault.FetchUsernameReq{UserNos: userNos})
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
	Id            int `json:"id" validation:"positive"`
	Name          string
	SensitiveMode string
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
	if r.SensitiveMode != "Y" && r.SensitiveMode != "N" {
		r.SensitiveMode = "N"
	}

	return tx.
		Exec("UPDATE file_info SET name = ?, sensitive_mode = ?, update_by = ? WHERE id = ? AND is_logic_deleted = 0 AND is_del = 0",
			r.Name, r.SensitiveMode, user.Username, r.Id).
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
	Filename         string `json:"filename"`
	FakeFstoreFileId string `json:"fstoreFileId"`
	ParentFile       string `json:"parentFile"`
}

func CreateFile(rail miso.Rail, tx *gorm.DB, r CreateFileReq, user common.User) error {
	fsf, e := fstore.FetchFileInfo(rail, fstore.FetchFileInfoReq{
		UploadFileId: r.FakeFstoreFileId,
	})
	if e != nil {
		if errors.Is(e, fstore.ErrFileNotFound) || errors.Is(e, fstore.ErrFileDeleted) {
			return miso.NewErr("File not found or deleted")
		}
		return fmt.Errorf("failed to fetch file info from fstore, %v", e)
	}
	if fsf.Status != FileStatusNormal {
		return miso.NewErr("File is deleted")
	}

	return SaveFileRecord(rail, tx, SaveFileReq{
		Filename:   r.Filename,
		Size:       fsf.Size,
		FileId:     fsf.FileId,
		ParentFile: r.ParentFile,
	}, user)
}

type SaveFileReq struct {
	Filename   string
	FileId     string
	Size       int64
	ParentFile string
}

func SaveFileRecord(rail miso.Rail, tx *gorm.DB, r SaveFileReq, user common.User) error {
	var f FileInfo
	f.Name = r.Filename
	f.Uuid = miso.GenIdP("ZZZ")
	f.FstoreFileId = r.FileId
	f.SizeInBytes = r.Size
	f.FileType = FileTypeFile

	if e := _saveFile(rail, tx, f, user); e != nil {
		return e
	}

	if r.ParentFile != "" {
		if e := MoveFileToDir(rail, tx, MoveIntoDirReq{Uuid: f.Uuid, ParentFileUuid: r.ParentFile}, user); e != nil {
			return e
		}
	}
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
	lock := fileLock(rail, req.Uuid)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

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

	if f.IsLogicDeleted == LDelY {
		return nil // deleted already
	}

	if f.FileType == FileTypeDir { // if it's dir make sure it's empty
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
}

func validateFileAccess(rail miso.Rail, tx *gorm.DB, fileKey string, userNo string) (FileDownloadInfo, error) {
	var f FileDownloadInfo

	t := tx.
		Select("fi.id 'file_id', fi.fstore_file_id, fi.name, fi.is_logic_deleted, fi.file_type, fi.uploader_no").
		Table("file_info fi").
		Where("fi.uuid = ? AND fi.is_del = 0", fileKey).
		Limit(1).
		Scan(&f)
	if t.Error != nil {
		return f, t.Error
	}
	if t.RowsAffected < 1 {
		return f, miso.NewErr("File not found")
	}
	if f.Deleted() {
		return f, miso.NewErr("File deleted")
	}
	if !f.IsFile() {
		return f, miso.NewErr("Downloading a directory is not supported")
	}

	// is uploader of the file
	permitted := f.UploaderNo == userNo

	// user may have access to the vfolder, which contains the file
	if !permitted {
		var uvid int
		e := tx.
			Select("ifnull(uv.id, 0) as id").
			Table("file_info fi").
			Joins("LEFT JOIN file_vfolder fv ON (fi.uuid = fv.uuid AND fv.is_del = 0)").
			Joins("LEFT JOIN user_vfolder uv ON (uv.user_no = ? AND uv.folder_no = fv.folder_no AND uv.is_del = 0)", userNo).
			Where("fi.id = ?", f.FileId).
			Limit(1).
			Scan(&uvid).Error
		if e != nil {
			return f, fmt.Errorf("failed to query user folder relation for file, id: %v, %v", f.FileId, e)
		}
		permitted = uvid > 0 // granted access to a folder that contains this file
	}

	if !permitted {
		return f, miso.NewErr("You are not permitted to access this file")
	}

	return f, nil
}

type GenerateTempTokenReq struct {
	FileKey string `json:"fileKey"`
}

func GenTempToken(rail miso.Rail, tx *gorm.DB, r GenerateTempTokenReq, user common.User) (string, error) {
	f, err := validateFileAccess(rail, tx, r.FileKey, user.UserNo)
	if err != nil {
		return "", fmt.Errorf("failed to validate file access, user: %+v, %w", user, err)
	}

	if f.FstoreFileId == "" {
		rail.Errorf("File %v doesn't have mini-fstore file_id", r.FileKey)
		return "", miso.NewErr("File cannot be downloaded, please contact system administrator")
	}

	return GetFstoreTmpToken(rail, f.FstoreFileId, f.Name)
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
	UploaderId   int    `json:"uploaderId"` // deprecated
	UploaderNo   string `json:"uploaderNo"`
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
	fir.UploaderNo = f.UploaderNo
	fir.UploaderName = f.UploaderName
	fir.IsDeleted = f.IsLogicDeleted == LDelY
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

type RemoveVFolderReq struct {
	FolderNo string
}

func RemoveVFolder(rail miso.Rail, tx *gorm.DB, user common.User, req RemoveVFolderReq) error {
	lock := NewVFolderLock(rail, req.FolderNo)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	var vfo VFolderWithOwnership
	var e error
	if vfo, e = findVFolder(rail, tx, req.FolderNo, user.UserNo); e != nil {
		return fmt.Errorf("failed to findVFolder, folderNo: %v, userNo: %v, %v", req.FolderNo, user.UserNo, e)
	}
	if !vfo.IsOwner() {
		return miso.NewErr("Operation not permitted")
	}
	if err := tx.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`UPDATE vfolder SET is_del = 1, update_by = ? WHERE folder_no = ?`, user.Username, req.FolderNo).Error
		if err != nil {
			return fmt.Errorf("failed to update vfolder, folderNo: %v, %v", req.FolderNo, err)
		}
		err = tx.Exec(`UPDATE user_vfolder SET is_del = 1, update_by = ? WHERE folder_no = ?`, user.Username, req.FolderNo).Error
		if err != nil {
			return fmt.Errorf("failed to update user_vfolder, folderNo: %v, %v", req.FolderNo, err)
		}
		err = tx.Exec(`UPDATE file_vfolder SET is_del = 1, update_by = ? WHERE folder_no = ?`, user.Username, req.FolderNo).Error
		if err != nil {
			return fmt.Errorf("failed to update file_vfolder, folderNo: %v, %v", req.FolderNo, err)
		}
		return nil
	}); err != nil {
		return err
	}

	rail.Infof("VFolder %v deleted by %v", req.FolderNo, user.Username)
	return nil
}

func BatchCalcDirSize(rail miso.Rail, db *gorm.DB) error {
	defer miso.TimeOp(rail, time.Now(), "BatchCalcDirSize")
	var limit int = 100
	var minId int = 0

	type TempFile struct {
		Id   int
		Uuid string
	}
	var files []TempFile

	for {
		t := db.Table("file_info").
			Select("uuid", "id").
			Where("file_type = 'DIR' AND is_del = 0 AND is_logic_deleted = 0").
			Where("id > ?", minId).
			Order("id asc").
			Scan(&files)
		if t.Error != nil {
			return fmt.Errorf("failed to list dir files, minId: %v, limit: %v, %v", minId, limit, t.Error)
		}

		if t.RowsAffected < 1 || len(files) < 1 {
			rail.Infof("BatchCalcDirSize finished, minId: %v, limit: %v", minId, limit)
			return nil
		}

		for _, f := range files {
			start := time.Now()
			if err := CalcDirSize(rail, f.Uuid, db); err != nil {
				return fmt.Errorf("failed to CalcDirSize for file: %v, %v", f.Uuid, err)
			}
			miso.TimeOp(rail, start, fmt.Sprintf("CalcDirSize, uuid: %v, id: %v", f.Uuid, f.Id))
		}

		minId = files[len(files)-1].Id
		rail.Infof("minId: %v", minId)
	}
}

func CalcDirSize(rail miso.Rail, fk string, db *gorm.DB) error {
	lock := fileLock(rail, fk)
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to lock, fileKey: %v, %w", fk, err)
	}
	defer lock.Unlock()

	var size int64
	err := db.Raw("SELECT IFNULL(SUM(size_in_bytes),0) FROM file_info WHERE parent_file = ? AND is_del = 0 AND is_logic_deleted = 0", fk).
		Scan(&size).
		Error
	if err != nil {
		return fmt.Errorf("failed to calculate dir size, fileKey: %v, %v", fk, err)
	}

	if err := db.Exec(`UPDATE file_info SET size_in_bytes = ? WHERE uuid = ?`, size, fk).Error; err != nil {
		return fmt.Errorf("failed to update dir's size, fileKey: %v, %v", fk, err)
	}

	return nil
}

type UnpackZipReq struct {
	FileKey       string // file key of the zip file
	ParentFileKey string // file key of current directory (not where the zip entries will be saved)
}

type UnpackZipExtra struct {
	FileKey       string // file key of the zip file
	ParentFileKey string // file key of the target directory
	UserId        int
	UserNo        string
	Username      string
}

func UnpackZip(rail miso.Rail, db *gorm.DB, user common.User, req UnpackZipReq) error {
	flock := fileLock(rail, req.FileKey)
	if err := flock.Lock(); err != nil {
		return err
	}
	defer flock.Unlock()

	fi, err := findFile(rail, db, req.FileKey)
	if err != nil {
		return miso.NewErr("File not found", "failed to find file, uuid: %v, %v", req.FileKey, err)
	}
	if fi.IsZero() {
		return miso.NewErr("File not found")
	}

	if fi.IsLogicDeleted == LDelY {
		return miso.NewErr("File is deleted")
	}

	if !strings.HasSuffix(strings.ToLower(fi.Name), ".zip") {
		return miso.NewErr("File is not a zip")
	}

	dir, err := MakeDir(rail, db, MakeDirReq{
		Name:       fi.Name + " unpacked " + time.Now().Format("20060102_150405"),
		ParentFile: req.ParentFileKey,
	}, user)
	if err != nil {
		return fmt.Errorf("failed to make directory before unpacking zip, %w", err)
	}

	extra, err := miso.WriteJson(UnpackZipExtra{
		FileKey:       req.FileKey,
		ParentFileKey: dir,
		UserId:        user.UserId,
		UserNo:        user.UserNo,
		Username:      user.Username,
	})
	if err != nil {
		return fmt.Errorf("failed to write json as extra, %w", err)
	}

	err = fstore.TriggerFileUnzip(rail, fstore.UnzipFileReq{
		FileId:          fi.FstoreFileId,
		ReplyToEventBus: VfmUnzipResultNotifyEventBus,
		Extra:           string(extra),
	})
	if err != nil {
		return fmt.Errorf("failed to TriggerFileUnZip, %w", err)
	}
	return nil
}

func HandleZipUnpackResult(rail miso.Rail, db *gorm.DB, evt fstore.UnzipFileReplyEvent) error {
	var extra UnpackZipExtra
	if err := miso.ParseJson([]byte(evt.Extra), &extra); err != nil {
		return fmt.Errorf("failed to unmarshal from extra, %v", err)
	}

	if len(evt.ZipEntries) < 1 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, ze := range evt.ZipEntries {
			err := SaveFileRecord(rail, tx, SaveFileReq{
				Filename:   ze.Name,
				FileId:     ze.FileId,
				Size:       ze.Size,
				ParentFile: extra.ParentFileKey,
			}, common.User{
				UserId:   extra.UserId,
				UserNo:   extra.UserNo,
				Username: extra.Username,
			})
			if err != nil {
				return fmt.Errorf("failed to save zip entry, entry: %v, %w", ze, err)
			}
		}
		return nil
	})
}