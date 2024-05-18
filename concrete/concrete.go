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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
)

type Environment = api.Environment

type Precompile interface {
	IsStatic(input []byte) bool
	Run(env Environment, input []byte) ([]byte, error)
}

func RunPrecompile(p Precompile, env *api.Env, input []byte, static bool) (ret []byte, remainingGas uint64, err error) {
	// We can either copy the input or trust the end developer to not modify it
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	if static && !p.IsStatic(inputCopy) {
		return nil, env.Gas(), api.ErrWriteProtection
	}
	output, err := p.Run(env, inputCopy)
	if env.Error() != nil {
		err = env.Error()
	} else if err != nil {
		err = api.ErrExecutionReverted
	}
	return output, env.Gas(), err
}

type PrecompileMap = map[common.Address]Precompile

type PrecompileRegistry interface {
	Precompile(address common.Address, blockNumber uint64) (Precompile, bool)
	Precompiles(blockNumber uint64) PrecompileMap
	ActivePrecompiles(blockNumber uint64) []common.Address
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
	for ii, startingBlock := range c.startingBlocks {
		if blockNumber < startingBlock {
			continue
		}
		if ii == len(c.startingBlocks)-1 {
			return ii
		}
		if blockNumber < c.startingBlocks[ii+1] {
			return ii
		}
	}
	return -1
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

	c.startingBlocks = insert[uint64](c.startingBlocks, idx+1, startingBlock)
	c.precompiles = insert[PrecompileMap](c.precompiles, idx+1, precompiles)
	c.addresses = insert[[]common.Address](c.addresses, idx+1, addresses)
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
	}

	c.startingBlocks = insert[uint64](c.startingBlocks, idx+1, startingBlock)
	c.precompiles = insert[PrecompileMap](c.precompiles, idx+1, PrecompileMap{address: precompile})
	c.addresses = insert[[]common.Address](c.addresses, idx+1, []common.Address{address})
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

func (c *GenericPrecompileRegistry) ActivePrecompiles(blockNumber uint64) []common.Address {
	idx := c.index(blockNumber)

	if idx < 0 {
		return []common.Address{}
	}
	return c.addresses[idx]
}

func insert[T any](slice []T, index int, value T) []T {
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
