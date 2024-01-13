package vfm

import (
	"testing"

	"github.com/curtisnewbie/miso/miso"
)

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
