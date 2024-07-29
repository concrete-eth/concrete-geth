package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/cmd/geth"
	"github.com/ethereum/go-ethereum/concrete"
	concrete_rpc "github.com/ethereum/go-ethereum/concrete/rpc"
)

func main() {
	app := geth.NewConcreteGethApp(&concrete.GenericPrecompileRegistry{}, []concrete_rpc.APIConstructor{})
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
