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

package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/concrete"
	cc_api "github.com/ethereum/go-ethereum/core/vm/concrete/api"
)

var (
	ConcretePrecompiledAddresses []common.Address
)

var ConcretePrecompiles = concrete.Precompiles

func init() {
	for k := range ConcretePrecompiles {
		ConcretePrecompiledAddresses = append(ConcretePrecompiledAddresses, k)
	}
}

// ActiveConcretePrecompiles returns the precompiles enabled with the current configuration.
func ActiveConcretePrecompiles() []common.Address {
	return ConcretePrecompiledAddresses
}

// RunConcretePrecompile runs and evaluates the output of a precompiled contract.
// It returns
// - the returned bytes,
// - the _remaining_ gas,
// - any error that occurred
func RunConcretePrecompile(evm concrete.EVM, addr common.Address, p concrete.Precompile, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	gasCost := p.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= gasCost

	if p.MutatesStorage(input) {
		if readOnly {
			return nil, suppliedGas, ErrWriteProtection
		}
	} else {
		evm = cc_api.NewReadOnlyEVM(evm)
	}

	api := cc_api.New(evm, addr)
	output, err := p.Run(api, input)
	return output, suppliedGas, err
}
