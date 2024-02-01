package vfm

import (
	"fmt"
	"time"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
	"gorm.io/gorm"
)

// GalleryImage.status (doesn't really matter anymore)
type ImgStatus string

const (
	NORMAL  ImgStatus = "NORMAL"
	DELETED ImgStatus = "DELETED"

	// 40mb is the maximum size for an image
	IMAGE_SIZE_THRESHOLD int64 = 40 * 1048576
)

type TransferGalleryImageReq struct {
	Images []CreateGalleryImageCmd
}

type TransferGalleryImageInDirReq struct {
	// gallery no
	GalleryNo string `json:"galleryNo" validation:"notEmpty"`

	// file key of the directory
	FileKey string `json:"fileKey" validation:"notEmpty"`
}

// Image that belongs to a Gallery
type GalleryImage struct {
	ID         int64
	GalleryNo  string
	ImageNo    string
	Name       string
	FileKey    string
	Status     ImgStatus
	CreateTime time.Time
	CreateBy   string
	UpdateTime time.Time
	UpdateBy   string
	IsDel      common.IS_DEL
}

func (GalleryImage) TableName() string {
	return "gallery_image"
}

type ThumbnailInfo struct {
	Name string
	Path string
}

type CreateGalleryImgEvent struct {
	Username     string `json:"username"`
	UserNo       string `json:"userNo"`
	DirFileKey   string `json:"dirFileKey"`
	DirName      string `json:"dirName"`
	ImageName    string `json:"imageName"`
	ImageFileKey string `json:"imageFileKey"`
}

type ListGalleryImagesCmd struct {
	GalleryNo   string `json:"galleryNo" validation:"notEmpty"`
	miso.Paging `json:"pagingVo"`
}

type ListGalleryImagesResp struct {
	Images []ImageInfo `json:"images"`
	Paging miso.Paging `json:"pagingVo"`
}

type ImageInfo struct {
	ThumbnailToken  string `json:"thumbnailToken"`
	FileTempToken   string `json:"fileTempToken"`
	ImageFileId     string `json:"-"`
	ThumbnailFileId string `json:"-"`
}

type CreateGalleryImageCmd struct {
	GalleryNo string `json:"galleryNo"`
	Name      string `json:"name"`
	FileKey   string `json:"fileKey"`
}

func DeleteGalleryImage(rail miso.Rail, tx *gorm.DB, fileKey string) error {
	return tx.Exec("delete from gallery_image where file_key = ?", fileKey).Error
}

// Create a gallery image record
func CreateGalleryImage(rail miso.Rail, cmd CreateGalleryImageCmd, userNo string, username string, tx *gorm.DB) error {
	creator, err := FindGalleryCreator(rail, cmd.GalleryNo, tx)
	if err != nil {
		return err
	}

	if *creator != userNo {
		return miso.NewErr("You are not allowed to upload image to this gallery")
	}

	lock := NewGalleryFileLock(rail, cmd.GalleryNo, cmd.FileKey)
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to obtain gallery image lock, gallery:%v, fileKey: %v", cmd.GalleryNo, cmd.FileKey)
	}
	defer lock.Unlock()

	if isCreated, e := isImgCreatedAlready(rail, tx, cmd.GalleryNo, cmd.FileKey); isCreated || e != nil {
		if e != nil {
			return e
		}
		rail.Infof("Image '%s' added already", cmd.Name)
		return nil
	}

	imageNo := miso.GenNoL("IMG", 25)
	return tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`insert into gallery_image (gallery_no, image_no, name, file_key, create_by) values (?, ?, ?, ?, ?)`,
			cmd.GalleryNo, imageNo, cmd.Name, cmd.FileKey, username).Error; err != nil {
			return err
		}
		return tx.Exec(`update gallery set update_time = ? where gallery_no = ?`, miso.Now(), cmd.GalleryNo).Error
	})
}

type FstoreTmpToken struct {
	FileId  string
	TempKey string
}

func GenFstoreTknAsync(rail miso.Rail, fileId string, name string) miso.Future[FstoreTmpToken] {
	return miso.SubmitAsync[FstoreTmpToken](fstoreTokenPool,
		func() (FstoreTmpToken, error) {
			tkn, err := GetFstoreTmpToken(rail.NextSpan(), fileId, name)
			if err != nil {
				return FstoreTmpToken{FileId: fileId}, err
			}
			return FstoreTmpToken{
				FileId:  fileId,
				TempKey: tkn,
			}, nil
		})
}

// List gallery images
func ListGalleryImages(rail miso.Rail, tx *gorm.DB, cmd ListGalleryImagesCmd, user common.User) (*ListGalleryImagesResp, error) {
	rail.Infof("ListGalleryImages, cmd: %+v", cmd)

	if hasAccess, err := HasAccessToGallery(rail, tx, user.UserNo, cmd.GalleryNo); err != nil || !hasAccess {
		if err != nil {
			return nil, fmt.Errorf("check HasAccessToGallery failed, %v", err)
		}
		return nil, miso.NewErr("You are not allowed to access this gallery")
	}

	const selectSql string = `
		select image_no, file_key from gallery_image
		where gallery_no = ?
		order by id desc
		limit ?, ?
	`
	var galleryImages []GalleryImage
	t := tx.Raw(selectSql, cmd.GalleryNo, cmd.Paging.GetOffset(), cmd.Paging.GetLimit()).Scan(&galleryImages)
	if t.Error != nil {
		return nil, fmt.Errorf("select gallery_image failed, %v", t.Error)
	}

	if galleryImages == nil {
		galleryImages = []GalleryImage{}
	}

	// count total asynchronoulsy (normally, when the SELECT is successful, the COUNT doesn't really fail)
	countFuture := miso.RunAsync(func() (int, error) {
		var total int
		t := tx.Raw(`select count(*) from gallery_image where gallery_no = ?`, cmd.GalleryNo).
			Scan(&total)
		if t.Error == nil {
			return total, nil
		}
		return total, fmt.Errorf("failed to count gallery_image, %v", t.Error)
	})

	// generate temp tokens for the actual files and the thumbnail, these are served by mini-fstore
	images := []ImageInfo{}
	if len(galleryImages) > 0 {

		genTknFutures := []miso.Future[FstoreTmpToken]{}

		for _, img := range galleryImages {
			r, e := findFile(rail, tx, img.FileKey)
			if e != nil {
				rail.Errorf("findFile failed, fileKey: %v, %v", img.FileKey, e)
				continue
			}
			fstoreFileId := r.FstoreFileId
			genTknFutures = append(genTknFutures, GenFstoreTknAsync(rail, fstoreFileId, r.Name))

			thumbnailFileId := r.Thumbnail
			if thumbnailFileId == "" {
				thumbnailFileId = fstoreFileId
			} else {
				genTknFutures = append(genTknFutures, GenFstoreTknAsync(rail, thumbnailFileId, r.Name))
			}

			images = append(images, ImageInfo{ImageFileId: fstoreFileId, ThumbnailFileId: thumbnailFileId})
		}

		tokens := []FstoreTmpToken{}

		for i := range genTknFutures {
			res, err := genTknFutures[i].Get()
			if err != nil {
				rail.Errorf("Failed to get mini-fstore temp token for fstore_file_id: %v, %v", res.FileId, err)
				continue
			}
			tokens = append(tokens, res)
		}

		idTknMap := map[string]string{}
		for _, t := range tokens {
			idTknMap[t.FileId] = t.TempKey
		}
		for i, im := range images {
			im.ThumbnailToken = idTknMap[im.ThumbnailFileId]
			im.FileTempToken = idTknMap[im.ImageFileId]
			images[i] = im
		}
	}

	total, errCnt := countFuture.Get()
	if errCnt != nil {
		return nil, errCnt
	}

	return &ListGalleryImagesResp{Images: images, Paging: miso.RespPage(cmd.Paging, total)}, nil
}

func BatchTransferAsync(rail miso.Rail, cmd TransferGalleryImageReq, user common.User, tx *gorm.DB) (any, error) {
	if cmd.Images == nil || len(cmd.Images) < 1 {
		return nil, nil
	}

	// validate the keys first
	for _, img := range cmd.Images {
		if isValid, e := ValidateFileOwner(rail, tx, ValidateFileOwnerReq{
			FileKey: img.FileKey,
			UserId:  user.UserId,
		}); e != nil || !isValid {
			if e != nil {
				return nil, e
			}
			return nil, miso.NewErr(fmt.Sprintf("Only file's owner can make it a gallery image ('%s')", img.Name))
		}
	}

	// start transferring
	go func(rail miso.Rail, images []CreateGalleryImageCmd) {
		for _, cmd := range images {
			fi, er := findFile(rail, tx, cmd.FileKey)
			if er != nil {
				rail.Errorf("Failed to fetch file info while transferring selected images, fi's fileKey: %s, error: %v", cmd.FileKey, er)
				continue
			}

			if fi.FileType == FileTypeFile { // a file
				if fi.FstoreFileId == "" {
					continue // doesn't have fstore fileId, cannot be transferred
				}

				if GuessIsImage(rail, fi) {
					nc := CreateGalleryImageCmd{GalleryNo: cmd.GalleryNo, Name: fi.Name, FileKey: fi.Uuid}
					if err := CreateGalleryImage(rail, nc, user.UserNo, user.Username, tx); err != nil {
						rail.Errorf("Failed to create gallery image, fi's fileKey: %s, error: %v", cmd.FileKey, err)
						continue
					}
				}
			} else { // a directory
				treq := TransferGalleryImageInDirReq{
					GalleryNo: cmd.GalleryNo,
					FileKey:   cmd.FileKey,
				}
				if err := TransferImagesInDir(rail, treq, user, tx); err != nil {
					rail.Errorf("Failed to transfer images in directory, fi's fileKey: %s, error: %v", cmd.FileKey, err)
					continue
				}
			}
		}
	}(rail, cmd.Images)

	return nil, nil
}

// Transfer images in dir
func TransferImagesInDir(rail miso.Rail, cmd TransferGalleryImageInDirReq, user common.User, tx *gorm.DB) error {
	fi, e := findFile(rail, tx, cmd.FileKey)
	if e != nil {
		return e
	}

	// only the owner of the directory can do this, by default directory is only visible to the uploader
	if fi.UploaderId != user.UserId {
		return miso.NewErr("Not permitted operation")
	}

	if fi.FileType != FileTypeDir {
		return miso.NewErr("This is not a directory")
	}

	if fi.IsLogicDeleted == LDelY || fi.IsPhysicDeleted == PDelY {
		return miso.NewErr("Directory is already deleted")
	}
	dirFileKey := cmd.FileKey
	galleryNo := cmd.GalleryNo
	start := time.Now()

	page := 1
	for {
		// dirFileKey, 100, page
		res, err := ListFilesInDir(rail, tx, ListFilesInDirReq{
			FileKey: dirFileKey,
			Limit:   100,
			Page:    page,
		})
		if err != nil {
			rail.Errorf("Failed to list files in dir, dir's fileKey: %s, error: %v", dirFileKey, err)
			break
		}
		if res == nil || len(res) < 1 {
			break
		}

		// starts fetching file one by one
		for i := 0; i < len(res); i++ {
			fk := res[i]
			fi, er := findFile(rail, tx, fk)
			if er != nil {
				rail.Errorf("Failed to fetch file info while looping files in dir, fi's fileKey: %s, error: %v", fk, er)
				continue
			}

			if GuessIsImage(rail, fi) {
				cmd := CreateGalleryImageCmd{GalleryNo: galleryNo, Name: fi.Name, FileKey: fi.Uuid}
				if err := CreateGalleryImage(rail, cmd, user.UserNo, user.Username, tx); err != nil {
					rail.Errorf("Failed to create gallery image, fi's fileKey: %s, error: %v", fk, err)
				}
			}
		}

		page += 1
	}

	rail.Infof("Finished TransferImagesInDir, dir's fileKey: %s, took: %s", dirFileKey, time.Since(start))
	return nil
}

// Guess whether a file is an image
func GuessIsImage(rail miso.Rail, f FileInfo) bool {
	if f.SizeInBytes > IMAGE_SIZE_THRESHOLD {
		return false
	}
	if f.FileType != FileTypeFile {
		return false
	}
	if f.Thumbnail == "" {
		rail.Infof("File doesn't have thumbnail, fileKey: %v", f.Uuid)
		return false
	}

	return true
}

// check whether the gallery image is created already
//
// return isImgCreated, error
func isImgCreatedAlready(rail miso.Rail, tx *gorm.DB, galleryNo string, fileKey string) (bool, error) {
	var id int
	tx = tx.Raw(`
		SELECT id FROM gallery_image
		WHERE gallery_no = ?
		AND file_key = ?
		AND is_del = 0
		`, galleryNo, fileKey).Scan(&id)

	if e := tx.Error; e != nil || tx.RowsAffected < 1 {
		return false, tx.Error
	}

	return true, nil
}

func NewGalleryFileLock(rail miso.Rail, galleryNo string, fileKey string) *miso.RLock {
	return miso.NewRLockf(rail, "gallery:image:%v:%v", galleryNo, fileKey)
}
