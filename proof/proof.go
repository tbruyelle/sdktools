package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"

	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"

	abci "github.com/cometbft/cometbft/abci/types"
)

func main() {
	verifyPacketReceipt()
	verifyGovParams()
}

func verifyGovParams() {
	var (
		// Proof of gov params key existence (path=store/gov/key, data=0x30)
		// using https://atomone-rpc.allinbits.services/abci_query?path=%22store/gov/key%22&data=0x30&prove=true&height=3272353
		atomOneTmResponseBz = []byte(`{
      "code": 0,
      "log": "",
      "info": "",
      "index": "0",
      "key": "MA==",
      "value": "ChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBQ==",
      "proofOps": {
        "ops": [
          {
            "type": "ics23:iavl",
            "key": "MA==",
            "data": "CvYFCgEwEuUBChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBRoLCAEYASABKgMAAgIiKwgBEgQCBAIgGiEgjkBfizMZnHPckgMhtvluJjWvXT3288+1zy6yhhSxJ/QiLAgBEigEBo7UgAMgC3FGm4XkJ940SMUTV88+GaktknbWwZELhuQitp0ostggIiwIARIoCBC0sIgDIObhZQlpgV7lAAkE4EVPzyRQ78D/Wga7zuRVOCLn8bPCICIsCAESKAom/v+LAyAK8XslV1h+x1gaPdiokPHEKqXJk0QYHXC7Nuy6dqArrSAiLAgBEigMQP7/iwMg717IfgAGlJB7r0r1V68XEUp5QWySWU6oc8NpVuyYxPkgIiwIARIoDnbQiI4DIIWS/4XmMV9e68ua2NehH9M9HE6LjJo6G/yiaSjT1IFuICItCAESKRDmAciwjwMgcRraZK72gzbJGYLly8sjVXSs4egx9WaFUg1qdK61JFwgIi0IARIpEs4DyLCPAyDkyI36z/ISSHqurALs9P6hCzwsbgbgAqBgN8PrXddrOCAiLQgBEikWzAjIsI8DIM7t/nSZbcGRfqp1YGWz2gkYz/QkdnV16GmURlfs5r8aICItCAESKRigEMiwjwMg41kGMrq3iF3B97BsnT3EQCeNZN5q8tOAYUf9o4gp23wgIi0IARIpGvwcyrmPAyBO5zfmU5nTCADyowyL/L7ncQT+smydLrSU/tt9NRayUyA="
          },
          {
            "type": "ics23:simple",
            "key": "Z292",
            "data": "CvsBCgNnb3YSIFLzcE9KBJLG9WckEb6FV9uFwR4auXG9zWG66V37KlkOGgkIARgBIAEqAQAiJQgBEiEBEaKNvFesWykCp7xJPv9ZB24/kB+z/T6qrqUsnrdZCikiJwgBEgEBGiC726NGVm34Tj9qA6XgZGbPDl4QS8oj24ixO0ddb6noziInCAESAQEaIIInMDlyhGeNFqLr7Ptt/V+Sx/ajq/mQrGL6TyrxAiGbIiUIARIhAU+MFmaiO6av3UlYtHRbdkSfpn+uY5Z1xbK/cq4yIXoMIicIARIBARogtQM5fvi3bFgNEh7PXrhks9meSOTLk8yBS9YeUfdCXQ4="
          }
        ]
      },
      "height": "3272353",
      "codespace": ""
    }`)
		// app hash for this proof
		// NOTE must be taken from the block after
		// atomoned q block 3272354|jq '.block.header.app_hash'
		// https://atomone-rpc.allinbits.services/block?height=3272354
		atomOneAppHash = "B2F11D67EE8D305A15234F3927D14074F8377B6AE1A2CD570E9F24BA50E0F7A4"
	)

	var res abci.ResponseQuery
	err := json.Unmarshal(atomOneTmResponseBz, &res)
	if err != nil {
		panic(err)
	}

	// Turn AtomOne tm proof into ics23 commitment proof used by tm light client
	proofs := make([]*ics23.CommitmentProof, len(res.ProofOps.Ops))
	for i, op := range res.ProofOps.Ops {
		var p ics23.CommitmentProof
		err = p.Unmarshal(op.Data)
		if err != nil || p.Proof == nil {
			panic(fmt.Sprintf("could not unmarshal proof op into CommitmentProof at index %d: %v", i, err))
		}
		proofs[i] = &p
	}
	merkleProof := commitmenttypes.MerkleProof{Proofs: proofs}

	// Verify proofs against app hash
	appHashBz, err := hex.DecodeString(atomOneAppHash)
	if err != nil {
		panic(err)
	}
	merkleRoot := commitmenttypes.NewMerkleRoot(appHashBz)
	specs := commitmenttypes.GetSDKSpecs()
	path := commitmenttypes.NewMerklePath([]byte("gov"), []byte{0x30})

	// FIXME invalid proof for now.
	// TODO try with a real packet committment
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, res.Value)
	fmt.Println("VERIFY GOV PARAMS", err)
}

func verifyPacketReceipt() {
	var (
		// packet receipt proof
		// cmd: atomoned q ibc channel packet-receipt transfer channel-2 5
		// (executed at block 3284732)
		packetReceiptProof = "Cv8GCvwGCjZyZWNlaXB0cy9wb3J0cy90cmFuc2Zlci9jaGFubmVscy9jaGFubmVsLTIvc2VxdWVuY2VzLzUSAQEaDggBGAEgASoGAALivqQBIi4IARIHAgTuoa0BIBohIB75sZwfkDzlhtkMWvBlVO/09xt32XuLSW3Gs1EFcVxcIi4IARIHBAjO3J0CIBohIKtT4vnhPqM+zHaDBUUxF++hHTcDHQaQm/72uRRts6C0Ii4IARIHBgzc3p0CIBohIBTxybwXEerVu2FhcmUF9dgR4sMOGcjpJB/7Oliar9ZqIi4IARIHCBr67Z0CIBohICCa73j418IwV7L+v3hjXt0h7gLacbhWb7gjT1o2CuRZIi4IARIHCjSYjp8CIBohIKMmb6s/bIrS2gTtk5GjmdR1vPWfsTFwYDYmwv2e84WuIi4IARIHDFy6jp8CIBohIMbHXPqlz/TIdZMLmFr2DCtG28WmxgkZAMmHWJ19Z6IHIi8IARIIDqQBuMyiAiAaISBMu69rcTqkOvH5yeH4heNGItguzVTcZYOEdhbaL4cF7yItCAESKRDUAsqhpAIg0r0Ym25yQ1YL7iXudXdE3GK12Z43Ho+UPAa/4io9k30gIi8IARIIEu4ElralAiAaISDce6GSOlj0zDVHnelIhXKONyawrvHBieZffPnlKeHNVSItCAESKRSmB5a2pQIgRp+zjKZ00W/NRt482naJ/P8n4+YQIKggq59HDtwC08ogIi8IARIIFqQPtMHCAiAaISDBUtua6oAfgs2xzL7M30XTxvFbqaKUO5hZGlE02HEsvyIvCAESCBiuGuKrhAMgGiEgemk0FvCYc15vtMpv29pJdIBwusmoRIFoA0TcUECMfCkiLwgBEgga1jPQ+5ADIBohIFiW4/yq3zE3g21/XQ5ATyav6joafuzKXTWkI+fKUHb7Ii0IARIpHIJK0PuQAyCfIHtkhwS2N3YElYPTj9lHSaTAodxrcWvfFfwzqK0BvyAiLQgBEikevHfQ+5ADIMEq49iU7lM1954NHBdwNvbYvYRga4VfG6h3Jus6OEKPICIuCAESKiDkpwLQ+5ADIBLQKsLEG+DPtcpquKcTwx1Mg47rfWS+vdqcEnB+UOVKICIuCAESKiSUzQfQ+5ADIN9RVUWddGT8w4hlUH5I+OIRXsOUapnfjMXYGwTm7tGNIAr+AQr7AQoDaWJjEiCi66yl8Vs8oNZp6bo7/sSzMynPtI9Iy233W5ki83oiUBoJCAEYASABKgEAIicIARIBARog71t4+pmoAEEVea1maHPr5hzLlyD/cxsk+U8+SsQdawAiJQgBEiEBbs7Zi9r/+uoeQizLo0dXgZnAM//N0yEiKTWldwb4h3EiJwgBEgEBGiAdZBXWzrgJl4GIz8kqNXnQeb2OUdLb//VrLWOeQuUTTyIlCAESIQEEnW5ItZcknp8R3x4+VdTI73Ixm2F9DIyi6vpYnLPj+SInCAESAQEaIPffzxYCAbdKN6vwBzoCzWg1t5X1Gp+5TFmnED+hRhfe"
		// cmd: atomoned q block 3284732|jq '.block.header.app_hash'
		packetReceiptAppHash = "416D75EE392246A41BEA5FBD350C13A5EA54DD9F57F75DB5991C3EB3D4BBACF0"
	)

	var merkleProof commitmenttypes.MerkleProof
	bz, err := base64.StdEncoding.DecodeString(packetReceiptProof)
	if err != nil {
		panic(err)
	}
	err = merkleProof.Unmarshal(bz)
	if err != nil {
		panic(err)
	}
	appHashBz, err := hex.DecodeString(packetReceiptAppHash)
	if err != nil {
		panic(err)
	}
	merkleRoot := commitmenttypes.NewMerkleRoot(appHashBz)
	specs := commitmenttypes.GetSDKSpecs()
	key := host.PacketReceiptKey("transfer", "channel-2", 5)
	path := commitmenttypes.NewMerklePath([]byte("ibc"), key)
	value := []byte{byte(1)} // value of packet receipt is always 1

	// FIXME invalid proof for now.
	// TODO try with a real packet committment
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, value)
	fmt.Println("VERIFY PACKET RECEIPT", err)
}
