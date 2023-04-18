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

package lib

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/concrete/api"
)

func GetData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return common.RightPadBytes(data[start:end], int(size))
}

var (
	errInvalidSelect = errors.New("invalid select value")
)

type PrecompileDemux map[int]api.Precompile

func (d PrecompileDemux) getSelect(input []byte) int {
	return int(new(big.Int).SetBytes(GetData(input, 0, 32)).Uint64())
}

func (d PrecompileDemux) MutatesStorage(input []byte) bool {
	sel := d.getSelect(input)
	pc, ok := d[sel]
	if !ok {
		return false
	}
	return pc.MutatesStorage(input[32:])
}

func (d PrecompileDemux) RequiredGas(input []byte) uint64 {
	sel := d.getSelect(input)
	pc, ok := d[sel]
	if !ok {
		return 0
	}
	return pc.RequiredGas(input[32:])
}

func (d PrecompileDemux) New(api api.API) error {
	for _, pc := range d {
		if err := pc.New(api); err != nil {
			return err
		}
	}
	return nil
}

func (d PrecompileDemux) Commit(api api.API) error {
	for _, pc := range d {
		if err := pc.Commit(api); err != nil {
			return err
		}
	}
	return nil
}

func (d PrecompileDemux) Run(api api.API, input []byte) ([]byte, error) {
	sel := d.getSelect(input)
	pc, ok := d[sel]
	if !ok {
		return nil, errInvalidSelect
	}
	return pc.Run(api, input[32:])
}
