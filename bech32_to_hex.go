package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func main() {
	_, bz, err := bech32.DecodeAndConvert(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bz))
}
