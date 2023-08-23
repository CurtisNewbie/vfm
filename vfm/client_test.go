package vfm

import (
	"testing"

	"github.com/curtisnewbie/miso/core"
)

func TestFindUserId(t *testing.T) {
	c := core.EmptyRail()
	id, e := FindUserId(c, "zhuangyongj")
	if e != nil {
		t.Fatal(e)
	}
	if id != 1 {
		t.Fatal("id != 1")
	}
}

func TestFindUser(t *testing.T) {
	c := core.EmptyRail()
	var uname string = "zhuangyongj"
	u, e := FindUser(c, FindUserReq{Username: &uname})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("user: %+v", u)
}

func TestFetchUsernames(t *testing.T) {
	c := core.EmptyRail()
	res, e := FetchUsernames(c, FetchUsernamesReq{UserNos: []string{"UE202205142310074386952"}})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("res: %+v", res)
}

func TestFetchFstoreFileInfo(t *testing.T) {
	f, e := FetchFstoreFileInfo(core.EmptyRail(), "file_688404712292352087399", "")
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("f: %+v", f)
}

func TestDeleteFstoreFile(t *testing.T) {
	e := DeleteFstoreFile(core.EmptyRail(), "file_688400412377088926527")
	if e != nil {
		t.Fatal(e)
	}
}

func TestGetFstoreTmpToken(t *testing.T) {
	tkn, e := GetFstoreTmpToken(core.EmptyRail(), "file_688399963701248926527", "tempfile")
	if e != nil {
		t.Fatal(e)
	}
	if tkn == "" {
		t.Fatal("temp token is empty")
	}
	t.Logf("tkn: %v", tkn)
}
