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

package main

import (
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/bridge"
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/bridge/wasm"
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/tinygo/core"
	precompile "github.com/therealbytes/concrete-wasm-precompiles/precompiles/target"
)

func main() {}

//export concrete_IsPure
func IsPure() uint64 {
	if precompile.IsPure() {
		return 1
	} else {
		return 0
	}
}

//export concrete_MutatesStorage
func MutatesStorage(pointer uint64) uint64 {
	input := core.GetValue(pointer)
	if precompile.MutatesStorage(input) {
		return 1
	} else {
		return 0
	}
}

//export concrete_RequiredGas
func RequiredGas(pointer uint64) uint64 {
	input := core.GetValue(pointer)
	gas := precompile.RequiredGas(input)
	return uint64(gas)
}

//export concrete_New
func New() uint64 {
	precompile.New(core.NewStateAPI())
	return bridge.NullPointer.Uint64()
}

//export concrete_Commit
func Commit() uint64 {
	precompile.Commit(core.NewStateAPI())
	return bridge.NullPointer.Uint64()
}

//export concrete_Run
func Run(pointer uint64) uint64 {
	input := core.GetValue(pointer)
	api := core.NewAPI()
	output, err := precompile.Run(api, input)
	return wasm.PutReturnWithError(core.Memory, [][]byte{output}, err).Uint64()
}
