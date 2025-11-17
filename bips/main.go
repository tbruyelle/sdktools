package main

import (
	"fmt"

	"github.com/cosmos/go-bip39"
	dbip32 "github.com/gofika/bip32"
	"github.com/tyler-smith/go-bip32"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {
	ent, err := bip39.NewEntropy(256)
	if err != nil {
		panic(err)
	}
	mnemonic, err := bip39.NewMnemonic(ent)
	if err != nil {
		panic(err)
	}
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		panic(err)
	}
	publicKey := masterKey.PublicKey()

	// Display mnemonic and keys
	fmt.Println("Mnemonic: ", mnemonic)
	fmt.Println("Master private key: ", masterKey)
	fmt.Println("Master public key: ", publicKey)

	dkey, err := dbip32.NewExtendedKey(seed)
	if err != nil {
		panic(err)
	}
	b58 := bip32.BitcoinBase58Encoding.EncodeToString(dkey.ECPrivateKeyBytes())
	fmt.Println("dMaster private key: ", b58)
	kkey := bip32.BitcoinBase58Encoding.EncodeToString(masterKey.Key)
	fmt.Println("xxxxxxx private key: ", kkey)

	// Derivation
	atomHDPath := "m/44'/118'/0'/0/0"
	dkey2, err := dbip32.DerivePath(dkey, atomHDPath)
	if err != nil {
		panic(err)
	}
	privKey := secp256k1.PrivKey{Key: dkey2.ECPrivateKeyBytes()}
	bech := types.MustBech32ifyAddressBytes("atone", privKey.PubKey().Address())
	fmt.Println("bech: ", bech)
}
