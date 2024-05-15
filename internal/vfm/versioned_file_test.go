package vfm

import (
	"testing"

	"github.com/curtisnewbie/miso/miso"
)

func TestCreateVerFile(t *testing.T) {
	corePreTest(t)
	rail := miso.EmptyRail()
	testFileKey := "ZZZ687238965264384971813"

	files, err := CreateVerFile(rail, miso.GetMySQL(), ApiCreateVerFileReq{FileKey: testFileKey}, testUser())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", files)
}

func TestListVerFile(t *testing.T) {
	corePreTest(t)
	rail := miso.EmptyRail()
	files, err := ListVerFile(rail, miso.GetMySQL(), ApiListVerFileReq{}, testUser())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", files)
}

func TestUpdateVerFile(t *testing.T) {
	corePreTest(t)
	rail := miso.EmptyRail()
	verFileId := "verf_1224870865715200431057"
	testFileKey := "ZZZ687250496077824971813"

	err := UpdateVerFile(rail, miso.GetMySQL(),
		ApiUpdateVerFileReq{VerFileId: verFileId, FileKey: testFileKey},
		testUser())

	if err != nil {
		t.Fatal(err)
	}
}

func TestDelVerFile(t *testing.T) {
	corePreTest(t)
	rail := miso.EmptyRail()
	verFileId := "verf_1224911895724032158144"

	err := DelVerFile(rail, miso.GetMySQL(),
		ApiDelVerFileReq{VerFileId: verFileId},
		testUser())

	if err != nil {
		t.Fatal(err)
	}
}
