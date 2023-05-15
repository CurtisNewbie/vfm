package vfm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/redis"
	"gorm.io/gorm"
)

const (
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
	Id           int          `json:"id"`
	Uuid         string       `json:"uuid"`
	Name         string       `json:"name"`
	UploadTime   common.ETime `json:"uploadTime"`
	UploaderName string       `json:"uploaderName"`
	SizeInBytes  int64        `json:"sizeInBytes"`
	UserGroup    int          `json:"userGroup"`
	IsOwner      bool         `json:"isOwner"`
	FileType     string       `json:"fileType"`
	UpdateTime   common.ETime `json:"updateTime"`
	UploaderId   string       `json:"-"`
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

type ListVfolderReq struct {
	Page common.Paging `json:"pagingVo"`
	Name string        `json:"name"`
}

type CreateVfolderReq struct {
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

type ListFilesRes struct {
	Page    common.Paging `json:"pagingVo"`
	Payload []ListedFile  `json:"payload"`
}

type ParentFileInfo struct {
	Zero     bool   `json:"-"`
	FileKey  string `json:"fileKey"`
	Filename string `json:"fileName"`
}

type FileInfo struct {
	Id               int
	Name             string
	Uuid             string
	FstoreFileId     string
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
		Select("fi.*").
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

func ListFiles(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	// query files in vfolder
	if r.FolderNo != nil && *r.FolderNo != "" {
		return listFilesInVFolder(c, r)
	}

	// based on whether tagName is present, we use different queries
	if r.TagName != nil && *r.TagName != "" {
		return listFilesForTags(c, r)
	}

	// tagName is not present
	return listFilesSelective(c, r)
}

func listFilesForTags(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	var files []ListedFile
	t := newListFilesForTagsQuery(c, mysql.GetMySql(), r).
		Select("fi.id, fi.name, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time").
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
		Select("fi.id, fi.name, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time").
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

func CreateVFolder(c common.ExecContext, r CreateVfolderReq) (string, error) {
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
		if vfo.Ownership != VFOWNERSHIP_OWNER {
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
		if vfo.Ownership != VFOWNERSHIP_OWNER {
			return common.NewWebErr("Operation not permitted")
		}
		return mysql.GetConn().
			Exec("delete from user_vfolder where folder_no = ? and user_no = ? and ownership = 'GRANTED'", r.FolderNo, r.UserNo).
			Error
	})
}
