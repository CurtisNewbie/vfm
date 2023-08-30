package vfm

import (
	"testing"

	"github.com/curtisnewbie/miso/consul"
	"github.com/curtisnewbie/miso/core"
	"github.com/sirupsen/logrus"
)

func preClientTest(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	_, e := consul.GetConsulClient()
	core.TestIsNil(t, e)
	consul.PollServiceListInstances()
}

func TestFindUserId(t *testing.T) {
	preClientTest(t)

	rail := core.EmptyRail()
	id, e := FindUserId(rail, "zhuangyongj")
	core.TestIsNil(t, e)
	core.TestEqual(t, id, 1)
}

func TestFindUser(t *testing.T) {
	preClientTest(t)

	c := core.EmptyRail()
	var uname string = "zhuangyongj"
	u, e := FindUser(c, FindUserReq{Username: &uname})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("user: %+v", u)
}

func TestFetchUsernames(t *testing.T) {
	preClientTest(t)

	c := core.EmptyRail()
	res, e := FetchUsernames(c, FetchUsernamesReq{UserNos: []string{"UE202205142310076187414"}})
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
