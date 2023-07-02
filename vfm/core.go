package vfm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/curtisnewbie/gocommon/bus"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/redis"
	"gorm.io/gorm"
)

const (
	comprImgProcBus   = "hammer.image.compress.processing"
	comprImgNotifyBus = "hammer.image.compress.notification"

	USER_GROUP_PUBLIC  = 0
	USER_GROUP_PRIVATE = 1

	FILE_TYPE_FILE = "FILE"
	FILE_TYPE_DIR  = "DIR"

	FOWNERSHIP_ALL   = 0
	FOWNERSHIP_OWNER = 1

	FILE_LDEL_N = 0
	FILE_LDEL_Y = 1

	FILE_PDEL_N = 0
	FILE_PDEL_Y = 1

	VFOWNERSHIP_OWNER   = "OWNER"
	VFOWNERSHIP_GRANTED = "GRANTED"
)

var (
	_imageSuffix = common.NewSet[string]()
)

func init() {
	_imageSuffix.AddAll([]string{"jpeg", "jpg", "gif", "png", "svg", "bmp", "webp", "apng", "avif"})
}

type CompressImageEvent struct {
	FileKey string // file key from vfm
	FileId  string // file id from mini-fstore
}

type ValidateFileOwnerReq struct {
	FileKey string `form:"fileKey"`
	UserId  int    `form:"userId"`
}

type ListFilesInDirReq struct {
	FileKey string `form:"fileKey"`
	Limit   int    `form:"limit"`
	Page    int    `form:"page"`
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

type GenerateTempTokenReq struct {
	FileKey string `json:"fileKey"`
}

type FileTag struct {
	Id         int
	FileId     int
	TagId      int
	UserId     int
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
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
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

func (t Tag) IsZero() bool {
	return t.Id < 1
}

type FileVFolder struct {
	FolderNo   string
	Uuid       string
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

type VFolderBrief struct {
	FolderNo string `json:"folderNo"`
	Name     string `json:"name"`
}

type FileSharing struct {
	Id         int
	FileId     int
	UserId     int
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
	UpdateBy   string
	IsDel      common.IS_DEL
}

type ListedDir struct {
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type ListedFile struct {
	Id             int          `json:"id"`
	Uuid           string       `json:"uuid"`
	Name           string       `json:"name"`
	UploadTime     common.ETime `json:"uploadTime"`
	UploaderName   string       `json:"uploaderName"`
	SizeInBytes    int64        `json:"sizeInBytes"`
	UserGroup      int          `json:"userGroup"`
	IsOwner        bool         `json:"isOwner"`
	FileType       string       `json:"fileType"`
	UpdateTime     common.ETime `json:"updateTime"`
	ParentFileName string       `json:"parentFileName"`
	ParentFile     string       `json:"-"`
	UploaderId     string       `json:"-"`
}

type MoveIntoDirReq struct {
	Uuid           string `json:"uuid" validation:"notEmpty"`
	ParentFileUuid string `json:"parentFileUuid"`
}

type MakeDirReq struct {
	ParentFile string `json:"parentFile"`                 // Key of parent file
	Name       string `json:"name" validation:"notEmpty"` // name of the directory
}

type GrantAccessReq struct {
	FileId    int    `json:"fileId" validation:"positive"`
	GrantedTo string `json:"grantedTo" validation:"notEmpty"`
}

type ListedFileSharing struct {
	Id         int          `json:"id"`
	UserId     int          `json:"userId"`
	Username   string       `json:"username"`
	CreateDate common.ETime `json:"CreateDate"`
	CreateBy   string       `json:"createBy"`
}

type ListGrantedAccessRes struct {
	Page    common.Paging       `json:"pagingVo"`
	Payload []ListedFileSharing `json:"payload"`
}

type ListGrantedAccessReq struct {
	Page   common.Paging `json:"pagingVo"`
	FileId int           `json:"fileId" validation:"positive"`
}

type RemoveGrantedAccessReq struct {
	FileId int `json:"fileId" validation:"positive"`
	UserId int `json:"userId" validation:"positive"`
}

type ListFileReq struct {
	Page       common.Paging `json:"pagingVo"`
	UserGroup  *int          `json:"userGroup"`
	Filename   *string       `json:"filename"`
	Ownership  *int          `json:"ownership"`
	TagName    *string       `json:"tagName"`
	FolderNo   *string       `json:"folderNo"`
	FileType   *string       `json:"fileType"`
	ParentFile *string       `json:"parentFile"`
}

type CreateFileReq struct {
	Filename         string   `json:"filename"`
	FakeFstoreFileId string   `json:"fstoreFileId"`
	UserGroup        int      `json:"userGroup"`
	Tags             []string `json:"tags"`
	ParentFile       string   `json:"parentFile"`
}

type DeleteFileReq struct {
	Uuid string `json:"uuid"`
}

type UpdateFileReq struct {
	Id        int    `json:"id" validation:"positive"`
	UserGroup int    `json:"userGroup"`
	Name      string `json:"name"`
}

type ListFileTagReq struct {
	Page   common.Paging `json:"pagingVo"`
	FileId int           `json:"fileId" validation:"positive"`
}

type ListFileTagRes struct {
	Page    common.Paging   `json:"pagingVo"`
	Payload []ListedFileTag `json:"payload"`
}

type ListedFileTag struct {
	Id         int          `json:"id"`
	Name       string       `json:"name"`
	CreateTime common.ETime `json:"createTime"`
	CreateBy   string       `json:"createBy"`
}

type TagFileReq struct {
	FileId  int    `json:"fileId" validation:"positive"`
	TagName string `json:"tagName" validation:"notEmpty"`
}

type UntagFileReq struct {
	FileId  int    `json:"fileId" validation:"positive"`
	TagName string `json:"tagName" validation:"notEmpty"`
}

type ListedVFolder struct {
	Id         int          `json:"id"`
	FolderNo   string       `json:"folderNo"`
	Name       string       `json:"name"`
	CreateTime common.ETime `json:"createTime"`
	CreateBy   string       `json:"createBy"`
	UpdateTime common.ETime `json:"updateTime"`
	UpdateBy   string       `json:"updateBy"`
	Ownership  string       `json:"ownership"`
}

type ListVFolderRes struct {
	Page    common.Paging   `json:"pagingVo"`
	Payload []ListedVFolder `json:"payload"`
}

type ListVFolderReq struct {
	Page common.Paging `json:"pagingVo"`
	Name string        `json:"name"`
}

type CreateVFolderReq struct {
	Name string `json:"name"`
}

type AddFileToVfolderReq struct {
	FolderNo string   `json:"folderNo"`
	FileKeys []string `json:"fileKeys"`
}

type RemoveFileFromVfolderReq struct {
	FolderNo string   `json:"folderNo"`
	FileKeys []string `json:"fileKeys"`
}

type ShareVfolderReq struct {
	FolderNo string `json:"folderNo"`
	Username string `json:"username"`
}

type RemoveGrantedFolderAccessReq struct {
	FolderNo string `json:"folderNo"`
	UserNo   string `json:"userNo"`
}

type ListGrantedFolderAccessReq struct {
	Page     common.Paging `json:"pagingVo"`
	FolderNo string        `json:"folderNo"`
}

type ListedFolderAccess struct {
	UserNo     string       `json:"userNo"`
	Username   string       `json:"username"`
	CreateTime common.ETime `json:"createTime"`
}

type ListGrantedFolderAccessRes struct {
	Page    common.Paging        `json:"pagingVo"`
	Payload []ListedFolderAccess `json:"payload"`
}

type ListFilesRes struct {
	Page    common.Paging `json:"pagingVo"`
	Payload []ListedFile  `json:"payload"`
}

type ParentFileInfo struct {
	Zero     bool   `json:"-"`
	FileKey  string `json:"fileKey"`
	Filename string `json:"fileName"`
}

type FileDownloadInfo struct {
	FileId         int
	Name           string
	UserGroup      int
	UploaderId     int
	IsLogicDeleted int
	FileType       string
	FileSharingId  int
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
	UploadTime       common.ETime
	LogicDeleteTime  common.ETime
	PhysicDeleteTime common.ETime
	UserGroup        int
	FsGroupId        int
	FileType         string
	ParentFile       string
	CreateTime       common.ETime
	CreateBy         string
	UpdateTime       common.ETime
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
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
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
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
	UpdateBy   string
}

type UserVFolder struct {
	Id         int
	UserNo     string
	Username   string
	FolderNo   string
	Ownership  string
	GrantedBy  string // grantedBy (user_no)
	CreateTime common.ETime
	CreateBy   string
	UpdateTime common.ETime
	UpdateBy   string
}

func newListFilesInVFolderQuery(c common.ExecContext, r ListFileReq) *gorm.DB {
	return mysql.GetMySql().
		Table("file_info fi").
		Joins("left join file_vfolder fv on (fi.uuid = fv.uuid and fv.is_del = 0)").
		Joins("left join user_vfolder uv on (fv.folder_no = uv.folder_no and uv.is_del = 0)").
		Where("uv.user_no = ? and uv.folder_no = ?", c.UserNo(), r.FolderNo)
}

func listFilesInVFolder(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	offset := r.Page.GetOffset()
	limit := r.Page.GetLimit()

	var files []ListedFile

	t := newListFilesInVFolderQuery(c, r).
		Select("fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time").
		Offset(offset).
		Limit(limit).
		Scan(&files)

	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files in vfolder, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == c.User.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesInVFolderQuery(c, r).
		Select("count(fi.id)").
		Scan(&total)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files in vfolder, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: common.RespPage(r.Page, total)}, nil
}

type FileKeyName struct {
	Name string
	Uuid string
}

func queryFilenames(fileKeys []string) (map[string]string, error) {
	var rec []FileKeyName
	e := mysql.GetConn().
		Select("uuid, name").
		Table("file_info").
		Where("uuid in ?", fileKeys).
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

func ListFiles(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	var res ListFilesRes
	var e error
	if r.FolderNo != nil && *r.FolderNo != "" {
		res, e = listFilesInVFolder(c, r)
	} else if r.TagName != nil && *r.TagName != "" {
		res, e = listFilesForTags(c, r)
	} else {
		res, e = listFilesSelective(c, r)
	}
	if e != nil {
		return res, e
	}

	parentFileKeys := common.NewSet[string]()
	for _, f := range res.Payload {
		if f.ParentFile != "" {
			parentFileKeys.Add(f.ParentFile)
		}
	}

	if !parentFileKeys.IsEmpty() {
		keyName, e := queryFilenames(parentFileKeys.CopyKeys())
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

func listFilesForTags(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	var files []ListedFile
	t := newListFilesForTagsQuery(c, mysql.GetMySql(), r).
		Select("fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time").
		Order("fi.id desc").
		Offset(r.Page.GetOffset()).
		Limit(r.Page.GetLimit()).
		Scan(&files)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == c.User.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesForTagsQuery(c, mysql.GetMySql(), r).
		Select("count(*)").
		Scan(&total)

	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: common.RespPage(r.Page, total)}, nil
}

func listFilesSelective(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	/*
	   If parentFile is empty, and filename/userGroup are not searched, then we only return the top level file or dir.
	   The query for tags will ignore parent_file param, so it's working fine
	*/
	if (r.ParentFile == nil || *r.ParentFile == "") && (r.Filename == nil || *r.Filename == "") && r.UserGroup == nil {
		r.Filename = nil
		var pf = ""
		r.ParentFile = &pf // top-level file/dir
	}

	var files []ListedFile
	t := newListFilesSelectiveQuery(c, mysql.GetMySql(), r).
		Select("fi.id, fi.name, fi.parent_file, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time").
		Order("fi.file_type asc, fi.id desc").
		Offset(r.Page.GetOffset()).
		Limit(r.Page.GetLimit()).
		Scan(&files)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == c.User.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = newListFilesSelectiveQuery(c, mysql.GetMySql(), r).
		Select("count(*)").
		Scan(&total)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to count files, %v", t.Error)
	}

	return ListFilesRes{Payload: files, Page: common.RespPage(r.Page, total)}, nil
}

func newListFilesSelectiveQuery(c common.ExecContext, t *gorm.DB, r ListFileReq) *gorm.DB {
	userId, _ := strconv.Atoi(c.UserId())

	t = t.Table("file_info fi").
		Table("file_info fi").
		Joins("left join file_sharing fs on (fi.id = fs.file_id and fs.user_id = ?)", userId)

	if r.Ownership != nil && *r.Ownership == FOWNERSHIP_OWNER {
		t = t.Where("fi.uploader_id = ?", userId)
	} else {
		if r.UserGroup == nil {
			t = t.Where("fi.uploader_id = ? or fi.user_group = 0 or (fs.id is not null and fs.is_del = 0)", userId)
		} else {
			if *r.UserGroup == USER_GROUP_PUBLIC {
				t = t.Where("fi.user_group = 0")
			} else {
				t = t.Where("fi.uploader_id = ? and fi.user_group = 1 ", userId)
			}
		}
	}

	if r.ParentFile != nil {
		t = t.Where("fi.parent_file = ?", *r.ParentFile)
	}

	if r.Filename != nil && *r.Filename != "" {
		t = t.Where("fi.name like ?", "%"+*r.Filename+"%")
	}

	if r.FileType != nil && *r.FileType != "" {
		t = t.Where("fi.file_type = ?", *r.FileType)
	}

	t = t.Where("fi.is_logic_deleted = 0 and fi.is_del = 0")

	return t
}

func newListFilesForTagsQuery(c common.ExecContext, t *gorm.DB, r ListFileReq) *gorm.DB {
	userId, _ := strconv.Atoi(c.UserId())

	t = t.Table("file_info fi").
		Joins("left join file_tag ft on (ft.user_id = ? and fi.id = ft.file_id)", userId).
		Joins("left join tag t on (ft.tag_id = t.id)").
		Joins("left join file_sharing fs on (fi.id = fs.file_id and fs.user_id = ?)", userId)

	if r.Ownership != nil && *r.Ownership == FOWNERSHIP_OWNER {
		t = t.Where("fi.uploader_id = ?", userId)
	} else {
		if r.UserGroup == nil {
			t = t.Where("fi.uploader_id = ? or fi.user_group = 0 or (fs.id is not null and fs.is_del = 0)", userId)
		} else {
			if *r.UserGroup == USER_GROUP_PUBLIC {
				t = t.Where("fi.user_group = 0")
			} else {
				t = t.Where("fi.uploader_id = ? and fi.user_group = 1 ", userId)
			}
		}
	}

	if r.Filename != nil && *r.Filename != "" {
		t = t.Where("fi.name like ?", "%"+*r.Filename+"%")
	}

	t = t.Where("fi.file_type = 'FILE'").
		Where("fi.is_del = 0").
		Where("fi.is_logic_deleted = 0").
		Where("ft.is_del = 0").
		Where("t.is_del = 0").
		Where("t.name = ?", *r.TagName)

	return t
}

func FileExists(c common.ExecContext, filename string, parentFileKey string) (any, error) {
	var id int
	t := mysql.GetMySql().
		Select("id").
		Table("file_info").
		Where("parent_file = ?", parentFileKey).
		Where("name = ?", filename).
		Where("uploader_id = ?", c.UserId()).
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

func ListFileTags(c common.ExecContext, r ListFileTagReq) (ListFileTagRes, error) {
	var ftags []ListedFileTag

	t := newListFileTagsQuery(c, r).
		Select("*").
		Scan(&ftags)
	if t.Error != nil {
		return ListFileTagRes{}, fmt.Errorf("failed to list file tags for req: %v, %v", r, t.Error)
	}

	var total int
	t = newListFileTagsQuery(c, r).
		Select("count(*)").
		Scan(&total)
	if t.Error != nil {
		return ListFileTagRes{}, fmt.Errorf("failed to count file tags for req: %v, %v", r, t.Error)
	}

	return ListFileTagRes{Payload: ftags}, nil
}

func newListFileTagsQuery(c common.ExecContext, r ListFileTagReq) *gorm.DB {
	userId, _ := c.UserIdI()
	return mysql.GetMySql().
		Table("file_tag ft").
		Joins("left join tag t on ft.tag_id = t.id").
		Where("t.user_id = ? and ft.file_id = ? and ft.is_del = 0 and t.is_del = 0", userId, r.FileId)
}

func findFile(c common.ExecContext, fileKey string) (FileInfo, error) {
	var f FileInfo

	t := mysql.GetConn().
		Raw("select * from file_info where uuid = ? and is_del = 0", fileKey).
		Scan(&f)
	if t.Error != nil {
		return f, t.Error
	}
	return f, nil
}

func findFileKey(c common.ExecContext, id int) (string, error) {
	var fk string
	t := mysql.GetConn().
		Raw("select uuid from file_info where id = ?", id).
		Scan(&fk)
	if t.Error != nil {
		return fk, t.Error
	}
	return fk, nil
}

func findFileById(c common.ExecContext, id int) (FileInfo, error) {
	var f FileInfo

	t := mysql.GetConn().
		Raw("select * from file_info where id = ? and is_del = 0", id).
		Scan(&f)
	if t.Error != nil {
		return f, t.Error
	}
	return f, nil
}

func FindParentFile(c common.ExecContext, fileKey string) (ParentFileInfo, error) {
	var f FileInfo
	f, e := findFile(c, fileKey)
	if e != nil {
		return ParentFileInfo{}, e
	}
	if f.IsZero() {
		return ParentFileInfo{}, common.NewWebErr("File not found")
	}

	// dir is only visible to the uploader for now
	userId, _ := c.UserIdI()
	if f.UploaderId != userId {
		return ParentFileInfo{}, common.NewWebErr("Not permitted")
	}

	if f.ParentFile == "" {
		return ParentFileInfo{Zero: true}, nil
	}

	pf, e := findFile(c, f.ParentFile)
	if e != nil {
		return ParentFileInfo{}, e
	}

	return ParentFileInfo{FileKey: pf.Uuid, Filename: pf.Name, Zero: false}, nil
}

func MakeDir(c common.ExecContext, r MakeDirReq) (string, error) {

	var dir FileInfo
	dir.Name = r.Name
	dir.Uuid = common.GenIdP("ZZZ")
	dir.SizeInBytes = 0
	dir.UserGroup = USER_GROUP_PRIVATE
	dir.FileType = FILE_TYPE_DIR

	if e := _saveFile(c, dir); e != nil {
		return "", e
	}

	if r.ParentFile != "" {
		if e := MoveFileToDir(c, MoveIntoDirReq{Uuid: dir.Uuid, ParentFileUuid: r.ParentFile}); e != nil {
			return dir.Uuid, e
		}
	}

	return dir.Uuid, nil
}

func MoveFileToDir(c common.ExecContext, req MoveIntoDirReq) error {
	if req.Uuid == "" || req.ParentFileUuid == "" || req.Uuid == req.ParentFileUuid {
		return nil
	}

	// lock the file
	return _lockFileExec(c, req.Uuid, func() error {

		// lock directory
		return _lockFileExec(c, req.ParentFileUuid, func() error {

			pr, e := findFile(c, req.ParentFileUuid)
			if e != nil {
				return fmt.Errorf("failed to find parentFile, %v", e)
			}
			c.Log.Debugf("parentFile: %+v", pr)

			if pr.IsZero() {
				return fmt.Errorf("perentFile not found, parentFileKey: %v", req.ParentFileUuid)
			}

			userId, _ := c.UserIdI()
			if pr.UploaderId != userId {
				return common.NewWebErr("You are not the owner of this directory")
			}

			if pr.FileType != FILE_TYPE_DIR {
				return common.NewWebErr("Target file is not a directory")
			}

			if pr.IsLogicDeleted != FILE_LDEL_N {
				return common.NewWebErr("Target file deleted")
			}

			return mysql.GetConn().
				Exec("update file_info set parent_file = ?, update_by = ?, update_time = ? where uuid = ?",
					req.ParentFileUuid, c.Username(), time.Now(), req.Uuid).Error
		})
	})
}

func _saveFile(c common.ExecContext, f FileInfo) error {
	userId, _ := c.UserIdI()
	uname := c.Username()
	now := common.ETime(time.Now())

	f.IsLogicDeleted = FILE_LDEL_N
	f.IsPhysicDeleted = FILE_PDEL_N
	f.UploaderId = userId
	f.UploaderName = uname
	f.CreateBy = uname
	f.UploadTime = now
	f.CreateTime = now

	return mysql.GetConn().Table("file_info").Omit("id", "update_time", "update_by").Create(&f).Error
}

func _lockFileExec(c common.ExecContext, fileKey string, r redis.Runnable) error {
	return redis.RLockExec(c, "file:uuid:"+fileKey, r)
}

func _lockFileGet(c common.ExecContext, fileKey string, r redis.LRunnable) (any, error) {
	return redis.RLockRun(c, "file:uuid:"+fileKey, r)
}

func CreateVFolder(c common.ExecContext, r CreateVFolderReq) (string, error) {
	userNo := c.UserNo()

	v, e := redis.RLockRun(c, "vfolder:user:"+userNo, func() (any, error) {

		var id int
		t := mysql.GetConn().
			Select("vf.id").
			Table("vfolder vf").
			Joins("left join user_vfolder uv on (vf.folder_no = uv.folder_no)").
			Where("uv.user_no = ? and uv.ownership = 'OWNER'", userNo).
			Where("vf.name = ?", r.Name).
			Where("vf.is_del = 0 and uv.is_del = 0").
			Limit(1).
			Scan(&id)
		if t.Error != nil {
			return "", t.Error
		}
		if id > 0 {
			return "", common.NewWebErr(fmt.Sprintf("Found folder with same name ('%s')", r.Name))
		}

		folderNo := common.GenIdP("VFLD")

		e := mysql.GetConn().Transaction(func(tx *gorm.DB) error {

			ctime := common.ETime(time.Now())

			// for the vfolder
			vf := VFolder{Name: r.Name, FolderNo: folderNo, CreateTime: ctime, CreateBy: c.Username()}
			if e := tx.Omit("id", "update_by", "update_time").Table("vfolder").Create(&vf).Error; e != nil {
				return fmt.Errorf("failed to save VFolder, %v", e)
			}

			// for the user - vfolder relation
			uv := UserVFolder{
				FolderNo:   folderNo,
				UserNo:     userNo,
				Username:   c.Username(),
				Ownership:  VFOWNERSHIP_OWNER,
				GrantedBy:  userNo,
				CreateTime: ctime,
				CreateBy:   c.Username()}
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

func ListDirs(c common.ExecContext) ([]ListedDir, error) {
	userId, _ := c.UserIdI()

	var dirs []ListedDir
	e := mysql.GetConn().
		Select("id, uuid, name").
		Table("file_info").
		Where("uploader_id = ?", userId).
		Where("file_type = 'DIR'").
		Where("is_logic_deleted = 0").
		Where("is_del = 0").
		Scan(&dirs).Error
	return dirs, e
}

func GranteFileAccess(c common.ExecContext, grantedToUserId int, fileId int) error {
	userId, _ := c.UserIdI()
	if grantedToUserId == userId {
		return common.NewWebErr("You can't grant file access to yourself")
	}

	f, e := findFileById(c, fileId)
	if e != nil {
		c.Log.Errorf("Failed to find find by id, %v", e)
		return common.NewWebErr("Unable to find file")
	}

	if f.IsZero() {
		return common.NewWebErr("File not found")
	}

	if f.IsLogicDeleted != FILE_LDEL_N {
		return common.NewWebErr("File deleted already")
	}

	if f.UploaderId != userId {
		return common.NewWebErr("Only uploader can grant access to the file")
	}

	if f.FileType != FILE_TYPE_FILE {
		return common.NewWebErr("You can't not grant access to directory type files")
	}
	c.Log.Debugf("Granting file access to file: %v (%v) to user: %v", fileId, f.Uuid, grantedToUserId)

	return _lockFileExec(c, f.Uuid, func() error {
		var fs FileSharing
		t := mysql.GetConn().
			Select("id, is_del").
			Table("file_sharing").
			Where("file_id = ?", fileId).
			Where("user_id = ?", grantedToUserId).
			Scan(&fs)
		if t.Error != nil {
			return t.Error
		}

		if t.RowsAffected < 1 {
			fs = FileSharing{
				UserId:     grantedToUserId,
				FileId:     fileId,
				CreateTime: common.ETime(time.Now()),
				CreateBy:   c.Username(),
				IsDel:      common.IS_DEL_N,
			}
			return mysql.GetConn().
				Table("file_sharing").
				Omit("id", "update_by", "update_time").
				Create(&fs).Error
		}

		if fs.IsDel == common.IS_DEL_Y {
			return mysql.GetConn().Exec("update file_sharing set is_del = 0 where id = ?", fs.Id).Error
		}
		return nil
	})
}

func ListGrantedFileAccess(c common.ExecContext, r ListGrantedAccessReq) (ListGrantedAccessRes, error) {
	var lfs []ListedFileSharing
	e := newListGrantedFileAccessQuery(c, r).
		Select("id, user_id, create_time 'create_date', 'create_by'").
		Order("id desc").
		Scan(&lfs).Error
	if e != nil {
		return ListGrantedAccessRes{}, fmt.Errorf("failed to list file_sharing, req: %+v, %v", r, e)
	}

	var total int
	e = newListGrantedFileAccessQuery(c, r).
		Select("count(*)").
		Scan(&total).Error
	if e != nil {
		return ListGrantedAccessRes{}, fmt.Errorf("failed to count file_sharing, req: %+v, %v", r, e)
	}
	return ListGrantedAccessRes{Page: common.RespPage(r.Page, total), Payload: lfs}, nil
}

func newListGrantedFileAccessQuery(c common.ExecContext, r ListGrantedAccessReq) *gorm.DB {
	return mysql.GetConn().
		Table("file_sharing").
		Where("file_id = ?", r.FileId).
		Where("is_del = 0")
}

func findVFolder(c common.ExecContext, folderNo string, userNo string) (VFolderWithOwnership, error) {
	var vfo VFolderWithOwnership
	t := mysql.GetConn().
		Table("vfolder vf").
		Select("vf.*, uv.ownership").
		Joins("left join user_vfolder uv on (vf.folder_no = uv.folder_no and uv.is_del = 0)").
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

func _lockFolderExec(c common.ExecContext, folderNo string, r redis.Runnable) error {
	return redis.RLockExec(c, "vfolder:"+folderNo, r)
}

func _lockFolderGet(c common.ExecContext, folderNo string, r redis.LRunnable) (any, error) {
	return redis.RLockRun(c, "vfolder:"+folderNo, r)
}

func ShareVFolder(c common.ExecContext, sharedTo UserInfo, folderNo string) error {
	if c.UserNo() == sharedTo.UserNo {
		return nil
	}
	return _lockFolderExec(c, folderNo, func() error {
		vfo, e := findVFolder(c, folderNo, c.UserNo())
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return common.NewWebErr("Operation not permitted")
		}

		var id int
		e = mysql.GetConn().
			Table("user_vfolder").
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
			c.Log.Infof("VFolder is shared already, folderNo: %s, sharedTo: %s", folderNo, sharedTo.Username)
			return nil
		}

		uv := UserVFolder{
			FolderNo:   folderNo,
			UserNo:     sharedTo.UserNo,
			Username:   sharedTo.Username,
			Ownership:  VFOWNERSHIP_GRANTED,
			GrantedBy:  c.Username(),
			CreateTime: common.ETime(time.Now()),
			CreateBy:   c.Username()}
		if e := mysql.GetConn().Omit("id", "update_by", "update_time").Table("user_vfolder").Create(&uv).Error; e != nil {
			return fmt.Errorf("failed to save UserVFolder, %v", e)
		}
		c.Log.Infof("VFolder %s shared to %s by %s", folderNo, sharedTo.Username, c.Username())
		return nil
	})
}

func RemoveVFolderAccess(c common.ExecContext, r RemoveGrantedFolderAccessReq) error {
	if c.UserNo() == r.UserNo {
		return nil
	}
	return _lockFolderExec(c, r.FolderNo, func() error {
		vfo, e := findVFolder(c, r.FolderNo, c.UserNo())
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return common.NewWebErr("Operation not permitted")
		}
		return mysql.GetConn().
			Exec("delete from user_vfolder where folder_no = ? and user_no = ? and ownership = 'GRANTED'", r.FolderNo, r.UserNo).
			Error
	})
}

func ListVFolderBrief(c common.ExecContext) ([]VFolderBrief, error) {
	var vfb []VFolderBrief
	e := mysql.GetConn().
		Select("f.folder_no, f.name").
		Table("vfolder f").
		Joins("left join user_vfolder uv on (f.folder_no = uv.folder_no and uv.is_del = 0)").
		Where("f.is_del = 0 and uv.user_no = ? and uv.ownership = 'OWNER'", c.UserNo()).
		Scan(&vfb).Error
	return vfb, e
}

func AddFileToVFolder(c common.ExecContext, r AddFileToVfolderReq) error {
	if len(r.FileKeys) < 1 {
		return nil
	}

	return _lockFolderExec(c, r.FolderNo, func() error {

		vfo, e := findVFolder(c, r.FolderNo, c.UserNo())
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return common.NewWebErr("Operation not permitted")
		}

		s := common.NewSet[string]()
		for _, v := range r.FileKeys {
			s.Add(v)
		}
		if s.IsEmpty() {
			return nil
		}

		filtered := common.KeysOfSet(s)
		userId, _ := c.UserIdI()
		now := common.ETime(time.Now())
		username := c.Username()
		for _, fk := range filtered {
			f, e := findFile(c, fk)
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
			e = mysql.GetConn().
				Select("id").
				Table("file_vfolder").
				Where("folder_no = ? and uuid = ?", r.FolderNo, fk).
				Scan(&id).
				Error
			if e != nil {
				return fmt.Errorf("failed to query file_vfolder record, %v", e)
			}
			if id > 0 {
				continue // file already in vfolder
			}

			fvf := FileVFolder{FolderNo: r.FolderNo, Uuid: fk, CreateTime: now, CreateBy: username}
			e = mysql.GetConn().Table("file_vfolder").Omit("id", "update_by", "update_time").Create(&fvf).Error
			if e != nil {
				return fmt.Errorf("failed to save file_vfolder record, %v", e)
			}
		}
		return nil
	})
}

func RemoveFileFromVFolder(c common.ExecContext, r RemoveFileFromVfolderReq) error {
	if len(r.FileKeys) < 1 {
		return nil
	}

	return _lockFolderExec(c, r.FolderNo, func() error {

		vfo, e := findVFolder(c, r.FolderNo, c.UserNo())
		if e != nil {
			return e
		}
		if !vfo.IsOwner() {
			return common.NewWebErr("Operation not permitted")
		}

		s := common.NewSet[string]()
		for _, v := range r.FileKeys {
			s.Add(v)
		}
		if s.IsEmpty() {
			return nil
		}

		filtered := common.KeysOfSet(s)
		userId, _ := c.UserIdI()
		for _, fk := range filtered {
			f, e := findFile(c, fk)
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

			e = mysql.GetConn().Exec("delete from file_vfolder where folder_no = ? and uuid = ?", r.FolderNo, fk).Error
			if e != nil {
				return fmt.Errorf("failed to delete file_vfolder record, %v", e)
			}
		}

		return nil
	})
}

func ListVFolders(c common.ExecContext, r ListVFolderReq) (ListVFolderRes, error) {
	t := newListVFoldersQuery(c, r).
		Select("f.id, f.create_time, f.create_by, f.update_time, f.update_by, f.folder_no, f.name, uv.ownership").
		Order("f.id desc")

	var lvf []ListedVFolder
	if e := t.Scan(&lvf).Error; e != nil {
		return ListVFolderRes{}, fmt.Errorf("failed to query vfolder, req: %+v, %v", r, e)
	}

	var total int
	e := newListVFoldersQuery(c, r).
		Select("count(*)").
		Scan(&total).Error
	if e != nil {
		return ListVFolderRes{}, fmt.Errorf("failed to count vfolder, req: %+v, %v", r, e)
	}

	return ListVFolderRes{Page: common.RespPage(r.Page, total), Payload: lvf}, nil
}

func newListVFoldersQuery(c common.ExecContext, r ListVFolderReq) *gorm.DB {
	t := mysql.GetConn().
		Table("vfolder f").
		Joins("left join user_vfolder uv on (f.folder_no = uv.folder_no and uv.is_del = 0)").
		Where("f.is_del = 0 and uv.user_no = ?", c.UserNo())

	if r.Name != "" {
		t = t.Where("f.name like ?", "%"+r.Name+"%")
	}
	return t
}

func RemoveGrantedFileAccess(c common.ExecContext, r RemoveGrantedAccessReq) error {
	f, e := findFileById(c, r.FileId)
	if e != nil {
		return fmt.Errorf("failed to find file, %v", e)
	}

	if f.IsZero() {
		return common.NewWebErr("File not found")
	}

	if f.IsLogicDeleted != FILE_LDEL_N {
		return common.NewWebErr("File deleted already")
	}

	userId, _ := c.UserIdI()
	if f.UploaderId != userId {
		return common.NewWebErr("Not permitted")
	}

	return _lockFileExec(c, f.Uuid, func() error {
		// it was a logical delete in file-server, it now becomes a physical delete
		return mysql.GetConn().
			Exec("delete from file_sharing where file_id = ? and user_id = ? limit 1", r.FileId, r.UserId).
			Error
	})
}

func ListGrantedFolderAccess(c common.ExecContext, r ListGrantedFolderAccessReq) (ListGrantedFolderAccessRes, error) {
	userNo := c.UserNo()
	folderNo := r.FolderNo
	vfo, e := findVFolder(c, folderNo, userNo)
	if e != nil {
		return ListGrantedFolderAccessRes{}, e
	}
	if !vfo.IsOwner() {
		return ListGrantedFolderAccessRes{}, common.NewWebErr("Operation not permitted")
	}

	var l []ListedFolderAccess
	e = newListGrantedFolderAccessQuery(c, r).
		Select("user_no", "create_time", "username").
		Offset(r.Page.GetOffset()).
		Limit(r.Page.GetLimit()).
		Scan(&l).Error
	if e != nil {
		return ListGrantedFolderAccessRes{}, fmt.Errorf("failed to list granted folder access, req: %+v, %v", r, e)
	}

	var total int
	e = newListGrantedFolderAccessQuery(c, r).
		Select("count(*)").
		Scan(&total).Error
	if e != nil {
		return ListGrantedFolderAccessRes{}, fmt.Errorf("failed to count granted folder access, req: %+v, %v", r, e)
	}

	userNos := []string{}
	for _, p := range l {
		if p.Username == "" {
			userNos = append(userNos, p.UserNo)
		}
	}

	if len(userNos) > 0 { // since v0.0.4 this is not needed anymore, but we keep it here for backward compatibility
		unr, e := FetchUsernames(c, FetchUsernamesReq{UserNos: userNos})
		if e != nil {
			c.Log.Errorf("Failed to fetch usernames, %v", e)
		} else {
			for i, p := range l {
				if name, ok := unr.UserNoToUsername[p.UserNo]; ok {
					p.Username = name
					l[i] = p
				}
			}
		}
	}

	return ListGrantedFolderAccessRes{Payload: l, Page: common.RespPage(r.Page, total)}, nil
}

func newListGrantedFolderAccessQuery(c common.ExecContext, r ListGrantedFolderAccessReq) *gorm.DB {
	return mysql.GetConn().
		Table("user_vfolder").
		Where("folder_no = ? and ownership = 'GRANTED' and is_del = 0", r.FolderNo)
}

func UpdateFile(c common.ExecContext, r UpdateFileReq) error {
	f, e := findFileById(c, r.Id)
	if e != nil {
		return e
	}
	if f.IsZero() {
		return common.NewWebErr("File not found")
	}

	// dir is only visible to the uploader for now
	userId, _ := c.UserIdI()
	if f.UploaderId != userId {
		return common.NewWebErr("Not permitted")
	}

	// Directory is by default private, and it's not allowed to update it
	if f.FileType == FILE_TYPE_DIR && r.UserGroup != f.UserGroup {
		return common.NewWebErr("Updating directory's UserGroup is not allowed")
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return common.NewWebErr("Name can't be empty")
	}

	return mysql.GetConn().
		Exec("update file_info set user_group = ?, name = ? where id = ? and is_logic_deleted = 0 and is_del = 0", r.UserGroup, r.Name, r.Id).Error
}

func ListAllTags(c common.ExecContext) ([]string, error) {
	var l []string
	userId, _ := c.UserIdI()
	e := mysql.GetConn().
		Raw("select t.name from tag t where t.user_id = ? and t.is_del = 0", userId).
		Scan(&l).
		Error

	return l, e
}

func TagFile(c common.ExecContext, req TagFileReq) error {
	req.TagName = strings.TrimSpace(req.TagName)
	userId, _ := c.UserIdI()
	return _lockFileTagExec(c, userId, req.TagName, func() error {

		// find the tag first, and create one for current user if necessary
		tagId, e := tryCreateTag(c, userId, req.TagName)
		if e != nil {
			return e
		}
		if tagId < 1 {
			return fmt.Errorf("tagId illegal, shouldn't be less than 1")
		}

		// check if it's already tagged
		ft, e := findFileTag(c, req.FileId, tagId)
		if e != nil {
			return e
		}

		if ft.IsZero() {
			ft = FileTag{UserId: userId, FileId: req.FileId, TagId: tagId, CreateTime: common.ETime(time.Now()), CreateBy: c.Username()}
			return mysql.GetConn().
				Table("file_tag").Omit("id", "update_time", "update_by").Create(&ft).Error
		}

		if ft.IsDel == common.IS_DEL_Y {
			return mysql.GetConn().
				Exec("update file_tag set is_del = 0, update_time = ?, update_by = ? where id = ?", time.Now(), c.Username(), ft.Id).
				Error
		}

		return nil
	})
}

func tryCreateTag(c common.ExecContext, userId int, tagName string) (int, error) {
	t, e := findTag(c, userId, tagName)
	if e != nil {
		return 0, fmt.Errorf("failed to find tag, userId: %v, tagName: %v, %e", userId, tagName, e)
	}

	if t.IsZero() {
		t = Tag{Name: tagName, UserId: userId, CreateBy: c.Username(), CreateTime: common.ETime(time.Now())}
		e := mysql.GetConn().Table("tag").Omit("id", "update_time", "update_by").Create(&t).Error
		if e != nil {
			return 0, fmt.Errorf("failed to create tag, userId: %v, tagName: %v, %e", userId, tagName, e)
		}
		return t.Id, nil
	}

	if t.IsDel == common.IS_DEL_Y {
		e := mysql.GetConn().Exec("update tag set is_del = 0, update_time = ?, update_by = ? where id = ?", time.Now(), c.Username(), t.Id).Error
		if e != nil {
			return 0, fmt.Errorf("failed to update tag, id: %v, %e", t.Id, e)
		}
	}

	return t.Id, nil
}

func findTag(c common.ExecContext, userId int, tagName string) (Tag, error) {
	var t Tag
	e := mysql.GetConn().
		Raw("select * from tag where user_id = ? and name = ?", userId, tagName).
		Scan(&t).Error
	return t, e
}

func findFileTag(c common.ExecContext, fileId int, tagId int) (FileTag, error) {
	var ft FileTag
	e := mysql.GetConn().
		Raw("select * from file_tag where file_id = ? and tag_id = ?", fileId, tagId).
		Scan(&ft).Error
	return ft, e
}

func UntagFile(c common.ExecContext, req UntagFileReq) error {
	req.TagName = strings.TrimSpace(req.TagName)
	userId, _ := c.UserIdI()
	return _lockFileTagExec(c, userId, req.TagName, func() error {
		// each tag is bound to a specific user
		tag, e := findTag(c, userId, req.TagName)
		if e != nil {
			return e
		}
		if tag.IsZero() {
			c.Log.Infof("Tag for '%v' doesn't exist, unable to untag file", req.TagName)
			return nil // tag doesn't exist
		}

		fileTag, e := findFileTag(c, req.FileId, tag.Id)
		if e != nil {
			return e
		}

		if fileTag.IsZero() || fileTag.IsDel == common.IS_DEL_Y {
			c.Log.Infof("FileTag for file_id: %d, tag_id: %d, doesn't exist", req.FileId, tag.Id)
			return nil
		}

		return mysql.GetConn().Transaction(func(tx *gorm.DB) error {
			// it was a logic delete in file-server, it now becomes a physical delete
			e = tx.Exec("delete from file_tag where id = ?", fileTag.Id).Error
			if e != nil {
				return fmt.Errorf("failed to update file_tag, %v", e)
			}

			c.Log.Infof("Untagged file, file_id: %d, tag_name: %s", req.FileId, req.TagName)

			/*
			   check if the tag is still associated with other files, if not, we remove it
			   remember, the tag is bound for a specific user only, so this doesn't affect
			   other users
			*/
			var anyFileTagId int
			e = tx.Table("file_tag").
				Where("tag_id = ? and is_del = 0", tag.Id).
				Limit(1).
				Scan(&anyFileTagId).
				Error
			if e != nil {
				return e
			}

			if anyFileTagId < 1 {
				// it was a logic delete in file-server, it now becomes a physical delete
				return tx.Exec("delete from tag where id = ?", tag.Id).Error
			}
			return nil
		})
	})
}

func _lockFileTagExec(c common.ExecContext, userId int, tagName string, run redis.Runnable) error {
	return redis.RLockExec(c, fmt.Sprintf("file:tag:uid:%d:name:%s", userId, tagName), run)
}

func CreateFile(c common.ExecContext, r CreateFileReq) error {
	fsf, e := FetchFstoreFileInfo(c, "", r.FakeFstoreFileId)
	if e != nil {
		return fmt.Errorf("failed to fetch file info from fstore, %v", e)
	}
	if fsf.IsZero() || fsf.Status != FS_STATUS_NORMAL {
		return common.NewWebErr("File not found or deleted")
	}

	var f FileInfo
	f.Name = r.Filename
	f.Uuid = common.GenIdP("ZZZ")
	f.FstoreFileId = fsf.FileId
	f.SizeInBytes = fsf.Size
	f.UserGroup = USER_GROUP_PRIVATE
	f.FileType = FILE_TYPE_FILE

	if e := _saveFile(c, f); e != nil {
		return e
	}

	if r.ParentFile != "" {
		if e := MoveFileToDir(c, MoveIntoDirReq{Uuid: f.Uuid, ParentFileUuid: r.ParentFile}); e != nil {
			return e
		}
	}

	// TODO: Since v0.0.4, this is based on event-pump binlog event
	// if isImage(f.Name) {
	// 	if e := bus.SendToEventBus(CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcBus); e != nil {
	// 		c.Log.Errorf("Failed to send CompressImageEvent, uuid: %v, %v", f.Uuid, e)
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

func DeleteFile(c common.ExecContext, r DeleteFileReq) error {
	return _lockFileExec(c, r.Uuid, func() error {
		f, e := findFile(c, r.Uuid)
		if e != nil {
			return fmt.Errorf("unable to find file, uuid: %v, %v", r.Uuid, e)
		}
		if f.IsZero() {
			return common.NewWebErr("File not found")
		}

		userId, _ := c.UserIdI()
		if f.UploaderId != userId {
			return common.NewWebErr("Not permitted")
		}

		if f.IsLogicDeleted == FILE_LDEL_Y {
			return nil // deleted already
		}

		if f.FileType == FILE_TYPE_DIR { // if it's dir make sure it's empty
			var anyId int
			e := mysql.GetConn().
				Select("id").
				Table("file_info").
				Where("parent_file = ? and is_logic_deleted = 0 and is_del = 0", r.Uuid).
				Limit(1).
				Scan(&anyId).Error
			if e != nil {
				return fmt.Errorf("failed to count files in dir, uuid: %v, %v", r.Uuid, e)
			}
			if anyId > 0 {
				return common.NewWebErr("Directory is not empty, unable to delete it")
			}
		}

		if f.FstoreFileId != "" {
			if e := DeleteFstoreFile(c, f.FstoreFileId); e != nil {
				return fmt.Errorf("failed to delete fstore file, fileId: %v, %v", f.FstoreFileId, e)
			}
		}

		if f.Thumbnail != "" {
			if e := DeleteFstoreFile(c, f.Thumbnail); e != nil {
				return fmt.Errorf("failed to delete fstore file (thumbnail), fileId: %v, %v", f.Thumbnail, e)
			}
		}

		return mysql.GetConn().
			Exec("UPDATE file_info SET is_logic_deleted = 1, logic_delete_time = NOW() WHERE id = ? AND is_logic_deleted = 0", f.Id).
			Error
	})
}

func validateFileAccess(c common.ExecContext, fileKey string) (bool, FileDownloadInfo, error) {
	userId := c.UserIdInt()
	var f FileDownloadInfo
	e := mysql.GetConn().
		Select("fi.id 'file_id', fi.fstore_file_id, fi.name, fi.user_group, fi.uploader_id, fi.is_logic_deleted, fi.file_type, fs.id 'file_sharing_id'").
		Table("file_info fi").
		Joins("left join file_sharing fs on (fi.id = fs.file_id and fs.user_id = ?)", userId).
		Where("fi.uuid = ? and fi.is_del = 0", fileKey).
		Limit(1).
		Scan(&f).Error
	if e != nil {
		return false, f, e
	}

	if f.FileId < 1 {
		return false, f, common.NewWebErr("File not found")
	}
	if f.IsLogicDeleted == FILE_LDEL_Y {
		return false, f, common.NewWebErr("File deleted")
	}
	if f.FileType != FILE_TYPE_FILE {
		return false, f, common.NewWebErr("Downloading a directory is not supported")
	}

	var permitted bool = f.UserGroup == USER_GROUP_PUBLIC // publicly accessible
	if !permitted {
		permitted = f.UploaderId == c.UserIdInt() // owner of the file
	}
	if !permitted {
		permitted = f.FileSharingId > 0 // granted access to the file
	}
	if !permitted {
		var uvid int
		e := mysql.GetConn().
			Select("uv.id").
			Table("file_info fi").
			Joins("left join file_vfolder fv on (fi.uuid = fv.uuid and fv.is_del = 0)").
			Joins("left join user_vfolder uv on (uv.user_no = ? and uv.folder_no = fv.folder_no and uv.is_del = 0)", c.UserNo()).
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

func GenTempToken(c common.ExecContext, r GenerateTempTokenReq) (string, error) {
	ok, f, e := validateFileAccess(c, r.FileKey)
	if e != nil {
		return "", e
	}
	if !ok {
		return "", common.NewWebErr("Not permitted")
	}
	if f.FstoreFileId == "" {
		return "", common.NewWebErr("File cannot be downloaded, please contact system administrator")
	}

	t, e := GetFstoreTmpToken(c, f.FstoreFileId, f.Name)
	if e != nil {
		return "", e
	}
	return t, nil
}

func ListFilesInDir(c common.ExecContext, q ListFilesInDirReq) ([]string, error) {
	if q.Limit < 0 || q.Limit > 100 {
		q.Limit = 100
	}
	if q.Page < 1 {
		q.Page = 1
	}

	var fileKeys []string
	e := mysql.GetConn().
		Table("file_info").
		Select("uuid").
		Where("parent_file = ?", q.FileKey).
		Where("file_type = 'FILE'").
		Where("is_del = 0").
		Offset((q.Page - 1) * q.Limit).
		Limit(q.Limit).
		Scan(&fileKeys).Error
	return fileKeys, e
}

func FetchFileInfoInternal(c common.ExecContext, fileKey string) (FileInfoResp, error) {
	var fir FileInfoResp
	f, e := findFile(c, fileKey)
	if e != nil {
		return fir, e
	}
	if f.IsZero() {
		return fir, common.NewWebErr("File not found")
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

func ValidateFileOwner(c common.ExecContext, q ValidateFileOwnerReq) (bool, error) {
	var id int
	e := mysql.GetConn().
		Select("id").
		Table("file_info").
		Where("uuid = ?", q.FileKey).
		Where("uploader_id = ?", q.UserId).
		Where("is_logic_deleted = 0").
		Limit(1).
		Scan(&id).Error
	return id > 0, e
}

func ReactOnImageCompressed(c common.ExecContext, evt CompressImageEvent) error {
	return _lockFileExec(c, evt.FileKey, func() error {
		f, e := findFile(c, evt.FileKey)
		if e != nil {
			c.Log.Errorf("unable to find file, uuid: %v, %v", evt.FileKey, e)
			return nil
		}
		if f.IsZero() {
			c.Log.Errorf("File not found, uuid: %v", evt.FileKey)
			return nil
		}

		return mysql.GetConn().
			Exec("update file_info set thumbnail = ? where uuid = ?", evt.FileId, evt.FileKey).
			Error
	})
}

type FileCompressInfo struct {
	Id           int
	Name         string
	Uuid         string
	FstoreFileId string
}

func CompensateImageCompression(c common.ExecContext) error {

	limit := 500
	minId := 0

	for {
		var files []FileCompressInfo
		t := mysql.GetConn().
			Raw(`select id, name, uuid, fstore_file_id
			from file_info
			where id > ?
			and file_type = 'file'
			and is_logic_deleted = 0
			and thumbnail = ''
			order by id asc
			limit ?`, minId, limit).
			Scan(&files)
		if t.Error != nil {
			return t.Error
		}
		if t.RowsAffected < 1 || len(files) < 1 {
			return nil // the end
		}

		for _, f := range files {
			if isImage(f.Name) {
				if e := bus.SendToEventBus(c, CompressImageEvent{FileKey: f.Uuid, FileId: f.FstoreFileId}, comprImgProcBus); e != nil {
					c.Log.Errorf("Failed to send CompressImageEvent, uuid: %v, %v", f.Uuid, e)
				}
			}
		}

		minId = files[len(files)-1].Id
		c.Log.Infof("CompensateImageCompression, minId: %v", minId)
	}
}
