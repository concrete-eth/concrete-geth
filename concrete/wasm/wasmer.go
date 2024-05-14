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

//go:build !mips && !mipsle && !mips64 && !mips64le

// This file will ignored when building for mips to prevent compatibility
// issues.

package wasm

import (
	"sync"

	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/host"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func NewWasmerPrecompile(code []byte) concrete.Precompile {
	config := wasmer.NewConfig().UseCraneliftCompiler()
	return newWasmerPrecompile(code, config)
}

func NewWasmerPrecompileWithConfig(code []byte, config *wasmer.Config) concrete.Precompile {
	return newWasmerPrecompile(code, config)
}

func newWasmerModule(envCall host.WasmerHostFunc, code []byte, engineConfig *wasmer.Config) (*wasmer.Module, *wasmer.Instance, error) {
	engine := wasmer.NewEngineWithConfig(engineConfig)
	store := wasmer.NewStore(engine)
	module, err := wasmer.NewModule(store, code)

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
	memory      memory.Memory
	allocator   memory.Allocator
	environment *api.Env
	expIsStatic wasmer.NativeFunction
	expFinalise wasmer.NativeFunction
	expCommit   wasmer.NativeFunction
	expRun      wasmer.NativeFunction
}

func newWasmerPrecompile(code []byte, engineConfig *wasmer.Config) *wasmerPrecompile {
	pc := &wasmerPrecompile{}

	envCall := host.NewWasmerEnvironmentCaller(func() api.Environment { return pc.environment })
	module, instance, err := newWasmerModule(envCall, code, engineConfig)
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
	retPointer := memory.MemPointer(_retPointer)
	retErr := memory.GetError(p.memory, retPointer)
	p.allocator.Free(retPointer)
	return retErr
}

func (p *wasmerPrecompile) call_Bytes_Uint64(expFunc wasmer.NativeFunction, input []byte) uint64 {
	pointer := memory.PutValue(p.memory, input)
	defer p.allocator.Free(pointer)
	_ret, err := expFunc(int64(pointer))
	if err != nil && p.environment.Error() == nil {
		panic(err)
	}
	ret, _ := _ret.(int64)
	return uint64(ret)
}

func (p *wasmerPrecompile) call_Bytes_BytesErr(expFunc wasmer.NativeFunction, input []byte) ([]byte, error) {
	_retPointer := p.call_Bytes_Uint64(expFunc, input)
	retPointer := memory.MemPointer(_retPointer)
	retValues, retErr := memory.GetReturnWithError(p.memory, retPointer, true)
	if len(retValues) == 0 {
		return nil, retErr
	}
	return retValues[0], retErr
}

func (p *wasmerPrecompile) before(env api.Environment) {
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

func (p *wasmerPrecompile) after(env api.Environment) {
	p.environment = nil
	// p.allocator.Prune()
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

var _ concrete.Precompile = (*wasmerPrecompile)(nil)
