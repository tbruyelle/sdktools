package main

import (
	"fmt"
	"strings"

	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	hdPath := hd.CreateHDPath(types.CoinType, 0, 0).String()
	for i := 0; ; i++ {
		// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
		entropySeed, err := bip39.NewEntropy(256)
		if err != nil {
			panic(err)
		}
		mnemonic, err := bip39.NewMnemonic(entropySeed)
		if err != nil {
			panic(err)
		}

		derivPriv, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath)
		if err != nil {
			panic(err)
		}
		pv := hd.Secp256k1.Generate()(derivPriv)
		addr := pv.PubKey().Address()
		res, err := types.Bech32ifyAddressBytes("atone", addr)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
		if strings.HasPrefix(res, "atone1t0m") {
			fmt.Println("FOUND AFTER ATTEMPTS", i)
			fmt.Println(mnemonic)
			break
		}
	}
}
