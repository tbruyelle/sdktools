package main

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	var i int
	for {
		i++
		pv := secp256k1.GenPrivKey()
		addr := pv.PubKey().Address()
		res, err := types.Bech32ifyAddressBytes("atone", addr)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
		if strings.HasPrefix(res, "atone1t0mt0m") {
			fmt.Println("FOUND AFTER ATTEMPTS", i)
			break
		}
	}
}
