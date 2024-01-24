package vfm

import "github.com/curtisnewbie/miso/miso"

var (
	vfmPool = miso.NewAsyncPool(5000, 100)
)
