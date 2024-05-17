package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/cmd/geth"
	"github.com/ethereum/go-ethereum/concrete"
)

func main() {
	app := geth.NewConcreteGethApp(&concrete.GenericPrecompileRegistry{})
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
