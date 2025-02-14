package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/cosmos/go-bip39"

	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	cpus := runtime.NumCPU()
	fmt.Println("CPUs=", cpus)
	hdPath := hd.CreateHDPath(types.CoinType, 0, 0).String()
	// add an errgroup to this main function
	g, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < cpus; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
					entropySeed, err := bip39.NewEntropy(256)
					if err != nil {
						return err
					}
					mnemonic, err := bip39.NewMnemonic(entropySeed)
					if err != nil {
						return err
					}

					derivPriv, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath)
					if err != nil {
						return err
					}
					pv := hd.Secp256k1.Generate()(derivPriv)
					addr := pv.PubKey().Address()
					res, err := types.Bech32ifyAddressBytes("atone", addr)
					if err != nil {
						return err
					}
					// fmt.Println(res)
					if strings.HasPrefix(res, "atone1t0m") && strings.HasSuffix(res, "t0m") {
						fmt.Println(res)
						fmt.Println(mnemonic)
						return errors.New("OK")
					}
				}
			}
		})
	}
	fmt.Println("DONE:", g.Wait())
}
