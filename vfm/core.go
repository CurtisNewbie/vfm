package vfm

import (
	"fmt"
	"strconv"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"gorm.io/gorm"
)

const (
	USER_GROUP_PUBLIC  = 0
	USER_GROUP_PRIVATE = 1

	FILE_TYPE_FILE = "FILE"
	FILE_TYPE_DIR  = "DIR"

	OWNERSHIP_ALL   = 0
	OWNERSHIP_OWNER = 1

	FILE_LDEL_N = 0
	FILE_LDEL_Y = 1
)

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
	UserGroup  int    `json:"userGroup"`                  // User Group
}

type GrantAccessReq struct {
	FileId    int    `json:"fileId" validation:"positive"`
	GrantedTo string `json:"grantedTo" validation:"notEmpty"`
}

type ListGrantedAcessReq struct {
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

func listFilesInVFolder(c common.ExecContext, r ListFileReq) (ListFilesRes, error) {
	userNo := c.UserNo()
	offset := r.Page.GetOffset()
	limit := r.Page.GetLimit()

	var files []ListedFile

	t := mysql.GetMySql().Raw(`
		select fi.id, fi.name, fi.uuid, fi.size_in_bytes, fi.user_group, fi.uploader_id, fi.uploader_name, fi.upload_time, fi.file_type, fi.update_time
		from file_info fi
		left join file_vfolder fv on (fi.uuid = fv.uuid and fv.is_del = 0)
		left join user_vfolder uv on (fv.folder_no = uv.folder_no and uv.is_del = 0)
		where uv.user_no = ? and uv.folder_no = ?
		limit ?, ?`, userNo, r.FolderNo, offset, limit).Scan(&files)
	if t.Error != nil {
		return ListFilesRes{}, fmt.Errorf("failed to list files in vfolder, %v", t.Error)
	}

	for i, f := range files {
		if f.UploaderId == c.User.UserId {
			files[i].IsOwner = true
		}
	}

	var total int
	t = mysql.GetMySql().Raw(`
		select count(*)
		from file_info fi
		left join file_vfolder fv on (fi.uuid = fv.uuid and fv.is_del = 0)
		left join user_vfolder uv on (fv.folder_no = uv.folder_no and uv.is_del = 0)
		where uv.user_no = ? and uv.folder_no = ?
		limit ?, ?`, userNo, r.FolderNo, offset, limit).Scan(&total)
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

	if r.Ownership != nil && *r.Ownership == OWNERSHIP_OWNER {
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

	if r.Ownership != nil && *r.Ownership == OWNERSHIP_OWNER {
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
