package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cosmos/go-bip39"
	"github.com/gofika/bip32"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	var mnemonic string
	// mnemonic = "burden junk salon cabbage energy damp view camp pole endorse isolate arrange struggle reflect easy hawk chat social finish prepare wagon utility drive input"
	// atone1rku58s0axgpex6e2uuarxpcrzu3gyur2wkhyqd
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Println("Reading mnemonic from stdin...")
		bz, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		mnemonic = strings.TrimSpace(string(bz))
	} else {
		ent, err := bip39.NewEntropy(256)
		if err != nil {
			panic(err)
		}
		mnemonic, err = bip39.NewMnemonic(ent)
		if err != nil {
			panic(err)
		}
		fmt.Println("Generated mnemonic:", mnemonic)
	}
	seed := bip39.NewSeed(mnemonic, "")

	// Following comments use	"github.com/tyler-smith/go-bip32"
	// It prints the addresses with the xpriv/xpub prefix
	// masterKey, _ := bip32.NewMasterKey(seed)
	// publicKey := masterKey.PublicKey()
	// fmt.Println("Master private key:", masterKey)
	// fmt.Println("Master public key:", publicKey)

	privkey, err := bip32.NewExtendedKey(seed)
	if err != nil {
		panic(err)
	}

	// Derivation
	atomHDPath := "m/44'/118'/0'/0/0"
	derivedPriv, err := bip32.DerivePath(privkey, atomHDPath)
	if err != nil {
		panic(err)
	}
	privKey := secp256k1.PrivKey{Key: derivedPriv.ECPrivateKeyBytes()}
	bech := types.MustBech32ifyAddressBytes("atone", privKey.PubKey().Address())
	fmt.Println("bech: ", bech)
}
