package main

import (
	"encoding/hex"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func main() {
	bz, err := hex.DecodeString(os.Args[1])
	if err != nil {
		panic(err)
	}
	addr := sdk.AccAddress(bz)
	s, err := bech32.ConvertAndEncode("atone", addr)
	fmt.Println(s, err)
}
