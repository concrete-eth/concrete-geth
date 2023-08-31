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

package proxy

import (
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/wasm/memory"
)

type HostFuncCaller func(pointer uint64) uint64

func NewWasmProxyEnvironment(mem memory.Memory, allocator memory.Allocator, envCaller HostFuncCaller) *api.Env {
	return api.NewProxyEnvironment(
		func(op api.OpCode, env *api.Env, args [][]byte) ([][]byte, error) {
			args = append([][]byte{op.Encode()}, args...)
			argsPointer := memory.PutArgs(mem, args)
			retPointer := memory.MemPointer(envCaller(argsPointer.Uint64()))
			retValues := memory.GetValues(mem, retPointer, true)
			return retValues, nil
		},
	)
}
