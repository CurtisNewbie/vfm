package vfm

import (
	"testing"

	"github.com/curtisnewbie/miso/miso"
	"github.com/sirupsen/logrus"
)

func preClientTest(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	_, e := miso.GetConsulClient()
	miso.TestIsNil(t, e)
	miso.PollServiceListInstances()
}

func TestFindUserId(t *testing.T) {
	preClientTest(t)

	rail := miso.EmptyRail()
	id, e := FindUserId(rail, "zhuangyongj")
	miso.TestIsNil(t, e)
	miso.TestEqual(t, id, 1)
}

func TestFindUser(t *testing.T) {
	preClientTest(t)

	c := miso.EmptyRail()
	var uname string = "zhuangyongj"
	u, e := FindUser(c, FindUserReq{Username: &uname})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("user: %+v", u)
}

func TestFetchUsernames(t *testing.T) {
	preClientTest(t)

	c := miso.EmptyRail()
	res, e := FetchUsernames(c, FetchUsernamesReq{UserNos: []string{"UE202205142310076187414"}})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("res: %+v", res)
}

func TestFetchFstoreFileInfo(t *testing.T) {
	f, e := FetchFstoreFileInfo(miso.EmptyRail(), "file_688404712292352087399", "")
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("f: %+v", f)
}

func TestDeleteFstoreFile(t *testing.T) {
	e := DeleteFstoreFile(miso.EmptyRail(), "file_688400412377088926527")
	if e != nil {
		t.Fatal(e)
	}
}

func TestGetFstoreTmpToken(t *testing.T) {
	tkn, e := GetFstoreTmpToken(miso.EmptyRail(), "file_688399963701248926527", "tempfile")
	if e != nil {
		t.Fatal(e)
	}
	if tkn == "" {
		t.Fatal("temp token is empty")
	}
	t.Logf("tkn: %v", tkn)
}
