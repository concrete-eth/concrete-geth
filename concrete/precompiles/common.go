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
	"math/big"
)

var (
	ErrMethodNotFound = errors.New("method not found")
	ErrInvalidInput   = errors.New("invalid input")
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
