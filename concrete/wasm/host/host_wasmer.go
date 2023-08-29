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

package host

import (
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	"github.com/wasmerio/wasmer-go/wasmer"
)

type wasmerAllocator struct {
	instance  *wasmer.Instance
	memory    *wasmer.Memory
	expMalloc wasmer.NativeFunction
	expFree   wasmer.NativeFunction
	expPrune  wasmer.NativeFunction
}

func NewWasmerAllocator(instance *wasmer.Instance) memory.Allocator {
	mem, err := instance.Exports.GetMemory("memory")
	if err != nil {
		panic(err)
	}
	if mem == nil {
		panic("memory not exported")
	}
	return &wasmerAllocator{instance: instance, memory: mem}
}

func (a *wasmerAllocator) Malloc(size uint32) memory.MemPointer {
	if size == 0 {
		return memory.NullPointer
	}
	if a.expMalloc == nil {
		var err error
		a.expMalloc, err = a.instance.Exports.GetFunction(Malloc_WasmFuncName)
		if err != nil {
			panic(err)
		}
	}
	_offset, err := a.expMalloc(int64(size))
	if err != nil {
		panic(err)
	}
	var pointer memory.MemPointer
	offset, _ := _offset.(int64)
	pointer.Pack(uint32(offset), size)
	return pointer
}

func (a *wasmerAllocator) Free(pointer memory.MemPointer) {
	if pointer.IsNull() {
		return
	}
	if a.expFree == nil {
		var err error
		a.expFree, err = a.instance.Exports.GetFunction(Free_WasmFuncName)
		if err != nil {
			panic(err)
		}
	}
	_, err := a.expFree(int64(pointer.Offset()))
	if err != nil {
		panic(err)
	}
}

func (a *wasmerAllocator) Prune() {
	if a.expPrune == nil {
		var err error
		a.expPrune, err = a.instance.Exports.GetFunction(Prune_WasmFuncName)
		if err != nil {
			panic(err)
		}
	}
	_, err := a.expPrune()
	if err != nil {
		panic(err)
	}
}

var _ memory.Allocator = (*wasmerAllocator)(nil)

type wasmerMemory struct {
	wasmerAllocator
}

func NewWasmerMemory(instance *wasmer.Instance) (memory.Memory, memory.Allocator) {
	alloc := NewWasmerAllocator(instance)
	return &wasmerMemory{wasmerAllocator: *alloc.(*wasmerAllocator)}, alloc
}

func NewWasmerMemoryFromAlloc(alloc *wasmerAllocator) memory.Allocator {
	return &wasmerMemory{*alloc}
}

func (m *wasmerMemory) Write(data []byte) memory.MemPointer {
	if len(data) == 0 {
		return memory.NullPointer
	}
	pointer := m.Malloc(uint32(len(data)))
	offset, size := pointer.Unpack()
	memSize := m.memory.Size()
	if uint(offset+size) >= memSize.ToBytes() {
		panic(ErrMemoryReadOutOfRange)
	}
	mem := m.memory.Data()
	copy(mem[offset:], data)
	return pointer
}

func (m *wasmerMemory) Read(pointer memory.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	offset, size := pointer.Unpack()
	memSize := m.memory.Size()
	if uint(offset+size) >= memSize.ToBytes() {
		panic(ErrMemoryReadOutOfRange)
	}
	mem := m.memory.Data()
	output := make([]byte, size)
	copy(output, mem[offset:])
	return output
}

var _ memory.Memory = (*wasmerMemory)(nil)

type WasmerHostFunc func(interface{}, []wasmer.Value) ([]wasmer.Value, error)

type WasmerEnvironment struct {
	instance *wasmer.Instance
}

func NewWasmerEnvironment() *WasmerEnvironment {
	return &WasmerEnvironment{}
}

func (e *WasmerEnvironment) Init(instance *wasmer.Instance) {
	e.instance = instance
}

func NewWasmerEnvironmentCaller(apiGetter func() api.Environment) WasmerHostFunc {
	return func(wasmerEnv interface{}, _pointer []wasmer.Value) ([]wasmer.Value, error) {
		pointer := memory.MemPointer(_pointer[0].I64())
		env := apiGetter()
		mem, _ := NewWasmerMemory(wasmerEnv.(*WasmerEnvironment).instance)

		args := memory.GetArgs(mem, pointer)
		var opcode api.OpCode
		opcode.Decode(args[0])
		args = args[1:]

		out, err := env.Execute(opcode, args)
		if err != nil {
			return nil, err
		}

		retPointer := memory.PutValues(mem, out)
		return []wasmer.Value{wasmer.NewI64(int64(retPointer))}, nil
	}
}
