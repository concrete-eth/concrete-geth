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
	"context"

	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	wz_api "github.com/tetratelabs/wazero/api"
)

// TODO: how is wasm panic handled?

type wazeroAllocator struct {
	ctx       context.Context
	module    wz_api.Module
	expMalloc wz_api.Function
	expFree   wz_api.Function
	expPrune  wz_api.Function
}

func NewWazeroAllocator(ctx context.Context, module wz_api.Module) memory.Allocator {
	return &wazeroAllocator{ctx: ctx, module: module}
}

func (a *wazeroAllocator) Malloc(size uint32) memory.MemPointer {
	if size == 0 {
		return memory.NullPointer
	}
	if a.expMalloc == nil {
		a.expMalloc = a.module.ExportedFunction(Malloc_WasmFuncName)
	}
	_offset, err := a.expMalloc.Call(a.ctx, uint64(size))
	if err != nil {
		panic(err)
	}
	var pointer memory.MemPointer
	pointer.Pack(uint32(_offset[0]), size)
	return pointer
}

func (a *wazeroAllocator) Free(pointer memory.MemPointer) {
	if pointer.IsNull() {
		return
	}
	if a.expFree == nil {
		a.expFree = a.module.ExportedFunction(Free_WasmFuncName)
	}
	_, err := a.expFree.Call(a.ctx, uint64(pointer.Offset()))
	if err != nil {
		panic(err)
	}
}

func (a *wazeroAllocator) Prune() {
	if a.expPrune == nil {
		a.expPrune = a.module.ExportedFunction(Prune_WasmFuncName)
	}
	_, err := a.expPrune.Call(a.ctx)
	if err != nil {
		panic(err)
	}
}

var _ memory.Allocator = (*wazeroAllocator)(nil)

type wazeroMemory struct {
	wazeroAllocator
}

func NewWazeroMemory(ctx context.Context, module wz_api.Module) (memory.Memory, memory.Allocator) {
	alloc := &wazeroAllocator{ctx: ctx, module: module}
	return &wazeroMemory{wazeroAllocator: *alloc}, alloc
}

func NewWazeroMemoryFromAlloc(alloc *wazeroAllocator) memory.Allocator {
	return &wazeroMemory{*alloc}
}

func (m *wazeroMemory) Write(data []byte) memory.MemPointer {
	if len(data) == 0 {
		return memory.NullPointer
	}
	pointer := m.Malloc(uint32(len(data)))
	ok := m.module.Memory().Write(pointer.Offset(), data)
	if !ok {
		panic(ErrMemoryReadOutOfRange)
	}
	return pointer
}

func (m *wazeroMemory) Read(pointer memory.MemPointer) []byte {
	if pointer.IsNull() {
		return []byte{}
	}
	output, ok := m.module.Memory().Read(pointer.Offset(), pointer.Size())
	if !ok {
		panic(ErrMemoryReadOutOfRange)
	}
	return output
}

var _ memory.Memory = (*wazeroMemory)(nil)

type WazeroHostFunc func(ctx context.Context, module wz_api.Module, pointer uint64) uint64

func NewWazeroEnvironmentCaller(apiGetter func() api.Environment) WazeroHostFunc {
	return func(ctx context.Context, module wz_api.Module, _pointer uint64) uint64 {
		pointer := memory.MemPointer(_pointer)
		env := apiGetter()
		mem, _ := NewWazeroMemory(ctx, module)

		args := memory.GetArgs(mem, pointer)
		var opcode api.OpCode
		opcode.Decode(args[0])
		args = args[1:]

		out, _ := env.Execute(opcode, args)

		// TODO: halt execution on error [?]

		return memory.PutValues(mem, out).Uint64()
	}
}
