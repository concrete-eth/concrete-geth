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
	"sync"

	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge"
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge/host"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func NewWasmerPrecompile(code []byte) precompiles.Precompile {
	return newWasmerPrecompile(code)
}

func newWasmerModule(envCall host.WasmerHostFunc, code []byte) (*wasmer.Module, *wasmer.Instance, error) {
	config := wasmer.NewConfig().UseCraneliftCompiler()
	engine := wasmer.NewEngineWithConfig(config)
	store := wasmer.NewStore(engine)
	module, err := wasmer.NewModule(store, code)

	// TODO: memory size [?]

	if err != nil {
		return nil, nil, err
	}

	wasiEnv, err := wasmer.NewWasiStateBuilder("wasi-program").Finalize()
	if err != nil {
		return nil, nil, err
	}
	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		return nil, nil, err
	}

	wasmerEnv := host.NewWasmerEnvironment()

	importObject.Register(
		"env",
		map[string]wasmer.IntoExtern{
			Environment_WasmFuncName: wasmer.NewFunctionWithEnvironment(
				store,
				wasmer.NewFunctionType(
					wasmer.NewValueTypes(wasmer.I64),
					wasmer.NewValueTypes(wasmer.I64),
				),
				wasmerEnv,
				envCall,
			),
		},
	)

	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, nil, err
	}
	wasmerEnv.Init(instance)

	return module, instance, nil
}

type wasmerPrecompile struct {
	instance    *wasmer.Instance
	module      *wasmer.Module
	mutex       sync.Mutex
	memory      bridge.Memory
	allocator   bridge.Allocator
	environment api.Environment
	expIsStatic wasmer.NativeFunction
	expFinalise wasmer.NativeFunction
	expCommit   wasmer.NativeFunction
	expRun      wasmer.NativeFunction
}

func newWasmerPrecompile(code []byte) *wasmerPrecompile {
	pc := &wasmerPrecompile{}

	envCall := host.NewWasmerEnvironmentCaller(func() api.Environment { return pc.environment })
	module, instance, err := newWasmerModule(envCall, code)
	if err != nil {
		panic(err)
	}

	pc.instance = instance
	pc.module = module
	pc.memory, pc.allocator = host.NewWasmerMemory(instance)

	pc.expIsStatic, err = instance.Exports.GetFunction(IsStatic_WasmFuncName)
	if err != nil {
		panic(err)
	}
	pc.expFinalise, err = instance.Exports.GetFunction(Finalise_WasmFuncName)
	if err != nil {
		panic(err)
	}
	pc.expCommit, err = instance.Exports.GetFunction(Commit_WasmFuncName)
	if err != nil {
		panic(err)
	}
	pc.expRun, err = instance.Exports.GetFunction(Run_WasmFuncName)
	if err != nil {
		panic(err)
	}

	return pc
}

func (p *wasmerPrecompile) close() {
	return // TODO
}

func (p *wasmerPrecompile) call__Uint64(expFunc wasmer.NativeFunction) uint64 {
	_ret, err := expFunc()
	if err != nil {
		panic(err)
	}
	ret, _ := _ret.(int64)
	return uint64(ret)
}

func (p *wasmerPrecompile) call__Err(expFunc wasmer.NativeFunction) error {
	_retPointer := p.call__Uint64(expFunc)
	retPointer := bridge.MemPointer(_retPointer)
	retErr := bridge.GetError(p.memory, retPointer)
	return retErr
}

func (p *wasmerPrecompile) call_Bytes_Uint64(expFunc wasmer.NativeFunction, input []byte) uint64 {
	pointer := bridge.PutValue(p.memory, input)
	defer p.allocator.Free(pointer)
	_ret, err := expFunc(int64(pointer.Uint64()))
	if err != nil {
		panic(err)
	}
	ret, _ := _ret.(int64)
	return uint64(ret)
}

func (p *wasmerPrecompile) call_Bytes_BytesErr(expFunc wasmer.NativeFunction, input []byte) ([]byte, error) {
	_retPointer := p.call_Bytes_Uint64(expFunc, input)
	retPointer := bridge.MemPointer(_retPointer)
	retValues, retErr := bridge.GetReturnWithError(p.memory, retPointer)
	return retValues[0], retErr
}

func (p *wasmerPrecompile) before(env api.Environment) {
	if env != nil {
		envImpl, ok := env.(*api.Env)
		if !ok {
			panic("invalid environment")
		}
		if !envImpl.Config().Trusted {
			panic("untrusted environment")
		}
	}
	p.mutex.Lock()
	p.environment = env
}

func (p *wasmerPrecompile) after(env api.Environment) {
	p.environment = nil
	p.allocator.Prune()
	p.mutex.Unlock()
}

func (p *wasmerPrecompile) IsStatic(input []byte) bool {
	p.before(nil)
	defer p.after(nil)
	return p.call_Bytes_Uint64(p.expIsStatic, input) != 0
}

func (p *wasmerPrecompile) Finalise(env api.Environment) error {
	p.before(env)
	defer p.after(env)
	return p.call__Err(p.expFinalise)
}

func (p *wasmerPrecompile) Commit(env api.Environment) error {
	p.before(env)
	defer p.after(env)
	return p.call__Err(p.expCommit)
}

func (p *wasmerPrecompile) Run(env api.Environment, input []byte) ([]byte, error) {
	p.before(env)
	defer p.after(env)
	return p.call_Bytes_BytesErr(p.expRun, input)
}

var _ precompiles.Precompile = (*wasmerPrecompile)(nil)
