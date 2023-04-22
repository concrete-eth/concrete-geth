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

package core

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/concrete/api"
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/bridge"
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/bridge/wasm"
)

// Note: This uses a uint64 instead of two result values for compatibility with
// WebAssembly 1.0.

type memory struct{}

func (m *memory) Ref(data []byte) bridge.MemPointer {
	if len(data) == 0 {
		return bridge.NullPointer
	}
	offset := uint32(uintptr(unsafe.Pointer(&data[0])))
	size := uint32(len(data))
	var pointer bridge.MemPointer
	pointer.Pack(offset, size)
	return pointer
}

func (m *memory) Deref(pointer bridge.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(offset),
		Len:  uintptr(size),
		Cap:  uintptr(size),
	}))
}

var Memory wasm.Memory = &memory{}

func PutValue(value []byte) uint64 {
	return uint64(wasm.PutValue(Memory, value))
}

func GetValue(pointer uint64) []byte {
	return wasm.GetValue(Memory, bridge.MemPointer(pointer))
}

//go:wasm-module env
//export concrete_LogBridge
func _LogBridge(pointer uint64) uint64

func Log(a ...any) uint64 {
	msg := fmt.Sprintln(a...)
	pointer := PutValue([]byte(msg))
	return _LogBridge(pointer)
}

//go:wasm-module env
//export concrete_EvmBridge
func _EvmBridge(pointer uint64) uint64

func EvmBridge(pointer uint64) uint64 {
	return _EvmBridge(pointer)
}

//go:wasm-module env
//export concrete_StateDBBridge
func _StateDBBridge(pointer uint64) uint64

func StateDBBridge(pointer uint64) uint64 {
	return _StateDBBridge(pointer)
}

//go:wasm-module env
//export concrete_AddressBridge
func _AddressBridge(pointer uint64) uint64

func AddressBridge() common.Address {
	pointer := _AddressBridge(0)
	return common.BytesToAddress(Memory.Deref(bridge.MemPointer(pointer)))
}

var allocs = sync.Map{}

//export concrete_Malloc
func Malloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	buf := make([]byte, size)
	ptr := unsafe.Pointer(&buf[0])
	allocs.Store(uintptr(ptr), buf)
	return ptr
}

//export concrete_Free
func Free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	if _, ok := allocs.Load(uintptr(ptr)); ok {
		allocs.Delete(uintptr(ptr))
	} else {
		panic("free: invalid pointer")
	}
}

//export concrete_Prune
func Prune() {
	allocs = sync.Map{}
}

func NewAPI() api.API {
	evm := wasm.NewProxyEVM(Memory, EvmBridge, StateDBBridge)
	return api.New(evm, AddressBridge())
}

func NewStateAPI() api.API {
	statedb := wasm.NewProxyStateDB(Memory, StateDBBridge)
	return api.NewStateAPI(statedb, AddressBridge())
}
