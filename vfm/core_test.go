package vfm

import (
	"bytes"
	"os"
	"testing"

	"github.com/curtisnewbie/gocommon/client"
	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/redis"
	"github.com/sirupsen/logrus"
)

func testUser() common.User {
	return common.User{
		UserId:   1,
		UserNo:   "UE202205142310076187414",
		Username: "zhuangyongj",
	}
}

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
	c := common.EmptyRail()
	var folderNo string = "hfKh3QZSsWjKufZWflqu8jb0n"
	r, e := listFilesInVFolder(c, mysql.GetConn(), ListFileReq{FolderNo: &folderNo}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesSelective(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	r, e := listFilesSelective(c, mysql.GetConn(), ListFileReq{}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesSelective(c, mysql.GetConn(), ListFileReq{
		Filename: &filename,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesForTags(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	var tagName string = "test"

	r, e := listFilesForTags(c, mysql.GetConn(), ListFileReq{TagName: &tagName}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesForTags(c, mysql.GetConn(), ListFileReq{
		Filename: &filename,
		TagName:  &tagName,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFileExists(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	fname := "test-files.zip"
	b, e := FileExists(c, mysql.GetConn(), PreflightCheckReq{Filename: fname}, testUser())
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
	c := common.EmptyRail()
	r, e := ListFileTags(c, ListFileTagReq{FileId: 1892}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFindParentFile(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	pf, e := FindParentFile(c, FetchParentFileReq{FileKey: "ZZZ687250496077824971813"}, testUser())
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
	c := common.EmptyRail()
	req := MoveIntoDirReq{
		Uuid:           "ZZZ687238965264384971813",
		ParentFileUuid: "ZZZ687238965264384925123",
	}
	e := MoveFileToDir(c, mysql.GetConn(), req, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestMakeDir(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	fileKey, e := MakeDir(c, MakeDirReq{Name: "mydir"}, testUser())
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
	c := common.EmptyRail()
	r, _ := common.ERand(5)
	folderNo, e := CreateVFolder(c, CreateVFolderReq{"MyFolder_" + r}, testUser())
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
	c := common.EmptyRail()
	dirs, e := ListDirs(c, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", dirs)
}

func TestGranteFileAccess(t *testing.T) {
	preTest(t)
	e := GranteFileAccess(common.EmptyRail(), 2, 3, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveGrantedFileAccess(t *testing.T) {
	preTest(t)
	e := RemoveGrantedFileAccess(common.EmptyRail(), RemoveGrantedAccessReq{FileId: 3, UserId: 2}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListGrantedFileAccess(t *testing.T) {
	preTest(t)
	l, e := ListGrantedFileAccess(common.EmptyRail(), ListGrantedAccessReq{FileId: 3})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestShareVFolder(t *testing.T) {
	preTest(t)
	if e := ShareVFolder(common.EmptyRail(),
		UserInfo{Id: 30, Username: "sharon", UserNo: "UE202205142310074386952"}, "VFLD20221001211317631020565809652", testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestRemoveVFolderAccess(t *testing.T) {
	preTest(t)
	req := RemoveGrantedFolderAccessReq{
		UserNo:   "UE202303190019399941339",
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
	}
	if e := RemoveVFolderAccess(common.EmptyRail(), req, testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestListVFolderBrief(t *testing.T) {
	preTest(t)
	v, e := ListVFolderBrief(common.EmptyRail(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", v)
}

func TestAddFileToVFolder(t *testing.T) {
	preTest(t)
	e := AddFileToVFolder(common.EmptyRail(),
		AddFileToVfolderReq{
			FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
			FileKeys: []string{"ZZZ687250481528832971813"},
		}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveFileFromVFolder(t *testing.T) {
	preTest(t)
	e := RemoveFileFromVFolder(common.EmptyRail(),
		RemoveFileFromVfolderReq{
			FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
			FileKeys: []string{"ZZZ687250481528832971813"},
		}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListVFolders(t *testing.T) {
	preTest(t)
	l, e := ListVFolders(common.EmptyRail(), ListVFolderReq{}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestListGrantedFolderAccess(t *testing.T) {
	preTest(t)
	l, e := ListGrantedFolderAccess(common.EmptyRail(),
		ListGrantedFolderAccessReq{FolderNo: "VFLD20221001211317631020565809652"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestUpdateFile(t *testing.T) {
	preTest(t)
	e := UpdateFile(common.EmptyRail(), UpdateFileReq{Id: 301, UserGroup: USER_GROUP_PRIVATE, Name: "test-files-222.zip"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListAllTags(t *testing.T) {
	preTest(t)
	tags, e := ListAllTags(common.EmptyRail(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", tags)
}

func TestTagFile(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()

	e := TagFile(c, TagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestUntagFile(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()

	e := UntagFile(c, UntagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestCreateFile(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()

	file, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(file)

	r := client.NewDynTClient(c, "/file", "fstore").
		AddHeader("filename", "README.md").
		Put(buf)
	if r.Err != nil {
		t.Fatal(r.Err)
	}

	resp, err := client.ReadGnResp[string](r)
	if err != nil {
		t.Fatal(err)
	}
	fakeFileId := resp.Data
	c.Infof("fake fileId: %v", fakeFileId)

	e := CreateFile(c, CreateFileReq{
		Filename:         "myfile",
		FakeFstoreFileId: fakeFileId,
		UserGroup:        USER_GROUP_PRIVATE,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestDeleteFile(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	e := DeleteFile(c, DeleteFileReq{Uuid: "ZZZ718078073798656022858"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestGenTempToken(t *testing.T) {
	preTest(t)
	c := common.EmptyRail()
	tkn, e := GenTempToken(c, GenerateTempTokenReq{"ZZZ687250496077824971813"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	if tkn == "" {
		t.Fatal("Token is empty")
	}
	t.Logf("tkn: %v", tkn)
}

func TestIsImage(t *testing.T) {
	n := "abc.jpg"
	if !isImage(n) {
		t.Fatal(n)
	}

	n = "abc.txt"
	if isImage(n) {
		t.Fatal(n)
	}
}
