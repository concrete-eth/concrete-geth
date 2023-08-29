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

package wasm

import (
	"context"
	_ "embed"
	"testing"

	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/host"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	"github.com/stretchr/testify/require"
)

type mockMemory []byte

func newMockMemory() memory.Memory {
	return &mockMemory{}
}

func (mem *mockMemory) Read(pointer memory.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	if offset+size > uint32(len(*mem)) {
		panic("out of memory")
	}
	return (*mem)[offset : offset+size]
}

func (mem *mockMemory) Write(data []byte) memory.MemPointer {
	size := len(data)
	if size == 0 {
		return memory.NullPointer
	}
	offset := len(*mem)
	*mem = append(*mem, data...)
	var pointer memory.MemPointer
	pointer.Pack(uint32(offset), uint32(size))
	return pointer
}

var _ memory.Memory = (*mockMemory)(nil)

func testMemoryReadWrite(t *testing.T, mem memory.Memory) {
	r := require.New(t)
	data := []byte{1, 2, 3, 4, 5}
	ptr := mem.Write(data)
	r.False(ptr.IsNull())
	readData := mem.Read(ptr)
	r.Equal(data, readData)
}

func testMemoryPutGetValues(t *testing.T, mem memory.Memory) {
	r := require.New(t)
	// Test PutValue and GetValue
	value := []byte{0x01, 0x02, 0x03}
	pointer := memory.PutValue(mem, value)
	r.NotEqual(memory.NullPointer, pointer)
	result := memory.GetValue(mem, pointer)
	r.Equal(value, result)

	// Test PutValues and GetValues
	values := [][]byte{{0x01, 0x02}, {0x03, 0x04}, {0x05, 0x06, 0x07}}
	pointer = memory.PutValues(mem, values)
	r.NotEqual(memory.NullPointer, pointer)
	resultValues := memory.GetValues(mem, pointer)
	r.Equal(values, resultValues)

	// Test PutValues with empty slice
	pointer = memory.PutValues(mem, [][]byte{})
	r.Equal(memory.NullPointer, pointer)

	// Test GetValues with null pointer
	resultValues = memory.GetValues(mem, memory.NullPointer)
	r.Equal([][]byte{}, resultValues)
}

func TestMockMemoryReadWrite(t *testing.T) {
	mem := newMockMemory()
	testMemoryReadWrite(t, mem)
}

func TestMockMemoryPutGetValues(t *testing.T) {
	mem := newMockMemory()
	testMemoryPutGetValues(t, mem)
}

//go:embed testdata/blank.wasm
var blankCode []byte

func newWazeroMemory() (memory.Memory, memory.Allocator) {
	envCall := host.NewWazeroEnvironmentCaller(func() api.Environment { return nil })
	mod, _, err := newWazeroModule(envCall, blankCode)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	return host.NewWazeroMemory(ctx, mod)
}

func TestWazeroMemoryFree(t *testing.T) {
	mem, alloc := newWazeroMemory()
	data := []byte{1, 2, 3, 4, 5}
	ptr := mem.Write(data)
	alloc.Free(ptr)
	require.Panics(t, func() { alloc.Free(ptr) })
}

func TestWazeroMemoryPrune(t *testing.T) {
	mem, alloc := newWazeroMemory()
	data := []byte{1, 2, 3, 4, 5}
	ptr1 := mem.Write(data)
	ptr2 := mem.Write(data)
	alloc.Prune()
	require.Panics(t, func() { alloc.Free(ptr1) })
	require.Panics(t, func() { alloc.Free(ptr2) })
}

func TestWazeroMemoryReadWrite(t *testing.T) {
	mem, _ := newWazeroMemory()
	testMemoryReadWrite(t, mem)
}

func TestWazeroMemoryPutGetValues(t *testing.T) {
	mem, _ := newWazeroMemory()
	testMemoryPutGetValues(t, mem)
}
