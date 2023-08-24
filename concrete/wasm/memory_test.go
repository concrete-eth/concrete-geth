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
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge"
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge/host"
	"github.com/stretchr/testify/require"
)

type mockMemory []byte

func newMockMemory() bridge.Memory {
	return &mockMemory{}
}

func (memory *mockMemory) Read(pointer bridge.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	if offset+size > uint32(len(*memory)) {
		panic("out of memory")
	}
	return (*memory)[offset : offset+size]
}

func (memory *mockMemory) Write(data []byte) bridge.MemPointer {
	size := len(data)
	if size == 0 {
		return bridge.NullPointer
	}
	offset := len(*memory)
	*memory = append(*memory, data...)
	var pointer bridge.MemPointer
	pointer.Pack(uint32(offset), uint32(size))
	return pointer
}

var _ bridge.Memory = (*mockMemory)(nil)

func testMemoryReadWrite(t *testing.T, memory bridge.Memory) {
	r := require.New(t)
	data := []byte{1, 2, 3, 4, 5}
	ptr := memory.Write(data)
	r.False(ptr.IsNull())
	readData := memory.Read(ptr)
	r.Equal(data, readData)
}

func testMemoryPutGetValues(t *testing.T, memory bridge.Memory) {
	r := require.New(t)
	// Test PutValue and GetValue
	value := []byte{0x01, 0x02, 0x03}
	pointer := bridge.PutValue(memory, value)
	r.NotEqual(bridge.NullPointer, pointer)
	result := bridge.GetValue(memory, pointer)
	r.Equal(value, result)

	// Test PutValues and GetValues
	values := [][]byte{{0x01, 0x02}, {0x03, 0x04}, {0x05, 0x06, 0x07}}
	pointer = bridge.PutValues(memory, values)
	r.NotEqual(bridge.NullPointer, pointer)
	resultValues := bridge.GetValues(memory, pointer)
	r.Equal(values, resultValues)

	// Test PutValues with empty slice
	pointer = bridge.PutValues(memory, [][]byte{})
	r.Equal(bridge.NullPointer, pointer)

	// Test GetValues with null pointer
	resultValues = bridge.GetValues(memory, bridge.NullPointer)
	r.Equal([][]byte{}, resultValues)
}

func TestMockMemoryReadWrite(t *testing.T) {
	memory := newMockMemory()
	testMemoryReadWrite(t, memory)
}

func TestMockMemoryPutGetValues(t *testing.T) {
	memory := newMockMemory()
	testMemoryPutGetValues(t, memory)
}

//go:embed testdata/blank.wasm
var blankCode []byte

func newWasmMemory() (bridge.Memory, bridge.Allocator) {
	envCall := host.NewEnvironmentCaller(func() api.Environment { return nil })
	mod, _, err := newModule(envCall, blankCode)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	return host.NewMemory(ctx, mod)
}

func TestWasmMemoryFree(t *testing.T) {
	memory, alloc := newWasmMemory()
	data := []byte{1, 2, 3, 4, 5}
	ptr := memory.Write(data)
	alloc.Free(ptr)
	require.Panics(t, func() { alloc.Free(ptr) })
}

func TestWasmMemoryPrune(t *testing.T) {
	memory, alloc := newWasmMemory()
	data := []byte{1, 2, 3, 4, 5}
	ptr1 := memory.Write(data)
	ptr2 := memory.Write(data)
	alloc.Prune()
	require.Panics(t, func() { alloc.Free(ptr1) })
	require.Panics(t, func() { alloc.Free(ptr2) })
}

func TestWasmMemoryReadWrite(t *testing.T) {
	memory, _ := newWasmMemory()
	testMemoryReadWrite(t, memory)
}

func TestWasmMemoryPutGetValues(t *testing.T) {
	memory, _ := newWasmMemory()
	testMemoryPutGetValues(t, memory)
}
