package vfm

import (
	"bytes"
	"os"
	"testing"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/client"
	"github.com/curtisnewbie/miso/core"
	"github.com/curtisnewbie/miso/mysql"
	"github.com/curtisnewbie/miso/redis"
	"github.com/sirupsen/logrus"
)

func testUser() common.User {
	return common.User{
		UserId:   1,
		UserNo:   "UE202205142310076187414",
		Username: "zhuangyongj",
	}
}

func corePreTest(t *testing.T) {
	user := "root"
	pw := ""
	db := "fileserver"
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
	corePreTest(t)
	c := core.EmptyRail()
	var folderNo string = "hfKh3QZSsWjKufZWflqu8jb0n"
	r, e := listFilesInVFolder(c, mysql.GetConn(), ListFileReq{FolderNo: &folderNo}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesSelective(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
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
	corePreTest(t)
	c := core.EmptyRail()
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
	corePreTest(t)
	c := core.EmptyRail()
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
	corePreTest(t)
	c := core.EmptyRail()
	r, e := ListFileTags(c, mysql.GetConn(), ListFileTagReq{FileId: 1892}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFindParentFile(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	pf, e := FindParentFile(c, mysql.GetConn(), FetchParentFileReq{FileKey: "ZZZ718071967023104410314"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	if pf.FileKey != "ZZZ718222444658688014704" {
		t.Fatalf("Incorrent ParentFileInfo, fileKey: %v, pf: %+v", pf.FileKey, pf)
	}
	t.Logf("%+v", pf)
}

func TestMoveFileToDir(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	req := MoveIntoDirReq{
		Uuid:           "ZZZ687238965264384971813",
		ParentFileUuid: "ZZZ718222444658688014704",
	}
	e := MoveFileToDir(c, mysql.GetConn(), req, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestMakeDir(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	fileKey, e := MakeDir(c, mysql.GetConn(), MakeDirReq{Name: "mydir"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	if fileKey == "" {
		t.Fatal("fileKey is empty")
	}
	t.Logf("fileKey: %v", fileKey)
}

func TestCreateVFolder(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	r, _ := core.ERand(5)
	folderNo, e := CreateVFolder(c, mysql.GetConn(), CreateVFolderReq{"MyFolder_" + r}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	if folderNo == "" {
		t.Fatal("folderNo is empty")
	}

	t.Logf("FolderNo: %v", folderNo)
}

func TestListDirs(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	dirs, e := ListDirs(c, mysql.GetConn(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", dirs)
}

func TestGranteFileAccess(t *testing.T) {
	corePreTest(t)
	e := GranteFileAccess(core.EmptyRail(), mysql.GetConn(), 2, 3, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveGrantedFileAccess(t *testing.T) {
	corePreTest(t)
	e := RemoveGrantedFileAccess(core.EmptyRail(), mysql.GetConn(), RemoveGrantedAccessReq{FileId: 3, UserId: 2}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListGrantedFileAccess(t *testing.T) {
	corePreTest(t)
	l, e := ListGrantedFileAccess(core.EmptyRail(), mysql.GetConn(), ListGrantedAccessReq{FileId: 3})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestShareVFolder(t *testing.T) {
	corePreTest(t)
	if e := ShareVFolder(core.EmptyRail(), mysql.GetConn(),
		UserInfo{Id: 30, Username: "sharon", UserNo: "UE202205142310074386952"}, "hfKh3QZSsWjKufZWflqu8jb0n", testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestRemoveVFolderAccess(t *testing.T) {
	corePreTest(t)
	req := RemoveGrantedFolderAccessReq{
		UserNo:   "UE202303190019399941339",
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
	}
	if e := RemoveVFolderAccess(core.EmptyRail(), mysql.GetConn(), req, testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestListVFolderBrief(t *testing.T) {
	corePreTest(t)
	v, e := ListVFolderBrief(core.EmptyRail(), mysql.GetConn(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", v)
}

func TestAddFileToVFolder(t *testing.T) {
	corePreTest(t)
	e := AddFileToVFolder(core.EmptyRail(), mysql.GetConn(),
		AddFileToVfolderReq{
			FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
			FileKeys: []string{"ZZZ687250481528832971813"},
		}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveFileFromVFolder(t *testing.T) {
	corePreTest(t)
	e := RemoveFileFromVFolder(core.EmptyRail(), mysql.GetConn(),
		RemoveFileFromVfolderReq{
			FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
			FileKeys: []string{"ZZZ687250481528832971813"},
		}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListVFolders(t *testing.T) {
	corePreTest(t)
	l, e := ListVFolders(core.EmptyRail(), mysql.GetConn(), ListVFolderReq{}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestListGrantedFolderAccess(t *testing.T) {
	corePreTest(t)
	l, e := ListGrantedFolderAccess(core.EmptyRail(), mysql.GetConn(),
		ListGrantedFolderAccessReq{FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestUpdateFile(t *testing.T) {
	corePreTest(t)
	e := UpdateFile(core.EmptyRail(), mysql.GetConn(), UpdateFileReq{Id: 301, UserGroup: USER_GROUP_PRIVATE, Name: "test-files-222.zip"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListAllTags(t *testing.T) {
	corePreTest(t)
	tags, e := ListAllTags(core.EmptyRail(), mysql.GetConn(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", tags)
}

func TestTagFile(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()

	e := TagFile(c, mysql.GetConn(), TagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestUntagFile(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()

	e := UntagFile(c, mysql.GetConn(), UntagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestCreateFile(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()

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

	e := CreateFile(c, mysql.GetConn(), CreateFileReq{
		Filename:         "myfile",
		FakeFstoreFileId: fakeFileId,
		UserGroup:        USER_GROUP_PRIVATE,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestDeleteFile(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	e := DeleteFile(c, mysql.GetConn(), DeleteFileReq{Uuid: "ZZZ718078073798656022858"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestGenTempToken(t *testing.T) {
	corePreTest(t)
	c := core.EmptyRail()
	tkn, e := GenTempToken(c, mysql.GetConn(), GenerateTempTokenReq{"ZZZ687250496077824971813"}, testUser())
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
