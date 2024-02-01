package vfm

import "github.com/curtisnewbie/miso/miso"

var (
	vfmPool = miso.NewAsyncPool(500, 20)
)
