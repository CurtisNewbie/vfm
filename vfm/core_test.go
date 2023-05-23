package vfm

import (
	"testing"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/redis"
	"github.com/sirupsen/logrus"
)

func preTest(t *testing.T) {
	user := "root"
	// pw := "123456"
	pw := ""
	db := "fileServer"
	host := "localhost"
	port := "3306"
	connParam := "charset=utf8mb4&parseTime=True&loc=Local&readTimeout=30s&writeTimeout=30s&timeout=3s"

	if e := mysql.InitMySql(user, pw, db, host, port, connParam); e != nil {
		t.Fatal(e)
	}
	if _, e := redis.InitRedisFromProp(); e != nil {
		t.Fatal(e)
	}
	logrus.SetLevel(logrus.DebugLevel)
}

func TestListFilesInVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserNo = "UE202205142310076187414"
	c.User.UserId = "1"
	var folderNo string = "hfKh3QZSsWjKufZWflqu8jb0n"
	r, e := listFilesInVFolder(c, ListFileReq{FolderNo: &folderNo})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesSelective(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserNo = "UE202205142310076187414"
	c.User.UserId = "1"

	r, e := listFilesSelective(c, ListFileReq{})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesSelective(c, ListFileReq{
		Filename: &filename,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesForTags(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserNo = "UE202205142310076187414"
	c.User.UserId = "1"
	var tagName string = "test"

	r, e := listFilesForTags(c, ListFileReq{
		TagName: &tagName,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesForTags(c, ListFileReq{
		Filename: &filename,
		TagName:  &tagName,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFileExists(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"

	fname := "test-files.zip"
	b, e := FileExists(c, fname, "")
	if e != nil {
		t.Fatal(e)
	}
	exist, ok := b.(bool)
	if !ok {
		t.Fatal("returned value is not of bool type")
	}
	t.Logf("%s exists? %v", fname, exist)
}

func TestListFileTags(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "3"

	r, e := ListFileTags(c, ListFileTagReq{FileId: 1892})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFindParentFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"

	pf, e := FindParentFile(c, "ZZZ687250496077824971813")
	if e != nil {
		t.Fatal(e)
	}
	if pf.FileKey != "ZZZ687238965264384925123" {
		t.Fatalf("Incorrent ParentFileInfo, fileKey: %v", pf.FileKey)
	}
	t.Logf("%+v", pf)
}

func TestMoveFileToDir(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"

	e := MoveFileToDir(c, MoveIntoDirReq{
		Uuid:           "ZZZ687238965264384971813",
		ParentFileUuid: "ZZZ687238965264384925123",
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestMakeDir(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	fileKey, e := MakeDir(c, MakeDirReq{Name: "mydir"})
	if e != nil {
		t.Fatal(e)
	}
	if fileKey == "" {
		t.Fatal("fileKey is empty")
	}
	t.Logf("fileKey: %v", fileKey)
}

func TestCreateVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	r, _ := common.ERand(5)
	folderNo, e := CreateVFolder(c, CreateVFolderReq{"MyFolder_" + r})
	if e != nil {
		t.Fatal(e)
	}
	if folderNo == "" {
		t.Fatal("folderNo is empty")
	}

	t.Logf("FolderNo: %v", folderNo)
}

func TestListDirs(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	dirs, e := ListDirs(c)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", dirs)
}

func TestGranteFileAccess(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	c.User.Username = "zhuangyongj"

	e := GranteFileAccess(c, 2, 3)
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveGrantedFileAccess(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	c.User.Username = "zhuangyongj"
	e := RemoveGrantedFileAccess(c, RemoveGrantedAccessReq{FileId: 3, UserId: 2})
	if e != nil {
		t.Fatal(e)
	}
}

func TestListGrantedFileAccess(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	c.User.Username = "zhuangyongj"
	l, e := ListGrantedFileAccess(c, ListGrantedAccessReq{FileId: 3})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestShareVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	c.User.Username = "zhuangyongj"

	if e := ShareVFolder(c, UserInfo{Id: 30, Username: "sharon", UserNo: "UE202205142310074386952"}, "VFLD20221001211317631020565809652"); e != nil {
		t.Fatal(e)
	}
}

func TestRemoveVFolderAccess(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	c.User.Username = "zhuangyongj"

	req := RemoveGrantedFolderAccessReq{
		UserNo:   "UE202303190019399941339",
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
	}
	if e := RemoveVFolderAccess(c, req); e != nil {
		t.Fatal(e)
	}
}

func TestListVFolderBrief(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	v, e := ListVFolderBrief(c)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", v)
}

func TestAddFileToVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := AddFileToVFolder(c, AddFileToVfolderReq{
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
		FileKeys: []string{"ZZZ687250481528832971813"},
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveFileFromVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := RemoveFileFromVFolder(c, RemoveFileFromVfolderReq{
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
		FileKeys: []string{"ZZZ687250481528832971813"},
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestListVFolders(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"
	l, e := ListVFolders(c, ListVFolderReq{})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestListGrantedFolderAccess(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	l, e := ListGrantedFolderAccess(c, ListGrantedFolderAccessReq{FolderNo: "VFLD20221001211317631020565809652"})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestUpdateFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := UpdateFile(c, UpdateFileReq{Id: 301, UserGroup: USER_GROUP_PRIVATE, Name: "test-files-222.zip"})
	if e != nil {
		t.Fatal(e)
	}
}

func TestListAllTags(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "3"
	tags, e := ListAllTags(c)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", tags)
}

func TestTagFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := TagFile(c, TagFileReq{FileId: 355, TagName: "mytag"})
	if e != nil {
		t.Fatal(e)
	}
}

func TestUntagFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := UntagFile(c, UntagFileReq{FileId: 355, TagName: "mytag"})
	if e != nil {
		t.Fatal(e)
	}
}

func TestCreateFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := CreateFile(c, CreateFileReq{
		Filename:     "myfile",
		FstoreFileId: "file_688404712292352087399",
		UserGroup:    USER_GROUP_PRIVATE,
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestDeleteFile(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserId = "1"
	c.User.UserNo = "UE202205142310076187414"

	e := DeleteFile(c, DeleteFileReq{Uuid: "ZZZ718078073798656022858"})
	if e != nil {
		t.Fatal(e)
	}
}
