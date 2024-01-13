package vfm

import (
	"fmt"
	"time"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
	"gorm.io/gorm"
)

// User's access to a Gallery
type GalleryUserAccess struct {
	ID         int64
	GalleryNo  string
	UserNo     string
	CreateTime time.Time
	CreateBy   string
	UpdateTime time.Time
	UpdateBy   string
	IsDel      common.IS_DEL
}

type UpdateGUAIsDelCmd struct {
	GalleryNo string
	UserNo    string
	IsDelFrom common.IS_DEL
	IsDelTo   common.IS_DEL
	UpdateBy  string
}

func (GalleryUserAccess) TableName() string {
	return "gallery_user_access"
}

/* Check if user has access to the gallery */
func HasAccessToGallery(rail miso.Rail, tx *gorm.DB, userNo string, galleryNo string) (bool, error) {

	gallery, e := FindGallery(rail, tx, galleryNo)
	if e != nil {
		return false, e
	}

	if gallery.UserNo == userNo {
		return true, nil
	}

	// check if the user has access to the gallery
	userAccess, err := findGalleryAccess(rail, tx, userNo, galleryNo)
	if err != nil {
		return false, err
	}

	if userAccess == nil || common.IsDeleted(userAccess.IsDel) {
		return false, nil
	}

	return true, nil
}

// Assign user access to the gallery
func CreateGalleryAccess(rail miso.Rail, tx *gorm.DB, userNo string, galleryNo string, operator string) error {

	// check if the user has access to the gallery
	userAccess, err := findGalleryAccess(rail, tx, userNo, galleryNo)
	if err != nil {
		return err
	}

	if userAccess != nil && !common.IsDeleted(userAccess.IsDel) {
		return nil
	}

	var e error
	if userAccess == nil {
		e = createUserAccess(rail, tx, userNo, galleryNo, operator)
	} else {
		e = updateUserAccessIsDelFlag(rail, tx, &UpdateGUAIsDelCmd{
			UserNo:    userNo,
			GalleryNo: galleryNo,
			IsDelFrom: common.IS_DEL_N,
			IsDelTo:   common.IS_DEL_Y,
			UpdateBy:  operator,
		})
	}

	return e
}

/* find GalleryUserAccess, is_del flag is ignored */
func findGalleryAccess(rail miso.Rail, tx *gorm.DB, userNo string, galleryNo string) (*GalleryUserAccess, error) {
	// check if the user has access to the gallery
	var userAccess *GalleryUserAccess = &GalleryUserAccess{}

	tx = tx.Raw(`
		SELECT * FROM gallery_user_access
		WHERE gallery_no = ?
		AND user_no = ? AND is_del = 0`, galleryNo, userNo).Scan(&userAccess)

	if e := tx.Error; e != nil || tx.RowsAffected < 1 {
		if e != nil {
			return nil, fmt.Errorf("failed to find gallery_user_access, %v", e)
		}
		return nil, nil
	}

	return userAccess, nil
}

// Insert a new gallery_user_access record
func createUserAccess(rail miso.Rail, tx *gorm.DB, userNo string, galleryNo string, createdBy string) error {
	tx = tx.Exec(`INSERT INTO gallery_user_access (gallery_no, user_no, create_by) VALUES (?, ?, ?)`, galleryNo, userNo, createdBy)
	if e := tx.Error; e != nil {
		return e
	}

	return nil
}

// Update is_del of the record
func updateUserAccessIsDelFlag(rail miso.Rail, tx *gorm.DB, cmd *UpdateGUAIsDelCmd) error {
	tx = tx.Exec(`
	UPDATE gallery_user_access SET is_del = ?, update_by = ?
	WHERE gallery_no = ? AND user_no = ? AND is_del = ?`, cmd.IsDelTo, cmd.UpdateBy, cmd.GalleryNo, cmd.UserNo, cmd.IsDelFrom)

	if e := tx.Error; e != nil {
		return e
	}

	return nil
}

type RemoveGalleryAccessCmd struct {
	GalleryNo string `json:"galleryNo" validation:"notEmpty"`
	UserNo    string `json:"userNo" validation:"notEmpty"`
}

type ListGrantedGalleryAccessCmd struct {
	GalleryNo string `json:"galleryNo" validation:"notEmpty"`
	PagingVo  miso.Paging
}

type ListedGalleryAccessRes struct {
	Id         int
	GalleryNo  string
	UserNo     string
	Username   string
	CreateTime miso.ETime
}

type PermitGalleryAccessCmd struct {
	GalleryNo string `validation:"notEmpty"`
	Username  string `validation:"notEmpty"`
}

func ListedGrantedGalleryAccess(rail miso.Rail, tx *gorm.DB, req ListGrantedGalleryAccessCmd, user common.User) (miso.PageRes[ListedGalleryAccessRes], error) {
	gallery, e := FindGallery(rail, tx, req.GalleryNo)
	if e != nil {
		return miso.PageRes[ListedGalleryAccessRes]{}, e
	}
	if gallery.UserNo != user.UserNo {
		return miso.PageRes[ListedGalleryAccessRes]{}, miso.NewErr("Operation not allowed")
	}

	qpp := miso.QueryPageParam[ListedGalleryAccessRes]{
		ReqPage: req.PagingVo,
		AddSelectQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "gallery_no", "user_no", "create_time")
		},
		GetBaseQuery: func(tx *gorm.DB) *gorm.DB {
			return tx.Table("gallery_user_access").Order("id DESC")
		},
		ApplyConditions: func(tx *gorm.DB) *gorm.DB {
			return tx.Where("gallery_no = ?", req.GalleryNo).Where("is_del = 0")
		},
	}
	res, err := qpp.ExecPageQuery(rail, tx)
	if err != nil {
		return res, err
	}

	if len(res.Payload) > 0 {
		for i, p := range res.Payload {
			var toUser vault.UserInfo
			var err error
			if toUser, err = vault.FindUser(rail, vault.FindUserReq{
				UserNo: &p.UserNo,
			}); err != nil {
				return res, err
			}
			p.Username = toUser.Username
			res.Payload[i] = p
		}
	}

	return res, nil
}

func RemoveGalleryAccess(rail miso.Rail, tx *gorm.DB, cmd RemoveGalleryAccessCmd, user common.User) error {
	gallery, e := FindGallery(rail, tx, cmd.GalleryNo)
	if e != nil {
		return e
	}
	if gallery.UserNo != user.UserNo {
		return miso.NewErr("Operation not allowed")
	}

	e = tx.Exec(`UPDATE gallery_user_access SET is_del = 1, update_by = ? WHERE gallery_no = ? AND user_no = ?`,
		user.Username, cmd.GalleryNo, user.UserNo).Error
	if e != nil {
		return fmt.Errorf("failed to update gallery_user_access, galleryNo: %v, userNo: %v, %v", cmd.GalleryNo, cmd.UserNo, e)
	}
	rail.Infof("Gallery %v user access to %v is removed by %v", cmd.GalleryNo, cmd.UserNo, user.Username)
	return nil
}

// Grant user's access to the gallery, only the owner can do so
func GrantGalleryAccessToUser(rail miso.Rail, tx *gorm.DB, cmd PermitGalleryAccessCmd, user common.User) error {
	gallery, e := FindGallery(rail, tx, cmd.GalleryNo)
	if e != nil {
		return e
	}

	var toUser vault.UserInfo
	var err error
	if toUser, err = vault.FindUser(rail, vault.FindUserReq{
		Username: &cmd.Username,
	}); err != nil {
		return miso.NewErr("Failed to find user", "failed to find user, username: %v, %v", cmd.Username, err)
	}
	if toUser.Id < 1 {
		return miso.NewErr("User not found")
	}

	if gallery.UserNo != user.UserNo {
		return miso.NewErr("You are not allowed to grant access to this gallery")
	}

	return CreateGalleryAccess(rail, tx, toUser.UserNo, cmd.GalleryNo, user.Username)
}
