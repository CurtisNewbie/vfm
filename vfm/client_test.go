package vfm

import (
	"testing"

	"github.com/curtisnewbie/gocommon/common"
)

func TestFindUserId(t *testing.T) {
	c := common.EmptyExecContext()
	id, e := FindUserId(c, "zhuangyongj")
	if e != nil {
		t.Fatal(e)
	}
	if id != 1 {
		t.Fatal("id != 1")
	}
}

func TestFindUser(t *testing.T) {
	c := common.EmptyExecContext()
	var uname string = "zhuangyongj"
	u, e := FindUser(c, FindUserReq{Username: &uname})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("user: %+v", u)
}

func TestFetchUsernames(t *testing.T) {
	c := common.EmptyExecContext()
	res, e := FetchUsernames(c, FetchUsernamesReq{UserNos: []string{"UE202205142310074386952"}})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("res: %+v", res)
}
