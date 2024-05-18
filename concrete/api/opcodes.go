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

type OpCode byte

func (opcode OpCode) Encode() []byte {
	return []byte{byte(opcode)}
}

func (opcode *OpCode) Decode(data []byte) {
	*opcode = OpCode(data[0])
}

const (
	// Meta-ops
	ManyOps_OpCode OpCode = 0x04
	// Meta-env
	EnableGasMetering_OpCode OpCode = 0x08
	// Debug
	Debug_OpCode   OpCode = 0x0c
	TimeNow_OpCode OpCode = 0x0d
	// Utils
	Keccak256_OpCode OpCode = 0x10
	UseGas_OpCode    OpCode = 0x50
	// Internal reads
	GetAddress_OpCode         OpCode = 0x30
	GetGasLeft_OpCode         OpCode = 0x31
	GetBlockNumber_OpCode     OpCode = 0x32
	GetBlockGasLimit_OpCode   OpCode = 0x33
	GetBlockTimestamp_OpCode  OpCode = 0x34
	GetBlockDifficulty_OpCode OpCode = 0x35
	GetBlockBaseFee_OpCode    OpCode = 0x36
	GetBlockCoinbase_OpCode   OpCode = 0x37
	GetPrevRandom_OpCode      OpCode = 0x38
	GetBlockHash_OpCode       OpCode = 0x39
	GetBalance_OpCode         OpCode = 0x3a
	GetTxGasPrice_OpCode      OpCode = 0x3b
	GetTxOrigin_OpCode        OpCode = 0x3c
	GetCallData_OpCode        OpCode = 0x3d
	GetCallDataSize_OpCode    OpCode = 0x3e
	GetCaller_OpCode          OpCode = 0x3f
	GetCallValue_OpCode       OpCode = 0x40
	StorageLoad_OpCode        OpCode = 0x41
	GetCode_OpCode            OpCode = 0x42
	GetCodeSize_OpCode        OpCode = 0x43
	// Internal writes
	StorageStore_OpCode OpCode = 0x51
	Log_OpCode          OpCode = 0x52
	// External reads
	GetExternalBalance_OpCode  OpCode = 0x60
	CallStatic_OpCode          OpCode = 0x61
	GetExternalCode_OpCode     OpCode = 0x62
	GetExternalCodeSize_OpCode OpCode = 0x63
	GetExternalCodeHash_OpCode OpCode = 0x64
	// External writes
	Call_OpCode         OpCode = 0x70
	CallDelegate_OpCode OpCode = 0x71
	Create_OpCode       OpCode = 0x72
	Create2_OpCode      OpCode = 0x73
)
