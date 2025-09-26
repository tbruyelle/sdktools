package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/types"
)

func TestBeezeeUpdateClient(t *testing.T) {
	bz := []byte(`{
		"height": 19282029,
			        "round": 0,
			        "block_id": {
			          "hash": "abcd8f05a2e7e15837a0b9b4bfaef404e9b5032f8247af43c81a8d03f57ffe5e",
			          "parts": {
			            "total": 1,
			            "hash": "3ea06a2c509be988424191439ce787ce9d641c06cab3b7f01ea3b09e7f216b6b"
			          }
			        },
			        "signatures": [
			          {
			            "block_id_flag": 1,
			            "validator_address": "",
			            "timestamp": "0001-01-01T00:00:00Z",
			            "signature": ""
			          },
			          {
			            "block_id_flag": 2,
			            "validator_address": "9982e3205f40c7047e0419640fd3377fec6a28d1",
			            "timestamp": "2025-09-25T14:00:49.373544902Z",
			            "signature": "rgUFLD3Yk9JdcfWSqcscXeUZVY+Q7E1jQ1Eumkh/LT2fUI4dsk2NBeOhjyKaF5Ey/q6DwoINss6iSxNPj8s2Ag=="
			          },
			          {
			            "block_id_flag": 2,
			            "validator_address": "cac6c644778e878843f5f51b90cd511cff5b55f7",
			            "timestamp": "2025-09-25T14:00:49.396076476Z",
			            "signature": "thHpLNT6555JHoNr10iBpp8NJemRMPKMwnV4QDL6gLpaYChqrQND+8K4hQrC8Ld2IAuIT92ZHFWOmoVH9bIcBw=="
			          }
			        ]
			      }`)

	var commit types.Commit
	err := json.Unmarshal(bz, &commit)
	require.NoError(t, err)
	// spew.Config.DisableMethods = true
	// spew.Dump(commit)

	var pubkeys []ed25519.PubKey
	bz, _ = base64.StdEncoding.DecodeString("wF3Daj7tqX8nVOJd2WCkMVvqyx/lPzuI5/y3wwSRwbY=")
	pubkeys = append(pubkeys, ed25519.PubKey(bz))
	bz, _ = base64.StdEncoding.DecodeString("dzCYrLu2sjpWSiEd2MVqsr+Q6ocBDUUUrRKBehGOeLM=")
	pubkeys = append(pubkeys, ed25519.PubKey(bz))
	bz, _ = base64.StdEncoding.DecodeString("tYyou2DP3JWDrvKPtbAZpTCEtzqr9h1nsi/srNZwLiY=")
	pubkeys = append(pubkeys, ed25519.PubKey(bz))
	chainid := "beezee-1"

	for idx, pubkey := range pubkeys {
		signBytes := commit.VoteSignBytes(chainid, int32(idx))

		bz = VoteBytesToSign(chainid, commit.Height, int64(commit.Round), commit.BlockID.Hash,
			commit.BlockID.PartSetHeader.Total, commit.BlockID.PartSetHeader.Hash,
			commit.Signatures[idx].BlockIDFlag, commit.Signatures[idx].Timestamp)
		fmt.Printf("> %x\n", signBytes)
		fmt.Printf("X %x\n", bz)
		require.Equal(t, signBytes, bz, "bytes to sign mismatch")

		println("SIGN", idx,
			pubkey.VerifySignature(signBytes, commit.Signatures[idx].Signature),
		)
	}
}

func TestOsmosisUpdateClient(t *testing.T) {
	// Fetched from https://www.mintscan.io/atomone/tx/E7A075A6B6BC56ED2006A91F7833354C7F721CC9618BF09F276B4F58B119149B?sector=json
	bz := []byte(`{
  "height": 44873552,
  "round": 0,
  "block_id": {
    "hash": "369888988268692688b9cc0db3972aa6032c2fff30c443d80b73f48c0e5a4122",
    "parts": {
      "total": 1,
      "hash": "42acf09cbcef8b121c533fa13de5108c357a35a2e41512970020b1248ac11f3c"
    }
  },
  "signatures": [
    {
      "block_id_flag": 2,
      "validator_address": "cb5a63b91e8f4ee8db935942cbe25724636479e0",
      "timestamp": "2025-09-25T07:55:57.306746166Z",
      "signature": "qtv1z4S2Q6T87vGQo0lrjRZqv9PrHIji4pTyviMnVyGx9td6eySdzwQwCthwmihU48ebNlFiMlFJ0CT891UmDg=="
    },
    {
      "block_id_flag": 2,
      "validator_address": "66b69666ebf776e7ebcbe197aba466a712e27076",
      "timestamp": "2025-09-25T07:55:57.310583641Z",
      "signature": "Q5E6Kjma00n/T98rC9qJmoB6JTGFX/IB+mDVs4Wd1h0eJ8fabY/6oI8zdoU6/7W6VR6wjpHyWBsJrpGT6C0LCg=="
    },
    {
      "block_id_flag": 1,
      "validator_address": "",
      "timestamp": "0001-01-01T00:00:00Z",
      "signature": ""
		}
  ]
}`)
	var commit types.Commit
	err := json.Unmarshal(bz, &commit)
	require.NoError(t, err)
	// spew.Config.DisableMethods = true
	// spew.Dump(commit)

	var pubkeys []ed25519.PubKey
	bz, _ = base64.StdEncoding.DecodeString("6Nz09YGHzwWxjczG0IhK4Iv0qY2IcX0P/5KitvRXTUc=")
	pubkeys = append(pubkeys, ed25519.PubKey(bz))
	bz, _ = hex.DecodeString("c01db94ad2f16f3983d2e4e21621fac724997741f5de4c9a9cd52fbe55296b7e")
	pubkeys = append(pubkeys, ed25519.PubKey(bz))
	chainid := "osmosis-1"
	for idx, pubkey := range pubkeys {
		signBytes := commit.VoteSignBytes(chainid, int32(idx))

		bz = VoteBytesToSign(chainid, commit.Height, int64(commit.Round), commit.BlockID.Hash,
			commit.BlockID.PartSetHeader.Total, commit.BlockID.PartSetHeader.Hash,
			commit.Signatures[idx].BlockIDFlag, commit.Signatures[idx].Timestamp,
		)
		fmt.Printf("> %x\n", signBytes)
		fmt.Printf("X %x\n", bz)
		require.Equal(t, signBytes, bz, "bytes to sign mismatch")

		println("SIGN", idx,
			pubkey.VerifySignature(signBytes, commit.Signatures[idx].Signature),
		)
	}
}
