package vfm

import (
	"testing"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/mysql"
	"github.com/curtisnewbie/gocommon/server"
)

func preTest(t *testing.T) {
	user := "root"
	pw := "123456"
	db := "fileServer"
	host := "localhost"
	port := "3306"
	connParam := "charset=utf8mb4&parseTime=True&loc=Local&readTimeout=30s&writeTimeout=30s&timeout=3s"

	server.ConfigureLogging()
	if e := mysql.InitMySql(user, pw, db, host, port, connParam); e != nil {
		t.Fatal(e)
	}
}

func TestListFilesInVFolder(t *testing.T) {
	preTest(t)
	c := common.EmptyExecContext()
	c.User.UserNo = "GyaYqTKsyGIxmAFaHgNYztA0y"
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
	c.User.UserNo = "GyaYqTKsyGIxmAFaHgNYztA0y"
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
	c.User.UserNo = "GyaYqTKsyGIxmAFaHgNYztA0y"
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
