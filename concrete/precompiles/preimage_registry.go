package precompiles

import (
	_ "embed"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

//go:embed sol/abi/PreimageRegistry.abi
var preimageRegistryABI string

var (
	AddPreimagePerByteGas = uint64(10)
	HasPreimageGas        = uint64(10)
	GetPreimageSizeGas    = uint64(10)
	GetPreimagePerByteGas = uint64(10)

	AddBigPreimagePerByteGas = uint64(10) // disabled
	HasBigPreimageGas        = uint64(10)
	GetBigPreimageSizeGas    = uint64(10)
	GetBigPreimagePerByteGas = uint64(10)
)

func init() {
	abiReader := strings.NewReader(preimageRegistryABI)
	ABI, err := abi.JSON(abiReader)
	if err != nil {
		panic(err)
	}

	preimageRegistry := lib.NewPrecompileWithABI(ABI, map[string]lib.MethodPrecompile{
		"addPreimage":     &addPreimage{},
		"hasPreimage":     &hasPreimage{},
		"getPreimageSize": &getPreimageSize{},
		"getPreimage":     &getPreimage{},
	})
	AddPrecompile(api.PreimageRegistryAddress, preimageRegistry, PrecompileMetadata{
		Name:        "PreimageRegistry",
		Version:     "0.1.0",
		Author:      "The concrete-geth Authors",
		Description: "A registry of stored preimages indexed by their hash.",
		Source:      "https://github.com/therealbytes/concrete-geth/tree/concrete/concrete/precompiles/pi_registry.go",
	})

	bigPreimageRegistry := lib.NewPrecompileWithABI(ABI, map[string]lib.MethodPrecompile{
		"addPreimage":     &addBigPreimage{},
		"hasPreimage":     &hasBigPreimage{},
		"getPreimageSize": &getBigPreimageSize{},
		"getPreimage":     &getBigPreimage{},
	})
	AddPrecompile(api.BigPreimageRegistryAddress, bigPreimageRegistry, PrecompileMetadata{
		Name:        "BigPreimageRegistry",
		Version:     "0.1.0",
		Author:      "The concrete-geth Authors",
		Description: "A registry of stored preimage merkle trees indexed by their root hash.",
		Source:      "https://github.com/therealbytes/concrete-geth/tree/concrete/concrete/precompiles/pi_registry.go",
	})
}

type addPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *addPreimage) RequiredGas(input []byte) uint64 {
	if len(input) < 64 {
		return 0
	}
	return AddPreimagePerByteGas * uint64(len(input)-64)
}

func (p *addPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		preimage := args[0].([]byte)
		hash := concrete.Persistent().AddPreimage(preimage)
		return []interface{}{hash}, nil
	}, concrete, input)
}

type hasPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *hasPreimage) RequiredGas(input []byte) uint64 {
	return HasPreimageGas
}

func (p *hasPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		hash := common.Hash(args[0].([32]byte))
		has := concrete.Persistent().HasPreimage(hash)
		return []interface{}{has}, nil
	}, concrete, input)
}

type getPreimageSize struct {
	lib.BlankMethodPrecompile
}

func (p *getPreimageSize) RequiredGas(input []byte) uint64 {
	return GetPreimageSizeGas
}

func (p *getPreimageSize) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		hash := common.Hash(args[0].([32]byte))
		var size int
		if concrete.Persistent().HasPreimage(hash) {
			size = concrete.StateDB().GetPersistentPreimageSize(hash)
		} else {
			size = 0
		}
		return []interface{}{big.NewInt(int64(size))}, nil
	}, concrete, input)
}

type getPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *getPreimage) RequiredGas(input []byte) uint64 {
	return p.CallRequiredGasWithArgs(func(args []interface{}) uint64 {
		size := args[0].(*big.Int)
		return GetPreimagePerByteGas * uint64(size.Int64())
	}, input)
}

func (p *getPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		size := int(args[0].(*big.Int).Int64())
		hash := common.Hash(args[1].([32]byte))
		per := concrete.Persistent()
		if !per.HasPreimage(hash) {
			return []interface{}{[]byte{}}, errors.New("preimage not found")
		}
		realSize := per.GetPreimageSize(hash)
		if size < realSize {
			return []interface{}{[]byte{}}, errors.New("provided preimage size too small")
		}
		preimage := per.GetPreimage(hash)
		return []interface{}{preimage}, nil
	}, concrete, input)
}

type addBigPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *addBigPreimage) RequiredGas(input []byte) uint64 {
	if len(input) < 64 {
		return 0
	}
	return AddBigPreimagePerByteGas * uint64(len(input)-64)
}

func (p *addBigPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		hash := common.Hash{}
		return []interface{}{hash}, errors.New("not implemented")
	}, concrete, input)
}

type hasBigPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *hasBigPreimage) RequiredGas(input []byte) uint64 {
	return HasBigPreimageGas
}

func (p *hasBigPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		hash := common.Hash(args[0].([32]byte))
		store := api.NewPersistentBigPreimageStore(concrete, -1, -1)
		has := store.HasPreimage(hash)
		return []interface{}{has}, nil
	}, concrete, input)
}

type getBigPreimageSize struct {
	lib.BlankMethodPrecompile
}

func (p *getBigPreimageSize) RequiredGas(input []byte) uint64 {
	return GetBigPreimageSizeGas
}

func (p *getBigPreimageSize) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		hash := common.Hash(args[0].([32]byte))
		store := api.NewPersistentBigPreimageStore(concrete, -1, -1)
		size := store.GetPreimageSize(hash)
		return []interface{}{big.NewInt(int64(size))}, nil
	}, concrete, input)
}

type getBigPreimage struct {
	lib.BlankMethodPrecompile
}

func (p *getBigPreimage) RequiredGas(input []byte) uint64 {
	return p.CallRequiredGasWithArgs(func(args []interface{}) uint64 {
		size := args[0].(*big.Int)
		return GetBigPreimagePerByteGas * uint64(size.Int64())
	}, input)
}

func (p *getBigPreimage) Run(concrete api.API, input []byte) ([]byte, error) {
	return p.CallRunWithArgs(func(concrete api.API, args []interface{}) ([]interface{}, error) {
		size := int(args[0].(*big.Int).Int64())
		hash := common.Hash(args[1].([32]byte))
		store := api.NewPersistentBigPreimageStore(concrete, -1, -1)
		if !store.HasPreimage(hash) {
			return []interface{}{[]byte{}}, errors.New("preimage not found")
		}
		realSize := store.GetPreimageSize(hash)
		if size < realSize {
			return []interface{}{[]byte{}}, errors.New("provided preimage size too small")
		}
		preimage := store.GetPreimage(hash)
		return []interface{}{preimage}, nil
	}, concrete, input)
}
