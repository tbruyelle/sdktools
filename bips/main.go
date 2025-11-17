package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/cosmos/go-bip39"
	"github.com/gofika/bip32"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

var hdPaths = map[string]string{
	"cosmos": "m/44'/118'/0'/0/0",
	"btc":    "m/84'/0'/0'/0/0",
}

func main() {
	prefix := flag.String("prefix", "atone", "prefix of the address")
	hdpath := flag.String("hdpath", "cosmos", "one of "+strings.Join(slices.Sorted(maps.Keys(hdPaths)), ", "))
	flag.Parse()
	if _, ok := hdPaths[*hdpath]; !ok {
		panic(fmt.Errorf("%s is not a valid hd path", *hdpath))
	}
	var mnemonic, passphrase string
	// mnemonic = "burden junk salon cabbage energy damp view camp pole endorse isolate arrange struggle reflect easy hawk chat social finish prepare wagon utility drive input"
	// atone1rku58s0axgpex6e2uuarxpcrzu3gyur2wkhyqd
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bz, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		// Looks like its impossible to read stdin from pipe and using
		// term.GetPassword, so for the passphrase we need to add it after the
		// mnemonic.
		// Ex: go run . <(echo "word1 word2 ..."; read -p "Passphrase:" -s pass; echo $pass)
		t := bytes.Split(bytes.TrimSpace(bz), []byte("\n"))
		mnemonic = string(t[0])
		if len(t) > 1 {
			passphrase = string(t[1])
		}
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
	seed := bip39.NewSeed(mnemonic, passphrase)

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
	derivedPriv, err := bip32.DerivePath(privkey, hdPaths[*hdpath])
	if err != nil {
		panic(err)
	}
	privKey := secp256k1.PrivKey{Key: derivedPriv.ECPrivateKeyBytes()}
	bech := types.MustBech32ifyAddressBytes(*prefix, privKey.PubKey().Address())
	fmt.Println("bech: ", bech)
}
