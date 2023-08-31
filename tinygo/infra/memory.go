// Copyright 2023 The concrete-geth Authors
//
// The concrete-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The concrete library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the concrete library. If not, see <http://www.gnu.org/licenses/>.

package infra

import (
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
)

var Memory memory.Memory = &mem{}
var Allocator memory.Allocator = &alloc{}

var allocs = make(map[uintptr][]byte)

//export concrete_Malloc
func Malloc(size uint64) uint64 {
	if size == 0 {
		return 0
	}
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	allocs[ptr] = buf
	return uint64(ptr)
}

//export concrete_Free
func Free(pointer uint64) {
	ptr := uintptr(pointer)
	if _, ok := allocs[ptr]; ok {
		delete(allocs, ptr)
	} else {
		panic("free: invalid pointer")
	}
}

//export concrete_Prune
func Prune() {
	allocs = make(map[uintptr][]byte)
}

type mem struct{}

func (m *mem) Write(data []byte) memory.MemPointer {
	if len(data) == 0 {
		return memory.NullPointer
	}
	offset := uint32(uintptr(unsafe.Pointer(&data[0])))
	size := uint32(len(data))
	var pointer memory.MemPointer
	pointer.Pack(offset, size)
	return pointer
}

func (m *mem) Read(pointer memory.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(offset),
		//nolint:typecheck
		Len: uintptr(size),
		//nolint:typecheck
		Cap: uintptr(size),
	}))
}

type alloc struct{}

func (m *alloc) Malloc(size uint32) memory.MemPointer {
	offset := Malloc(uint64(size))
	var pointer memory.MemPointer
	pointer.Pack(uint32(offset), size)
	return pointer
}

func (m *alloc) Free(pointer memory.MemPointer) {
	Free(uint64(pointer.Offset()))
}

func (m *alloc) Prune() {
	Prune()
}
