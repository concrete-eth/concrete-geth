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

package tinygo

import (
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge"
	"github.com/ethereum/go-ethereum/concrete/wasm/bridge/wasm"
	"github.com/ethereum/go-ethereum/tinygo/infra"
)

var precompile api.Precompile

func WasmWrap(pc api.Precompile) {
	precompile = pc
}

// Note: This uses a uint64 instead of two result values for compatibility with
// WebAssembly 1.0.

//go:wasm-module env
//export concrete_Environment
func _environment(pointer uint64) uint64

func environment(pointer uint64) uint64 {
	return _environment(pointer)
}

func newEnvironment() *api.Env {
	return wasm.NewProxyEnvironment(infra.Memory, infra.Allocator, environment)
}

//export concrete_IsStatic
func isStatic(pointer uint64) uint64 {
	input := bridge.GetValue(infra.Memory, bridge.MemPointer(pointer))
	if precompile.IsStatic(input) {
		return 1
	} else {
		return 0
	}
}

//export concrete_Finalise
func finalise() uint64 {
	env := newEnvironment()
	precompile.Finalise(env)
	return bridge.NullPointer.Uint64()
}

//export concrete_Commit
func commit() uint64 {
	env := newEnvironment()
	precompile.Commit(env)
	return bridge.NullPointer.Uint64()
}

//export concrete_Run
func run(pointer uint64) uint64 {
	input := bridge.GetValue(infra.Memory, bridge.MemPointer(pointer))
	env := newEnvironment()
	output, err := precompile.Run(env, input)
	return bridge.PutReturnWithError(infra.Memory, [][]byte{output}, err).Uint64()
}
