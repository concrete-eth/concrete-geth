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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
)

type Precompile interface {
	IsStatic(input []byte) bool
	Finalise(env api.Environment) error
	Commit(env api.Environment) error
	Run(env api.Environment, input []byte) ([]byte, error)
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
	precompiles          = make(map[common.Address]Precompile)
	precompiledAddresses = make([]common.Address, 0)
	precompileMetadata   = make([]PrecompileMetadata, 0)
	metadataByAddress    = make(map[common.Address]*PrecompileMetadata)
	metadataByName       = make(map[string]*PrecompileMetadata)
)

func AddPrecompileWithMetadata(addr common.Address, p Precompile, metadata PrecompileMetadata) error {
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

func AddPrecompile(addr common.Address, p Precompile) error {
	return AddPrecompileWithMetadata(addr, p, PrecompileMetadata{})
}

func GetPrecompile(addr common.Address) (Precompile, bool) {
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

func RunPrecompile(p Precompile, env *api.Env, input []byte, static bool) (ret []byte, remainingGas uint64, err error) {
	if static && !p.IsStatic(input) {
		return nil, env.GetGasLeft(), api.ErrWriteProtection
	}
	output, err := p.Run(env, input)
	if err == nil {
		err = env.Error()
	}
	return output, env.Gas(), err
}
