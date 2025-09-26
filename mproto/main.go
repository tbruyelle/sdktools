package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/types"
)

func main() {
	var (
		chainID          = "atomone-1"
		heights          = []int64{5, 3}
		round      int64 = 0
		timestamp        = toTime("2025-09-25T07:55:57.306746166Z")
		blockhash        = b64Dec("NpiImIJoaSaIucwNs5cqpgMsL/8wxEPYC3P0jA5aQSI=")
		parsethash       = b64Dec("QqzwnLzvixIcUz+hPeUQjDV6NaLkFRKXACCxJIrBHzw=")
		privkeyStr       = flag.String("privkey", "", "base64 encoded private key")
	)
	flag.Parse()
	privk := ed25519.GenPrivKey()
	if *privkeyStr != "" {
		privk = ed25519.PrivKey(b64Dec(*privkeyStr))
	}
	fmt.Println("ADDR:", base64.StdEncoding.EncodeToString(privk.PubKey().Address()))
	fmt.Println("PUBK:", base64.StdEncoding.EncodeToString(privk.PubKey().Bytes()))
	fmt.Println("PRIV:", base64.StdEncoding.EncodeToString(privk.Bytes()))
	for _, height := range heights {
		vote := cmtproto.CanonicalVote{
			Type:   types.PrecommitType,
			Height: height,
			Round:  round,
			BlockID: &cmtproto.CanonicalBlockID{
				Hash: blockhash,
				PartSetHeader: cmtproto.CanonicalPartSetHeader{
					Total: 1,
					Hash:  parsethash,
				},
			},
			Timestamp: timestamp,
			ChainID:   chainID,
		}
		bz, err := protoio.MarshalDelimited(&vote)
		if err != nil {
			panic(err)
		}

		signature, err := privk.Sign(bz)
		if err != nil {
			panic(err)
		}

		fmt.Printf("SIGN h=%d: %s\n", height, base64.StdEncoding.EncodeToString(signature))
	}
}

func b64Dec(s string) []byte {
	bz, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}

func toTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}
