package vfm

import (
	"testing"

	"github.com/curtisnewbie/gocommon/common"
	"github.com/curtisnewbie/gocommon/consul"
)

func TestFindUserId(t *testing.T) {
	if _, e := consul.GetConsulClient(); e != nil {
		t.Fatal(e)
	}
	c := common.EmptyExecContext()
	id, e := FindUserId(c, "zhuangyongj")
	if e != nil {
		t.Fatal(e)
	}
	if id != 1 {
		t.Fatal("id != 1")
	}
}
