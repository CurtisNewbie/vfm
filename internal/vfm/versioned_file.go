package vfm

import (
	"fmt"

	"github.com/curtisnewbie/miso/middleware/user-vault/common"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

func CheckVerFile(rail miso.Rail, db *gorm.DB, fileKey string, userNo string) (*FileInfo, error) {
	f, err := findFile(rail, db, fileKey)
	if err != nil {
		return f, miso.NewErrf("File not found", "fileKey: %v, %v", fileKey, err)
	}
	if f == nil {
		return f, miso.NewErrf("File not found", "fileKey: %v", fileKey)
	}
	if f.UploaderNo != userNo {
		return f, miso.NewErrf("Not permitted")
	}
	if f.FileType != FileTypeFile {
		return f, miso.NewErrf("Illegal File Type")
	}
	if f.IsLogicDeleted == LDelY {
		return f, miso.NewErrf("File already deleted")
	}
	return f, nil
}

type ApiCreateVerFileReq struct {
	FileKey string `valid:"notEmpty" desc:"File Key"`
}

type ApiCreateVerFileRes struct {
	VerFileId string `desc:"Versioned File Id"`
}

// Hide normal file_info record, normally, versioned_file records are created based on newly updated file.
// We hide these records just to avoid user operations.
func HideVfmFile(rail miso.Rail, db *gorm.DB, fileKey string, user common.User) error {
	err := db.Exec(`
		UPDATE file_info SET hidden = 1, update_by = ?
		WHERE uuid = ?
	`, user.Username, fileKey).Error
	if err != nil {
		return fmt.Errorf("failed to hide file_info record, %v, %v", fileKey, err)
	}
	return nil
}

func CreateVerFile(rail miso.Rail, db *gorm.DB, req ApiCreateVerFileReq, user common.User) (ApiCreateVerFileRes, error) {
	var res ApiCreateVerFileRes
	f, err := CheckVerFile(rail, db, req.FileKey, user.UserNo)
	if err != nil {
		return res, err
	}
	verFileId := miso.GenIdP("verf_")

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
			INSERT INTO versioned_file (ver_file_id,file_key,name,size_in_bytes,uploader_no,uploader_name,upload_time,created_by)
			VALUES (?,?,?,?,?,?,?,?)
		`, verFileId, f.Uuid, f.Name, f.SizeInBytes, f.UploaderNo, f.UploaderName, miso.Now(), user.Username).Error
		if err != nil {
			return fmt.Errorf("failed to insert versioned_file, req: #%v, %w", req, err)
		}

		if err := SaveVerFileLog(rail, tx,
			SaveVerFileLogReq{VerFileId: verFileId, FileKey: req.FileKey, Username: user.Username}); err != nil {
			return err
		}

		if err := HideVfmFile(rail, db, req.FileKey, user); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return res, err
	}
	rail.Infof("Versioned file %v created for %v", verFileId, f.Uuid)
	return ApiCreateVerFileRes{VerFileId: verFileId}, nil
}

type ApiUpdateVerFileReq struct {
	VerFileId string `valid:"notEmpty" desc:"versioned file id"`
	FileKey   string `valid:"notEmpty" desc:"file key"`
}

type UpdateVerFileInf struct {
	FileKey    string
	UploaderNo string
	Deleted    bool
}

func NewVerFileLock(rail miso.Rail, verFileId string) *miso.RLock {
	return miso.NewRLockf(rail, "rlock:version_file:%v", verFileId)
}

func UpdateVerFile(rail miso.Rail, db *gorm.DB, req ApiUpdateVerFileReq, user common.User) error {
	lock := NewVerFileLock(rail, req.VerFileId)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	var uvf UpdateVerFileInf
	tx := db.Raw(`
		SELECT file_key,uploader_no,deleted
		FROM versioned_file
		WHERE ver_file_id = ?
	`, req.VerFileId).Scan(&uvf)
	if tx.Error != nil {
		return fmt.Errorf("failed to query versioned_file, req: %#v, %v", req, tx.Error)
	}
	if tx.RowsAffected < 1 {
		return miso.NewErrf("File not found", "ver_file_id not found, %v", req.VerFileId)
	}
	if uvf.UploaderNo != user.UserNo {
		return miso.NewErrf("Not permitted")
	}
	if uvf.Deleted {
		return miso.NewErrf("File already deleted")
	}
	if uvf.FileKey == req.FileKey {
		rail.Infof("Versioned File %v already using file_key: %v", req.VerFileId, req.FileKey)
		return nil
	}

	f, err := CheckVerFile(rail, db, req.FileKey, user.UserNo)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		svlr := SaveVerFileLogReq{VerFileId: req.VerFileId, FileKey: req.FileKey, Username: user.Username}
		if err := SaveVerFileLog(rail, tx, svlr); err != nil {
			return err
		}

		err = tx.Exec(`
			UPDATE versioned_file
			SET file_key = ?,
				name = ?,
				size_in_bytes = ?,
				upload_time = ?,
				updated_by = ?
			WHERE ver_file_id = ?
		`, f.Uuid, f.Name, f.SizeInBytes, miso.Now(), user.Username, req.VerFileId).Error
		if err != nil {
			return fmt.Errorf("failed to update versioned_file, req: #%v, %w", req, err)
		}

		if err := HideVfmFile(rail, db, req.FileKey, user); err != nil {
			return err
		}

		rail.Infof("Versioned file %v updated using %v", req.VerFileId, f.Uuid)

		return nil
	})

}

type ApiListVerFileReq struct {
	Paging miso.Paging `desc:"paging params"`
	Name   *string     `desc:"file name"`
}

type ApiListVerFileRes struct {
	VerFileId   string     `desc:"versioned file id"`
	Name        string     `desc:"file name"`
	FileKey     string     `desc:"file key"`
	SizeInBytes int64      `desc:"size in bytes"`
	UploadTime  miso.ETime `desc:"last upload time"`
	CreateTime  miso.ETime `desc:"create time of the versioned file record"`
	UpdateTime  miso.ETime `desc:"Update time of the versioned file record"`
	Deleted     bool       `desc:"whether version file record is deleted"`
	DeleteTime  miso.ETime `desc:"delete time of the versioned file record"`
}

func ListVerFile(rail miso.Rail, db *gorm.DB, req ApiListVerFileReq, user common.User) (miso.PageRes[ApiListVerFileRes], error) {
	return miso.NewPageQuery[ApiListVerFileRes]().
		WithPage(req.Paging).
		WithBaseQuery(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Table(`versioned_file`).
				Where(`uploader_no = ?`, user.UserNo).
				Where(`deleted = 0`)

			if req.Name != nil && *req.Name != "" {
				tx = tx.Where("match(name) against (? IN NATURAL LANGUAGE MODE)", *req.Name)
			}

			return tx
		}).
		WithSelectQuery(func(tx *gorm.DB) *gorm.DB {
			return tx.Select(`ver_file_id,name,file_key,size_in_bytes,upload_time,create_time,update_time,deleted,delete_time`)
		}).Exec(rail, db)
}

type SaveVerFileLogReq struct {
	VerFileId string
	FileKey   string
	Username  string
}

func SaveVerFileLog(rail miso.Rail, db *gorm.DB, req SaveVerFileLogReq) error {
	err := db.Exec(`INSERT INTO versioned_file_log (ver_file_id, file_key, created_by) VALUES (?,?,?)`,
		req.VerFileId, req.FileKey, req.Username).Error
	if err != nil {
		return fmt.Errorf("failed to save versioned_file_log, %#v, %v", req, err)
	}
	return err
}

type ApiDelVerFileReq struct {
	VerFileId string `desc:"Versioned File Id" valid:"notEmpty"`
}

func DelVerFile(rail miso.Rail, db *gorm.DB, req ApiDelVerFileReq, user common.User) error {
	lock := NewVerFileLock(rail, req.VerFileId)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	var uvf UpdateVerFileInf
	tx := db.Raw(`
		SELECT file_key,uploader_no,deleted
		FROM versioned_file
		WHERE ver_file_id = ?
	`, req.VerFileId).Scan(&uvf)
	if tx.Error != nil {
		return fmt.Errorf("failed to query versioned_file, req: %#v, %v", req, tx.Error)
	}
	if tx.RowsAffected < 1 {
		return miso.NewErrf("File not found", "ver_file_id not found, %v", req.VerFileId)
	}
	if uvf.UploaderNo != user.UserNo {
		return miso.NewErrf("Not permitted")
	}
	if uvf.Deleted {
		return miso.NewErrf("File already deleted")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`UPDATE versioned_file SET deleted = 1, updated_by = ?, delete_time = ? WHERE ver_file_id = ?`,
			user.Username, miso.Now(), req.VerFileId).Error
		if err != nil {
			return fmt.Errorf("failed to mark versioend_file deleted, %v, %w", req.VerFileId, err)
		}

		var fks []string
		if err := tx.Raw(`SELECT vf.file_key FROM versioned_file_log vf WHERE vf.ver_file_id = ?`, req.VerFileId).
			Scan(&fks).Error; err != nil {
			return fmt.Errorf("failed to query versioend_file_log, %v, %w", req.VerFileId, err)
		}

		for _, fk := range fks {
			if err := DeleteFile(rail, tx, DeleteFileReq{Uuid: fk}, user, nil); err != nil {
				return fmt.Errorf("failed to delete file in versioend_file_log, %v, %v, %w", req.VerFileId, fk, err)
			}
		}
		return nil
	})
}
