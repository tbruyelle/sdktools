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

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

var hdPaths = map[string][]uint32{
	"cosmos": { // m/44'/118'/0'/0/0
		44 + hdkeychain.HardenedKeyStart,
		118 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart, 0, 0,
	},
	"segwit": { // "m/84'/0'/0'/0/0",
		84 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart, 0, 0,
	},
}

func main() {
	prefix := flag.String("prefix", "atone", "prefix of the address")
	hdpath := flag.String("hdpath", "cosmos", "one of "+strings.Join(slices.Sorted(maps.Keys(hdPaths)), ", "))
	flag.Parse()
	if _, ok := hdPaths[*hdpath]; !ok {
		panic(fmt.Errorf("%s is not a valid hd path", *hdpath))
	}
	var mnemonic, passphrase string
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
	fmt.Println("bech: ", deriveBech32(mnemonic, passphrase, hdPaths[*hdpath], *prefix, nil))
	fmt.Println("btc: ", deriveBtc(mnemonic, passphrase))
}

func deriveBech32(mnemonic, passphrase string, hdpath []uint32, prefix string, witnessVersion []byte) string {
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Derive the master private key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}

	// Derive the first child private key
	currentKey := masterKey
	for _, index := range hdpath {
		currentKey, err = currentKey.Derive(index)
		if err != nil {
			panic(err)
		}
	}
	// Following comments use	"github.com/tyler-smith/go-bip32"
	// It prints the addresses with the xpriv/xpub prefix
	// masterKey, _ := hbip32.NewMasterKey(seed)
	// publicKey := masterKey.PublicKey()
	// fmt.Println("Master private key:", masterKey)
	// fmt.Println("Master public key:", publicKey)

	// Get the private key
	privKey, err := currentKey.ECPrivKey()
	if err != nil {
		panic(err)
	}
	// Get the public key
	pubKey := privKey.PubKey()
	witnessProg := btcutil.Hash160(pubKey.SerializeCompressed())
	bz, err := bech32.ConvertBits(witnessProg, 8, 5, true)
	if err != nil {
		panic(err)
	}
	bz = append(witnessVersion, bz...)
	addr, err := bech32.Encode(prefix, bz)
	if err != nil {
		panic(err)
	}
	return addr
}

func deriveBtc(mnemonic, passphrase string) string {
	// Convert mnemonic to seed
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Derive the master private key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}

	// Derive the first child private key (m/84'/0'/0'/0/0 for P2WPKH)
	path := []uint32{84 + hdkeychain.HardenedKeyStart, 0 + hdkeychain.HardenedKeyStart, 0 + hdkeychain.HardenedKeyStart, 0, 0}
	currentKey := masterKey
	for _, index := range path {
		currentKey, err = currentKey.Derive(index)
		if err != nil {
			panic(err)
		}
	}

	// Get the private key
	privKey, err := currentKey.ECPrivKey()
	if err != nil {
		panic(err)
	}
	// Get the public key
	pubKey := privKey.PubKey()
	witnessProg := btcutil.Hash160(pubKey.SerializeCompressed())
	address, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	// Encode the address in Bech32 format
	return address.EncodeAddress()
}
