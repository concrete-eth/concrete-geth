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

package concrete

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/vm/concrete/api"
	wasm_pcs "github.com/ethereum/go-ethereum/core/vm/concrete/wasm"
	ntv_pcs "github.com/therealbytes/concrete-native-precompiles/precompiles"
)

type API = api.API
type EVM = api.EVM
type StateDB = api.StateDB
type Precompile = api.Precompile

var Precompiles = ntv_pcs.Precompiles

// type EchoPrecompile struct {
// 	Precompile
// }

// func (p *EchoPrecompile) Run(api API, input []byte) ([]byte, error) {
// 	return input, nil
// }

// var Precompiles = map[common.Address]Precompile{
// 	ntv_pcs.BlankPrecompileAddress: &EchoPrecompile{ntv_pcs.Blank},
// }

func init() {
	for addr, pc := range wasm_pcs.Precompiles {
		if _, ok := Precompiles[addr]; ok {
			panic(fmt.Errorf("could not add wasm precompile: precompile address %x already taken", addr))
		}
		Precompiles[addr] = pc
	}
}
