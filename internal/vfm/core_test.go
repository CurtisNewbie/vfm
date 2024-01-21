package vfm

import (
	"bytes"
	"os"
	"testing"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/miso/miso"
	vault "github.com/curtisnewbie/user-vault/api"
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
	port := 3306
	rail := miso.EmptyRail()

	p := miso.MySQLConnParam{
		User:     user,
		Password: pw,
		Schema:   db,
		Host:     host,
		Port:     port,
	}

	if e := miso.InitMySQL(rail, p); e != nil {
		t.Fatal(e)
	}
	if _, e := miso.InitRedisFromProp(rail); e != nil {
		t.Fatal(e)
	}

	miso.SetProp(miso.PropRabbitMqUsername, "guest")
	miso.SetProp(miso.PropRabbitMqPassword, "guest")
	if e := miso.StartRabbitMqClient(rail); e != nil {
		t.Fatal(e)
	}
	logrus.SetLevel(logrus.DebugLevel)
}

func TestListFilesInVFolder(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	var folderNo string = "hfKh3QZSsWjKufZWflqu8jb0n"
	r, e := listFilesInVFolder(c, miso.GetMySQL(), ListFileReq{FolderNo: &folderNo}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesSelective(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	r, e := listFilesSelective(c, miso.GetMySQL(), ListFileReq{}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesSelective(c, miso.GetMySQL(), ListFileReq{
		Filename: &filename,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestListFilesForTags(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	var tagName string = "test"

	r, e := listFilesForTags(c, miso.GetMySQL(), ListFileReq{TagName: &tagName}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)

	var filename = "myfile"
	r, e = listFilesForTags(c, miso.GetMySQL(), ListFileReq{
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
	c := miso.EmptyRail()
	fname := "test-files.zip"
	b, e := FileExists(c, miso.GetMySQL(), PreflightCheckReq{Filename: fname}, testUser())
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
	c := miso.EmptyRail()
	r, e := ListFileTags(c, miso.GetMySQL(), ListFileTagReq{FileId: 1892}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", r)
}

func TestFindParentFile(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	pf, e := FindParentFile(c, miso.GetMySQL(), FetchParentFileReq{FileKey: "ZZZ718071967023104410314"}, testUser())
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
	c := miso.EmptyRail()
	req := MoveIntoDirReq{
		Uuid: "eb6bc04f-15c5-4f85-a84d-be3d5a7236d8",
		// ParentFileUuid: "5ddf49ca-dec9-4ecf-962d-47b0f3eab90c",
		ParentFileUuid: "",
	}
	e := MoveFileToDir(c, miso.GetMySQL(), req, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestMakeDir(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	fileKey, e := MakeDir(c, miso.GetMySQL(), MakeDirReq{Name: "mydir"}, testUser())
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
	c := miso.EmptyRail()
	r := miso.ERand(5)
	folderNo, e := CreateVFolder(c, miso.GetMySQL(), CreateVFolderReq{"MyFolder_" + r}, testUser())
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
	c := miso.EmptyRail()
	dirs, e := ListDirs(c, miso.GetMySQL(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", dirs)
}

func TestShareVFolder(t *testing.T) {
	corePreTest(t)
	if e := ShareVFolder(miso.EmptyRail(), miso.GetMySQL(),
		vault.UserInfo{Id: 30, Username: "sharon", UserNo: "UE202205142310074386952"}, "hfKh3QZSsWjKufZWflqu8jb0n", testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestRemoveVFolderAccess(t *testing.T) {
	corePreTest(t)
	req := RemoveGrantedFolderAccessReq{
		UserNo:   "UE202303190019399941339",
		FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
	}
	if e := RemoveVFolderAccess(miso.EmptyRail(), miso.GetMySQL(), req, testUser()); e != nil {
		t.Fatal(e)
	}
}

func TestListVFolderBrief(t *testing.T) {
	corePreTest(t)
	v, e := ListVFolderBrief(miso.EmptyRail(), miso.GetMySQL(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", v)
}

func TestAddFileToVFolder(t *testing.T) {
	corePreTest(t)
	e := AddFileToVFolder(miso.EmptyRail(), miso.GetMySQL(),
		AddFileToVfolderReq{
			FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n",
			FileKeys: []string{"ZZZ687250481528832971813"},
			Sync:     true,
		}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestRemoveFileFromVFolder(t *testing.T) {
	corePreTest(t)
	e := RemoveFileFromVFolder(miso.EmptyRail(), miso.GetMySQL(),
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
	l, e := ListVFolders(miso.EmptyRail(), miso.GetMySQL(), ListVFolderReq{}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestListGrantedFolderAccess(t *testing.T) {
	corePreTest(t)
	l, e := ListGrantedFolderAccess(miso.EmptyRail(), miso.GetMySQL(),
		ListGrantedFolderAccessReq{FolderNo: "hfKh3QZSsWjKufZWflqu8jb0n"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", l)
}

func TestUpdateFile(t *testing.T) {
	corePreTest(t)
	e := UpdateFile(miso.EmptyRail(), miso.GetMySQL(), UpdateFileReq{Id: 301, Name: "test-files-222.zip"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestListAllTags(t *testing.T) {
	corePreTest(t)
	tags, e := ListAllTags(miso.EmptyRail(), miso.GetMySQL(), testUser())
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", tags)
}

func TestTagFile(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()

	e := TagFile(c, miso.GetMySQL(), TagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestUntagFile(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()

	e := UntagFile(c, miso.GetMySQL(), UntagFileReq{FileId: 355, TagName: "mytag"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestCreateFile(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()

	file, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(file)

	var r miso.GnResp[string]
	err = miso.NewDynTClient(c, "/file", "fstore").
		AddHeader("filename", "README.md").
		Put(buf).
		Json(&r)
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	fakeFileId := r.Data
	c.Infof("fake fileId: %v", fakeFileId)

	e := CreateFile(c, miso.GetMySQL(), CreateFileReq{
		Filename:         "myfile",
		FakeFstoreFileId: fakeFileId,
	}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestDeleteFile(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	e := DeleteFile(c, miso.GetMySQL(), DeleteFileReq{Uuid: "ZZZ718078073798656022858"}, testUser())
	if e != nil {
		t.Fatal(e)
	}
}

func TestGenTempToken(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	tkn, e := GenTempToken(c, miso.GetMySQL(), GenerateTempTokenReq{"ZZZ687250496077824971813"}, testUser())
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

func TestBatchClacDirSize(t *testing.T) {
	corePreTest(t)
	c := miso.EmptyRail()
	err := BatchCalcDirSize(c, miso.GetMySQL())
	if err != nil {
		t.Fatal(err)
	}
}
