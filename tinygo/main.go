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
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
	"github.com/ethereum/go-ethereum/concrete/wasm/proxy"
	"github.com/ethereum/go-ethereum/tinygo/infra"
)

var precompile concrete.Precompile

func WasmWrap(pc concrete.Precompile) {
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
	return proxy.NewWasmProxyEnvironment(infra.Memory, infra.Allocator, environment)
}

//export concrete_IsStatic
func isStatic(pointer uint64) uint64 {
	input := memory.GetValue(infra.Memory, memory.MemPointer(pointer))
	if precompile.IsStatic(input) {
		return 1
	} else {
		return 0
	}
}

//export concrete_Run
func run(pointer uint64) uint64 {
	env := newEnvironment()
	input := memory.GetValue(infra.Memory, memory.MemPointer(pointer))
	output, err := precompile.Run(env, input)
	return memory.PutReturnWithError(infra.Memory, [][]byte{output}, err).Uint64()
}
