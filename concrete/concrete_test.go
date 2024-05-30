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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/stretchr/testify/require"
)

type pcSet struct {
	blockNumber uint64
	precompiles PrecompileMap
}

type pcSingle struct {
	blockNumber uint64
	address     common.Address
	precompile  Precompile
}

type pcBlank struct{}

func (pc *pcBlank) IsStatic(input []byte) bool {
	return true
}

func (pc *pcBlank) Run(API api.Environment, input []byte) ([]byte, error) {
	return []byte{}, nil
}

var _ Precompile = &pcBlank{}

var (
	addrIncl1 = common.BytesToAddress([]byte{128})
	addrIncl2 = common.BytesToAddress([]byte{129})
	addrExcl  = common.BytesToAddress([]byte{130})
	// Block numbers are deliberately not in order
	pcSets = []pcSet{
		{
			blockNumber: 0,
			precompiles: PrecompileMap{
				addrIncl1: &pcBlank{},
				addrIncl2: &pcBlank{},
			},
		},
		{
			blockNumber: 20,
			precompiles: PrecompileMap{
				addrIncl2: &pcBlank{},
			},
		},
		{
			blockNumber: 10,
			precompiles: PrecompileMap{
				addrIncl1: &pcBlank{},
			},
		},
		{
			blockNumber: 40,
			precompiles: PrecompileMap{
				addrIncl1: &pcBlank{},
				addrIncl2: &pcBlank{},
			},
		},
		{
			blockNumber: 30,
			precompiles: PrecompileMap{},
		},
	}
	pcSingles = []pcSingle{
		{
			blockNumber: 5,
			address:     addrIncl1,
			precompile:  &pcBlank{},
		},
		{
			blockNumber: 15,
			address:     addrIncl2,
			precompile:  &pcBlank{},
		},
	}
)

func verifyPrecompileSet(t *testing.T, registry *GenericPrecompileRegistry, num uint64, p pcSet) {
	r := require.New(t)
	// Assert that PrecompiledAddresses returns the correct slice of addresses
	pcsAddr := registry.PrecompiledAddresses(num)
	expPcsAddr := make([]common.Address, 0, len(p.precompiles))
	for address := range p.precompiles {
		expPcsAddr = append(expPcsAddr, address)
	}
	r.ElementsMatch(expPcsAddr, pcsAddr)
	// Assert that all active addresses map to the correct precompile
	for address, setPc := range p.precompiles {
		registryPc, ok := registry.Precompile(address, num)
		r.True(ok)
		r.Equal(setPc, registryPc)
	}
	// Assert that inactive addresses do not map to a precompile
	pc, ok := registry.Precompile(addrExcl, num)
	r.Nil(pc)
	r.False(ok)
	// Assert that Precompiles returns the correct set of precompiles
	pcs := registry.Precompiles(num)
	r.Equal(p.precompiles, pcs)
}

func verifyPrecompileSingle(t *testing.T, registry *GenericPrecompileRegistry, num uint64, p pcSingle) {
	r := require.New(t)
	// Assert that PrecompiledAddresses returns the correct slice of addresses
	addresses := registry.PrecompiledAddresses(num)
	r.Len(addresses, 1)
	r.Equal(p.address, addresses[0])
	// Assert that all active addresses map to the correct precompile
	registryPc, ok := registry.Precompile(p.address, num)
	r.True(ok)
	r.Equal(p.precompile, registryPc)
	// Assert that inactive addresses do not map to a precompile
	pc, ok := registry.Precompile(addrExcl, num)
	r.Nil(pc)
	r.False(ok)
	// Assert that Precompiles returns the correct set of precompiles
	pcs := registry.Precompiles(num)
	r.Len(pcs, 1)
	r.Equal(p.precompile, pcs[p.address])
}

func TestPrecompileRegistry(t *testing.T) {
	t.Run("AddPrecompiles", func(t *testing.T) {
		registry := NewRegistry()
		for _, d := range pcSets {
			registry.AddPrecompiles(d.blockNumber, d.precompiles)
		}
		for _, d := range pcSets {
			require.Panics(t, func() {
				registry.AddPrecompiles(d.blockNumber, d.precompiles)
			})
		}
		for _, d := range pcSets {
			// Check that the precompiles are returned correctly for the first, second and last
			// block in each range
			for _, delta := range []uint64{0, 1, 9} {
				blockNumber := d.blockNumber + delta
				verifyPrecompileSet(t, registry, blockNumber, d)
			}
		}
	})
	t.Run("AddPrecompile", func(t *testing.T) {
		t.Run("OnEmpty", func(t *testing.T) {
			registry := NewRegistry()
			for _, d := range pcSingles {
				registry.AddPrecompile(d.blockNumber, d.address, d.precompile)
			}
			for _, d := range pcSingles {
				require.Panics(t, func() {
					registry.AddPrecompile(d.blockNumber, d.address, d.precompile)
				})
			}
			for _, d := range pcSingles {
				// Check that the precompiles are returned correctly for the first, second and last
				// block in each range
				for _, delta := range []uint64{0, 1, 9} {
					blockNumber := d.blockNumber + delta
					verifyPrecompileSingle(t, registry, blockNumber, d)
				}
			}
		})
		t.Run("OnExisting", func(t *testing.T) {
			registry := NewRegistry()
			for _, d := range pcSets {
				registry.AddPrecompiles(d.blockNumber, d.precompiles)
			}
			for _, d := range pcSingles {
				registry.AddPrecompile(d.blockNumber, d.address, d.precompile)
			}
			for _, d := range pcSingles {
				require.Panics(t, func() {
					registry.AddPrecompile(d.blockNumber, d.address, d.precompile)
				})
			}
			for _, d := range pcSets {
				blockNumber := d.blockNumber
				verifyPrecompileSet(t, registry, blockNumber, d)
			}
			for _, d := range pcSingles {
				blockNumber := d.blockNumber
				verifyPrecompileSingle(t, registry, blockNumber, d)
			}
		})
	})
}
