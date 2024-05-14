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

package host

import "errors"

var (
	ErrMemoryReadOutOfRange = errors.New("go: memory read out of range of memory size")
)

const (
	Malloc_WasmFuncName = "concrete_Malloc"
	Free_WasmFuncName   = "concrete_Free"
	Prune_WasmFuncName  = "concrete_Prune"
)
