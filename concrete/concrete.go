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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/holiman/uint256"
)

type Precompile interface {
	IsStatic(input []byte) bool
	Run(env api.Environment, input []byte) ([]byte, error)
}

func RunPrecompile(p Precompile, env *api.Env, input []byte, gas uint64, value *uint256.Int) (ret []byte, remainingGas uint64, err error) {
	// We can either copy the input or trust the end developer to not modify it
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)

	static := env.Config().IsStatic
	if static && !p.IsStatic(inputCopy) {
		return nil, env.Gas(), api.ErrWriteProtection
	}

	contract := env.Contract()
	contract.Input = inputCopy
	contract.Gas = gas
	contract.Value = value

	defer func() {
		if r := recover(); r != nil {
			if revertErr := env.RevertError(); revertErr != nil {
				// Execution reverted
				ret = []byte(revertErr.Error()) // Return the revert reason
				err = api.ErrExecutionReverted
				remainingGas = env.Gas()
			} else if nonRevertErr := env.NonRevertError(); nonRevertErr != nil {
				// Explicit panic
				ret = nil
				err = nonRevertErr
				remainingGas = 0
			} else {
				// Runtime panic
				ret = nil
				if e, ok := r.(error); ok {
					err = e
				} else if m, ok := r.(string); ok {
					err = fmt.Errorf("runtime panic: %s", m)
				} else {
					err = fmt.Errorf("runtime panic: %v", r)
				}
				remainingGas = 0
			}
		}
	}()

	ret, err = p.Run(env, inputCopy)
	if err != nil {
		// Returning an error is equivalent to reverting
		ret = []byte(err.Error()) // Return the revert reason
		err = api.ErrExecutionReverted
	}

	return ret, env.Gas(), err
}

type PrecompileMap = map[common.Address]Precompile

type PrecompileRegistry interface {
	Precompile(address common.Address, blockNumber uint64) (Precompile, bool)
	Precompiles(blockNumber uint64) PrecompileMap
	PrecompiledAddresses(blockNumber uint64) []common.Address
}

type GenericPrecompileRegistry struct {
	startingBlocks []uint64
	precompiles    []PrecompileMap
	addresses      [][]common.Address
}

var _ PrecompileRegistry = (*GenericPrecompileRegistry)(nil)

func NewRegistry() *GenericPrecompileRegistry {
	return &GenericPrecompileRegistry{
		startingBlocks: make([]uint64, 0),
		precompiles:    make([]PrecompileMap, 0),
		addresses:      make([][]common.Address, 0),
	}
}

func (c *GenericPrecompileRegistry) index(blockNumber uint64) int {
	for idx, startingBlock := range c.startingBlocks {
		if blockNumber < startingBlock {
			return idx - 1
		}
	}
	return len(c.startingBlocks) - 1
}

func (c *GenericPrecompileRegistry) AddPrecompiles(startingBlock uint64, precompiles PrecompileMap) {
	idx := c.index(startingBlock)
	if idx >= 0 && c.startingBlocks[idx] == startingBlock {
		panic("precompiles already set for this block")
	}
	addresses := []common.Address{}
	for address := range precompiles {
		addresses = append(addresses, address)
	}

	c.startingBlocks = insert(c.startingBlocks, idx+1, startingBlock)
	c.precompiles = insert(c.precompiles, idx+1, precompiles)
	c.addresses = insert(c.addresses, idx+1, addresses)

}

func (c *GenericPrecompileRegistry) AddPrecompile(startingBlock uint64, address common.Address, precompile Precompile) {
	idx := c.index(startingBlock)
	if idx >= 0 && c.startingBlocks[idx] == startingBlock {
		// There already are precompiles for this block
		precompiles := c.precompiles[idx]
		if _, ok := precompiles[address]; ok {
			panic("precompile already set at this address for this block")
		}
		precompiles[address] = precompile
		c.addresses[idx] = append(c.addresses[idx], address)
	} else {
		c.startingBlocks = insert(c.startingBlocks, idx+1, startingBlock)
		c.precompiles = insert(c.precompiles, idx+1, PrecompileMap{address: precompile})
		c.addresses = insert(c.addresses, idx+1, []common.Address{address})
	}
}

func (c *GenericPrecompileRegistry) Precompile(address common.Address, blockNumber uint64) (Precompile, bool) {
	idx := c.index(blockNumber)
	if idx < 0 {
		return nil, false
	}
	pc, ok := c.precompiles[idx][address]
	if !ok {
		return nil, false
	}
	return pc, true
}

func (c *GenericPrecompileRegistry) Precompiles(blockNumber uint64) PrecompileMap {
	idx := c.index(blockNumber)
	if idx < 0 {
		return PrecompileMap{}
	}
	return c.precompiles[idx]
}

func (c *GenericPrecompileRegistry) PrecompiledAddresses(blockNumber uint64) []common.Address {
	idx := c.index(blockNumber)
	if idx < 0 {
		return []common.Address{}
	}
	return c.addresses[idx]
}

func insert[T any](slice []T, index int, value T) []T {
	if index < 0 || index > len(slice) {
		panic("index out of bounds")
	}
	if len(slice) == 0 {
		return []T{value}
	} else if index == len(slice) {
		return append(slice, value)
	} else {
		slice = append(slice[:index+1], slice[index:]...)
		slice[index] = value
		return slice
	}
}
