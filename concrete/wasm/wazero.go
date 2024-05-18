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
	"sync"

	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/host"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	"github.com/tetratelabs/wazero"
	wz_api "github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Note: For trusted use only. Precompiles can trigger a panic in the host.

func NewWazeroPrecompile(code []byte) concrete.Precompile {
	config := wazero.NewRuntimeConfigCompiler()
	return newWazeroPrecompile(code, config)
}

func NewWazeroPrecompileWithConfig(code []byte, config wazero.RuntimeConfig) concrete.Precompile {
	return newWazeroPrecompile(code, config)
}

func newWazeroModule(envCall host.WazeroHostFunc, code []byte, runtimeConfig wazero.RuntimeConfig) (wz_api.Module, wazero.Runtime, error) {
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(envCall).Export(Environment_WasmFuncName).
		Instantiate(ctx)
	if err != nil {
		return nil, nil, err
	}
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	mod, err := r.Instantiate(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	return mod, r, nil
}

type wazeroPrecompile struct {
	runtime     wazero.Runtime
	module      wz_api.Module
	mutex       sync.Mutex
	memory      memory.Memory
	allocator   memory.Allocator
	environment *api.Env
	expIsStatic wz_api.Function
	// expFinalise wz_api.Function
	// expCommit   wz_api.Function
	expRun wz_api.Function
}

func newWazeroPrecompile(code []byte, runtimeConfig wazero.RuntimeConfig) *wazeroPrecompile {
	pc := &wazeroPrecompile{}

	envCall := host.NewWazeroEnvironmentCaller(func() api.Environment { return pc.environment })
	mod, r, err := newWazeroModule(envCall, code, runtimeConfig)
	if err != nil {
		panic(err)
	}

	pc.runtime = r
	pc.module = mod
	pc.memory, pc.allocator = host.NewWazeroMemory(context.Background(), mod)

	pc.expIsStatic = mod.ExportedFunction(IsStatic_WasmFuncName)
	if pc.expIsStatic == nil {
		panic("isStatic not exported")
	}
	// pc.expFinalise = mod.ExportedFunction(Finalise_WasmFuncName)
	// if pc.expFinalise == nil {
	// 	panic("finalise not exported")
	// }
	// pc.expCommit = mod.ExportedFunction(Commit_WasmFuncName)
	// if pc.expCommit == nil {
	// 	panic("commit not exported")
	// }
	pc.expRun = mod.ExportedFunction(Run_WasmFuncName)
	if pc.expRun == nil {
		panic("run not exported")
	}

	return pc
}

func (p *wazeroPrecompile) call__Uint64(expFunc wz_api.Function) uint64 {
	ctx := context.Background()
	_ret, err := expFunc.Call(ctx)
	if err != nil {
		panic(err)
	}
	return _ret[0]
}

func (p *wazeroPrecompile) call__Err(expFunc wz_api.Function) error {
	_retPointer := p.call__Uint64(expFunc)
	retPointer := memory.MemPointer(_retPointer)
	retErr := memory.GetError(p.memory, retPointer)
	p.allocator.Free(retPointer)
	return retErr
}

func (p *wazeroPrecompile) call_Bytes_Uint64(expFunc wz_api.Function, input []byte) (ret uint64) {
	defer func() {
		if r := recover(); r != nil {
			if p.environment.Error() == nil {
				panic(r)
			}
			ret = memory.NullPointer.Uint64()
		}
	}()
	ctx := context.Background()
	pointer := memory.PutValue(p.memory, input)
	defer p.allocator.Free(pointer)
	_ret, err := expFunc.Call(ctx, pointer.Uint64())
	if err != nil {
		panic(err)
	}
	return _ret[0]
}

func (p *wazeroPrecompile) call_Bytes_BytesErr(expFunc wz_api.Function, input []byte) ([]byte, error) {
	_retPointer := p.call_Bytes_Uint64(expFunc, input)
	retPointer := memory.MemPointer(_retPointer)
	retValues, retErr := memory.GetReturnWithError(p.memory, retPointer, true)
	if len(retValues) == 0 {
		return nil, retErr
	}
	return retValues[0], retErr
}

func (p *wazeroPrecompile) before(env api.Environment) {
	var envImpl *api.Env
	if env != nil {
		envImpl = env.(*api.Env)
		if !envImpl.Config().Trusted {
			panic("untrusted environment")
		}
	}
	p.mutex.Lock()
	p.environment = envImpl
}

func (p *wazeroPrecompile) after(env api.Environment) {
	p.environment = nil
	// p.allocator.Prune()
	p.mutex.Unlock()
}

func (p *wazeroPrecompile) IsStatic(input []byte) bool {
	p.before(nil)
	defer p.after(nil)
	return p.call_Bytes_Uint64(p.expIsStatic, input) != 0
}

func (p *wazeroPrecompile) Run(env api.Environment, input []byte) ([]byte, error) {
	p.before(env)
	defer p.after(env)
	return p.call_Bytes_BytesErr(p.expRun, input)
}

var _ concrete.Precompile = (*wazeroPrecompile)(nil)
