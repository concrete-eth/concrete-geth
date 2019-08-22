// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 appengine gccgo

package blake2b

func hashBlocks(h *[8]uint64, c *[2]uint64, flag uint64, blocks []byte) {
	hashBlocksGeneric(h, c, flag, blocks)
}

func f(h *[8]uint64, m [16]uint64, c0, c1 uint64, flag uint64, rounds int) {
	fGeneric(h, m, c0, c1, flag, rounds)
}
