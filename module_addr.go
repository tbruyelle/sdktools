package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func main() {
	prefix := os.Args[1]
	moduleName := os.Args[2]
	addr := address.Module(moduleName)
	s, err := bech32.ConvertAndEncode(prefix, addr)
	fmt.Println(s, err)
}
