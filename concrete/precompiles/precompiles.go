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

package precompiles

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
)

type Version = struct {
	Major *big.Int `json:"major"`
	Minor *big.Int `json:"minor"`
	Patch *big.Int `json:"patch"`
}

func NewVersion(major, minor, patch int) Version {
	return Version{
		Major: big.NewInt(int64(major)),
		Minor: big.NewInt(int64(minor)),
		Patch: big.NewInt(int64(patch)),
	}
}

type PrecompileMetadata = struct {
	Address     common.Address `json:"addr"`
	Name        string         `json:"name"`
	Version     Version        `json:"version"`
	Author      string         `json:"author"`
	Description string         `json:"description"`
	Source      string         `json:"source"`
	ABI         string         `json:"ABI"`
}

var (
	precompiles          = make(map[common.Address]api.Precompile)
	precompiledAddresses = make([]common.Address, 0)
	precompileMetadata   = make([]PrecompileMetadata, 0)
	metadataByAddress    = make(map[common.Address]*PrecompileMetadata)
	metadataByName       = make(map[string]*PrecompileMetadata)
)

func AddPrecompileWithMetadata(addr common.Address, p api.Precompile, metadata PrecompileMetadata) error {
	if _, ok := precompiles[addr]; ok {
		return fmt.Errorf("precompile already exists at address %x", addr)
	}
	if len(metadata.Name) > 0 {
		if _, ok := metadataByName[metadata.Name]; ok {
			return fmt.Errorf("precompile already exists with name %s", metadata.Name)
		}
	}

	metadata.Address = addr

	precompiles[addr] = p
	precompiledAddresses = append(precompiledAddresses, addr)
	precompileMetadata = append(precompileMetadata, metadata)
	metadataByAddress[addr] = &metadata

	if metadata.Name != "" {
		metadataByName[metadata.Name] = &metadata
	}

	return nil
}

func AddPrecompile(addr common.Address, p api.Precompile) error {
	return AddPrecompileWithMetadata(addr, p, PrecompileMetadata{})
}

func GetPrecompile(addr common.Address) (api.Precompile, bool) {
	pc, ok := precompiles[addr]
	return pc, ok
}

func ActivePrecompiles() []common.Address {
	return precompiledAddresses
}

func GetPrecompileMetadataByAddress(addr common.Address) *PrecompileMetadata {
	pc, ok := metadataByAddress[addr]
	if !ok {
		return nil
	}
	return pc
}

func GetPrecompileMetadataByName(name string) *PrecompileMetadata {
	pc, ok := metadataByName[name]
	if !ok {
		return nil
	}
	return pc
}

func RunPrecompile(p api.Precompile, env *api.Env, input []byte, static bool) (ret []byte, remainingGas uint64, err error) {
	if p.IsStatic(input) && static {
		// TODO: error
		return nil, env.GetGasLeft(), errors.New("write protection")
	}
	output, err := p.Run(env, input)
	if err == nil {
		err = env.Error()
	}
	return output, env.Gas(), err
}
