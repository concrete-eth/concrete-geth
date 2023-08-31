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
	pointer, _ := malloc(int(size))
	return pointer.Uint64()
}

//export concrete_Free
func Free(_pointer uint64) {
	pointer := memory.MemPointer(_pointer)
	free(pointer)
}

//export concrete_Prune
func Prune() {
	prune()
}

func malloc(size int) (memory.MemPointer, []byte) {
	if size == 0 {
		return 0, []byte{}
	}
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	allocs[ptr] = buf
	var pointer memory.MemPointer
	pointer.Pack(uint32(ptr), uint32(size))
	return pointer, buf
}

func free(pointer memory.MemPointer) {
	ptr := uintptr(pointer.Offset())
	if _, ok := allocs[ptr]; ok {
		delete(allocs, ptr)
	} else {
		panic("free: invalid pointer")
	}
}

func prune() {
	allocs = make(map[uintptr][]byte)
}

type mem struct{}

func (m *mem) Allocator() memory.Allocator {
	return Allocator
}

func (m *mem) Write(data []byte) memory.MemPointer {
	if len(data) == 0 {
		return memory.NullPointer
	}
	pointer, buf := malloc(len(data))
	copy(buf, data)
	return pointer
}

func (m *mem) Read(pointer memory.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	buf := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(offset),
		//nolint:typecheck
		Len: uintptr(size),
		//nolint:typecheck
		Cap: uintptr(size),
	}))
	data := make([]byte, size)
	copy(data, buf)
	return data
}

type alloc struct{}

func (m *alloc) Malloc(size int) memory.MemPointer {
	pointer, _ := malloc(size)
	return pointer
}

func (m *alloc) Free(pointer memory.MemPointer) {
	free(pointer)
}

func (m *alloc) Prune() {
	prune()
}
