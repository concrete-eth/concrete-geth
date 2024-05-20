package rpc

import (
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/rpc"
)

type APIConstructor func(*eth.Ethereum) rpc.API
