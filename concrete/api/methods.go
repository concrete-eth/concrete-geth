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

package api

import "errors"

var (
	ErrEnvNotTrusted   = errors.New("environment not trusted")
	ErrWriteProtection = errors.New("write protection")
	ErrOutOfGas        = errors.New("out of gas")
	ErrGasUintOverflow = errors.New("gas uint64 overflow")
	ErrFeatureDisabled = errors.New("feature disabled")
	ErrInvalidOpCode   = errors.New("invalid opcode")
	ErrInvalidInput    = errors.New("invalid input")
	ErrNoData          = errors.New("no data")
)

const (
	GasQuickStep   uint64 = 2
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 8
	GasSlowStep    uint64 = 10
	GasExtStep     uint64 = 20
)

type (
	executionFunc func(env *Env, args [][]byte) ([][]byte, error)
	gasFunc       func(env *Env, args [][]byte) (uint64, error)
)

type operation struct {
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
	trusted     bool
	static      bool
}

type JumpTable [256]*operation
