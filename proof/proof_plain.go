//go:build plain
package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/davecgh/go-spew/spew"

	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"

	abci "github.com/cometbft/cometbft/abci/types"

	gnoabci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	gnomerkle "github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	gnoiavl "github.com/gnolang/gno/tm2/pkg/iavl"
	gnorootmulti "github.com/gnolang/gno/tm2/pkg/store/rootmulti"
)

func main() {
	spew.Config.DisableMethods = true

	verifyGnoGasPrice()
	verifyA1PacketReceipt()
	verifyA1GovParams()
	// TODO check A1/Gno non existence
	verifyGnoAbsence()

	// TODO find how to determine key to r/ibc packet commitment/ack
	// objectID: oid:<OBJECT_ID> where OBJECT_ID is REALM_ID:sequence
	// is it possible to fetch that from a gno sc and returns it to the
	// caller ?
}

func verifyGnoGasPrice() {
	var (
		// tm2 abci query with proof
		// from https://rpc.gno.land/abci_query?path=.store/main/key&data=Z2FzUHJpY2U=&prove=true
		// {
		//   "jsonrpc": "2.0",
		//   "id": "",
		//   "result": {
		//     "response": {
		//       "ResponseBase": {
		//         "Error": null,
		//         "Data": null,
		//         "Events": null,
		//         "Log": "",
		//         "Info": ""
		//       },
		//       "Key": "Z2FzUHJpY2U=",
		//       "Value": "CNAPEgYxdWdub3Q=",
		//       "Proof": {
		//         "ops": [
		//           {
		//             "type": "iavl:v",
		//             "key": "Z2FzUHJpY2U=",
		//             "data": "pwUKpAUKLAgeELbdARiM70IiINMgtJjpgT/AGM88doO3aSbS6ah5l7vjlp552hTOPb7nCisIHBC+XBiM70IiIDiovZ2sL7oAt42DrhDGHa/3aE/gEZgR4Xr2FXSCA0WnCisIGhCePBiM70IiIKkrz/9nfsEum4Dxo6NDosk6NqljFQJCldXDSFK5QliXCisIGBCKHBiM70IqIFtng0BYYCXPEiX42jodxi2rbco3wyZTJwIsmbEbZS0UCisIFhCiEBiM70IiIFG6UXzWkY1wHX8RzmYzRbYf0zp8v3ltKYtv5s031by1CisIFBCKCBiM70IiIHfzys8A4eP67H0WmmcCJ3uXdCdhVslydAVnejEv2x+gCisIEBCABBiM70IiIJzzUhruzOjIGxqy3XHv6szhYnmpqoWncoaq04HSuRdRCisIDhCAAhiM70IqIPR/pcvgNdgUICxaC5+EfSWajv3PuNAveq78jgolqdWTCisIDBCAARiM70IqIHSIu9wK0pjwS6BJC2LXNU+OBruzNQ4AhFxxmZcXno1MCioIChBAGIzvQiogkBOFQ4fze4rfpyRjQkC6yY975kUeP7QSjd0XHkO7x4wKKggIECAYjO9CKiD+MHdW+K+TJIi5f3Z0HKKb2txg0FiOMhDJ57rGpPH+rAoqCAYQEBiM70IiIJY5g9TlmAlMFUhJweBxm7AMVjEMiRa5kd5AYwqUEM/vCioIBBAIGIzvQiog9goMmM0VmW/1MDIB6pMB/rBiLRdrhF5l1g8a8ZswRK0KKggCEAQYjO9CIiBbt5DisYgvM78Isi7Xg/zg2n+5r0X8wNZa10TAzNZcsRowCghnYXNQcmljZRIgoQVt4nMhH3LZH6D1JMQXjvzN5DX++k+Wa6m4XcEwejEYjO9C"
		//           },
		//           {
		//             "type": "multistore",
		//             "key": "bWFpbg==",
		//             "data": "PAo6CjAKBG1haW4SKAomCIzvQhIgBS5dNZtUDsSlUZCBGLVo11+XSjXWjM37z2IrtAUsrhgKBgoEYmFzZQ=="
		//           }
		//         ]
		//       },
		//       "Height": "547782"
		//     }
		//   }
		// }
		// NOTE: for some reason, the Height field had to be manually changed into
		// a number
		abciResponseQueryBz = []byte(`{
      "ResponseBase": {
        "Error": null,
        "Data": null,
        "Events": null,
        "Log": "",
        "Info": ""
      },
      "Key": "Z2FzUHJpY2U=",
      "Value": "CNAPEgYxdWdub3Q=",
      "Proof": {
        "ops": [
          {
            "type": "iavl:v",
            "key": "Z2FzUHJpY2U=",
            "data": "pwUKpAUKLAgeELbdARiM70IiINMgtJjpgT/AGM88doO3aSbS6ah5l7vjlp552hTOPb7nCisIHBC+XBiM70IiIDiovZ2sL7oAt42DrhDGHa/3aE/gEZgR4Xr2FXSCA0WnCisIGhCePBiM70IiIKkrz/9nfsEum4Dxo6NDosk6NqljFQJCldXDSFK5QliXCisIGBCKHBiM70IqIFtng0BYYCXPEiX42jodxi2rbco3wyZTJwIsmbEbZS0UCisIFhCiEBiM70IiIFG6UXzWkY1wHX8RzmYzRbYf0zp8v3ltKYtv5s031by1CisIFBCKCBiM70IiIHfzys8A4eP67H0WmmcCJ3uXdCdhVslydAVnejEv2x+gCisIEBCABBiM70IiIJzzUhruzOjIGxqy3XHv6szhYnmpqoWncoaq04HSuRdRCisIDhCAAhiM70IqIPR/pcvgNdgUICxaC5+EfSWajv3PuNAveq78jgolqdWTCisIDBCAARiM70IqIHSIu9wK0pjwS6BJC2LXNU+OBruzNQ4AhFxxmZcXno1MCioIChBAGIzvQiogkBOFQ4fze4rfpyRjQkC6yY975kUeP7QSjd0XHkO7x4wKKggIECAYjO9CKiD+MHdW+K+TJIi5f3Z0HKKb2txg0FiOMhDJ57rGpPH+rAoqCAYQEBiM70IiIJY5g9TlmAlMFUhJweBxm7AMVjEMiRa5kd5AYwqUEM/vCioIBBAIGIzvQiog9goMmM0VmW/1MDIB6pMB/rBiLRdrhF5l1g8a8ZswRK0KKggCEAQYjO9CIiBbt5DisYgvM78Isi7Xg/zg2n+5r0X8wNZa10TAzNZcsRowCghnYXNQcmljZRIgoQVt4nMhH3LZH6D1JMQXjvzN5DX++k+Wa6m4XcEwejEYjO9C"
          },
          {
            "type": "multistore",
            "key": "bWFpbg==",
            "data": "PAo6CjAKBG1haW4SKAomCIzvQhIgBS5dNZtUDsSlUZCBGLVo11+XSjXWjM37z2IrtAUsrhgKBgoEYmFzZQ=="
          }
        ]
      },
      "Height": 547782
    }`)
		// app hash from https://rpc.gno.land/abci_info
		// {
		//  "jsonrpc": "2.0",
		//  "id": "",
		//  "result": {
		//    "response": {
		//      "ResponseBase": {
		//        "Error": null,
		//        "Data": "Z25vbGFuZA==",
		//        "Events": null,
		//        "Log": "",
		//        "Info": ""
		//      },
		//      "ABCIVersion": "",
		//      "AppVersion": "",
		//      "LastBlockHeight": "547782",
		//      "LastBlockAppHash": "0P9gq1X8hqEYS7xglsYwzW2WUcbCtBKyRoo6xQWO48A="
		//    }
		//  }
		// }
		appHash = "0P9gq1X8hqEYS7xglsYwzW2WUcbCtBKyRoo6xQWO48A="
	)
	var res gnoabci.ResponseQuery
	err := json.Unmarshal(abciResponseQueryBz, &res)
	if err != nil {
		panic(err)
	}
	prf := gnorootmulti.DefaultProofRuntime()
	proofOps := make(gnomerkle.ProofOperators, len(res.Proof.Ops))
	for i, op := range res.Proof.Ops {
		po, err := prf.Decode(op)
		if err != nil {
			panic(err)
		}
		proofOps[i] = po
	}

	// Verify proofs against app hash
	appHashBz, err := base64.StdEncoding.DecodeString(appHash)
	if err != nil {
		panic(err)
	}

	err = proofOps.VerifyValue(appHashBz, "/main/gasPrice", res.Value)
	fmt.Println("VERIFY GNO GAS PRICE", err)
	return // TODO remove me when able to transform proofOps into ics23 format

	// TODO Turn gno proof into ics23 commitment proof so it can be used by the
	// default 07-tendermint light client implementation
	tmProofs := make([]*ics23.CommitmentProof, len(proofOps))
	for i, p := range proofOps {
		pp := p.(gnoiavl.IAVLValueOp)
		tmProofs[i] = &ics23.CommitmentProof{
			Proof: &ics23.CommitmentProof_Exist{
				Exist: &ics23.ExistenceProof{
					Key:   pp.GetKey(),
					Value: res.Value,
					Leaf: &ics23.LeafOp{
						Hash:         ics23.HashOp_SHA256,
						PrehashKey:   ics23.HashOp_NO_HASH,
						PrehashValue: ics23.HashOp_SHA256,
						Length:       ics23.LengthOp_VAR_PROTO,
						Prefix:       nil,
					},
					// ...?
				},
			},
		}
	}
	merkleProof := commitmenttypes.MerkleProof{Proofs: tmProofs}
	merkleRoot := commitmenttypes.NewMerkleRoot(appHashBz)
	specs := commitmenttypes.GetSDKSpecs()
	path := commitmenttypes.NewMerklePath([]byte("main"), []byte("gasPrice"))
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, res.Value)
	fmt.Println("VERIFY GNO GAS PRICE FROM TM LIGHTCLIENT CODE", err)
}

func verifyGnoAbsence() {
	var (
		// NOTE: Use key 'XX_does_not_exist_XX'
		// https://rpc.gno.land/abci_query?path=.store/main/key&data=WFhfZG9lc19ub3RfZXhpc3RfWFg=&prove=true
		abciResponseQueryBz = []byte(`{
      "ResponseBase": {
        "Error": null,
        "Data": null,
        "Events": null,
        "Log": "",
        "Info": ""
      },
      "Key": "WFhfZG9lc19ub3RfZXhpc3RfWFg=",
      "Value": null,
      "Proof": {
        "ops": [
          {
            "type": "iavl:a",
            "key": "WFhfZG9lc19ub3RfZXhpc3RfWFg=",
            "data": "igYKhwYKLAgeEPTdARiG5VciILE41jGsr8V4v6mzRqhD+ER+udyPprdAjcVV5bROiFRmCisIHBDiXBiG5VciIAwI7JBAxRPYTQrVbXjtH/zP2dZWqXAMWxqkRhjF4kA6CisIGhC0PBiG5VcqIC9uOSt/tfSRwzwFvfyUuaSG17Ps2/tVvu30PB820CV9CisIGBCiIBigjVcqINpzRVsEdbW8qmixdxkWNpJ48AjO51rhA7TnzilsAYm5CisIFhCIEBjO90IqIGaLY/HoZKK9GEpsLkikkMDLSqhtiMX1zrRrwldirGFYCikIEhCACBgCIiCNPlQZi5oJRcf2Xi/OCujX9MJCpm0nThjRuP+rLcAipgopCBAQgAQYAiIg+sye/nI7MkACzqkSawWVC7iN/krfeJ37cPRyS7zUnFQKKQgOEIACGAIiIGmV6Ebi0LmgUyRkieLDfQF0zWt8SVn6aR9WoPQdq0xYCikIDBCAARgCIiDv6ol1S2QNv7I9CcUaAXnU7hO4nFRJERIGwdArrDlhRwooCAoQQBgCIiBjcuD4UgUWfCp5HSqa01NpTMkMgppOGUnl3G205sIlNQooCAgQIBgCIiC7DiPos+ZT4F85y1kFPxla1k2jkoe1rGexAd6wep+4kAooCAYQEBgCKiAqzKonU9Wk1JbXapiHZ1vDbWudc+FrverG9kSmBGB5FAooCAQQCBgCIiCs3Z+zB8j1O36CYX7quMgM6nfoJaoJAw/CsKFV40EQgQooCAIQBBgCKiB9ty9svKHHu0dFw9XTglW1r94CVObSH/15B8MKJXKYYBIAGlEKKzlmZDc1MjAyYzg2NGYxNTMzNjdmMWM4MTA5NGUxYjY5YWM5NTZhMGI6MjQSIKz0x0omsZj9vJBW4fhBWFELt3zh6bD9BGBtlZb3aWuwGAIaUAoqYTA1ZTI1NzMwYWRlNzIxZWE0ZjI2ZmZmMDgzMmE5NDk0NmQxNmViMjoyEiDr8Yfucg30ShPM0uY/6aukBIzICZ/z2k7ZOGlCn32pWxgC"
          },
          {
            "type": "multistore",
            "key": "bWFpbg==",
            "data": "PAo6CjAKBG1haW4SKAomCIblVxIgrzsY8KuhjcPGO9GI0ZT45fW8OeXLecJ9VrVyljLn84YKBgoEYmFzZQ=="
          }
        ]
      },
      "Height": 719171
    }`)
		// app hash from https://rpc.gno.land/abci_info for height below
		appHash = "KhuwgJGgaL7ThS8TXPc5paw3kkOoypZM0G1CCr4gwlc="
	)
	var res gnoabci.ResponseQuery
	err := json.Unmarshal(abciResponseQueryBz, &res)
	if err != nil {
		panic(err)
	}
	prf := gnorootmulti.DefaultProofRuntime()
	proofOps := make(gnomerkle.ProofOperators, len(res.Proof.Ops))
	for i, op := range res.Proof.Ops {
		po, err := prf.Decode(op)
		if err != nil {
			panic(err)
		}
		proofOps[i] = po
	}
	spew.Dump(proofOps)

	// Verify proofs against app hash
	appHashBz, err := base64.StdEncoding.DecodeString(appHash)
	if err != nil {
		panic(err)
	}

	err = proofOps.Verify(appHashBz, "/main/XX_does_not_exist_XX", nil)
	fmt.Println("VERIFY GNO ABSENCE", err)
}

func verifyA1GovParams() {
	var (
		// Proof of gov params key existence (path=store/gov/key, data=0x30)
		// using https://atomone-rpc.allinbits.services/abci_query?path=%22store/gov/key%22&data=0x30&prove=true&height=3272353
		abciResponseQueryBz = []byte(`{
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
		appHash = "B2F11D67EE8D305A15234F3927D14074F8377B6AE1A2CD570E9F24BA50E0F7A4"
	)

	var res abci.ResponseQuery
	err := json.Unmarshal(abciResponseQueryBz, &res)
	if err != nil {
		panic(err)
	}

	// Turn tm proof into ics23 commitment proof used by tm light client
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
	appHashBz, err := hex.DecodeString(appHash)
	if err != nil {
		panic(err)
	}
	merkleRoot := commitmenttypes.NewMerkleRoot(appHashBz)
	specs := commitmenttypes.GetSDKSpecs()
	path := commitmenttypes.NewMerklePath([]byte("gov"), []byte{0x30})

	err = merkleProof.VerifyMembership(specs, merkleRoot, path, res.Value)
	fmt.Println("VERIFY A1 GOV PARAMS", err)
}

func verifyA1PacketReceipt() {
	var (
		// packet receipt proof
		// cmd: atomoned q ibc channel packet-receipt transfer channel-2 5
		// (executed at block 3284732)
		// the underlying command fetches the proof from the usual tendermint abci
		// query with:
		// - path = "store/ibc/key"
		// - data (==key) = `host.PacketReceiptKey(portID, channelID, sequence)`
		// and then turn it into a commitmenttypes.MerkleProof, so we just need to
		// decode it from that.
		proof = "Cv8GCvwGCjZyZWNlaXB0cy9wb3J0cy90cmFuc2Zlci9jaGFubmVscy9jaGFubmVsLTIvc2VxdWVuY2VzLzUSAQEaDggBGAEgASoGAALivqQBIi4IARIHAgTuoa0BIBohIB75sZwfkDzlhtkMWvBlVO/09xt32XuLSW3Gs1EFcVxcIi4IARIHBAjO3J0CIBohIKtT4vnhPqM+zHaDBUUxF++hHTcDHQaQm/72uRRts6C0Ii4IARIHBgzc3p0CIBohIBTxybwXEerVu2FhcmUF9dgR4sMOGcjpJB/7Oliar9ZqIi4IARIHCBr67Z0CIBohICCa73j418IwV7L+v3hjXt0h7gLacbhWb7gjT1o2CuRZIi4IARIHCjSYjp8CIBohIKMmb6s/bIrS2gTtk5GjmdR1vPWfsTFwYDYmwv2e84WuIi4IARIHDFy6jp8CIBohIMbHXPqlz/TIdZMLmFr2DCtG28WmxgkZAMmHWJ19Z6IHIi8IARIIDqQBuMyiAiAaISBMu69rcTqkOvH5yeH4heNGItguzVTcZYOEdhbaL4cF7yItCAESKRDUAsqhpAIg0r0Ym25yQ1YL7iXudXdE3GK12Z43Ho+UPAa/4io9k30gIi8IARIIEu4ElralAiAaISDce6GSOlj0zDVHnelIhXKONyawrvHBieZffPnlKeHNVSItCAESKRSmB5a2pQIgRp+zjKZ00W/NRt482naJ/P8n4+YQIKggq59HDtwC08ogIi8IARIIFqQPtMHCAiAaISDBUtua6oAfgs2xzL7M30XTxvFbqaKUO5hZGlE02HEsvyIvCAESCBiuGuKrhAMgGiEgemk0FvCYc15vtMpv29pJdIBwusmoRIFoA0TcUECMfCkiLwgBEgga1jPQ+5ADIBohIFiW4/yq3zE3g21/XQ5ATyav6joafuzKXTWkI+fKUHb7Ii0IARIpHIJK0PuQAyCfIHtkhwS2N3YElYPTj9lHSaTAodxrcWvfFfwzqK0BvyAiLQgBEikevHfQ+5ADIMEq49iU7lM1954NHBdwNvbYvYRga4VfG6h3Jus6OEKPICIuCAESKiDkpwLQ+5ADIBLQKsLEG+DPtcpquKcTwx1Mg47rfWS+vdqcEnB+UOVKICIuCAESKiSUzQfQ+5ADIN9RVUWddGT8w4hlUH5I+OIRXsOUapnfjMXYGwTm7tGNIAr+AQr7AQoDaWJjEiCi66yl8Vs8oNZp6bo7/sSzMynPtI9Iy233W5ki83oiUBoJCAEYASABKgEAIicIARIBARog71t4+pmoAEEVea1maHPr5hzLlyD/cxsk+U8+SsQdawAiJQgBEiEBbs7Zi9r/+uoeQizLo0dXgZnAM//N0yEiKTWldwb4h3EiJwgBEgEBGiAdZBXWzrgJl4GIz8kqNXnQeb2OUdLb//VrLWOeQuUTTyIlCAESIQEEnW5ItZcknp8R3x4+VdTI73Ixm2F9DIyi6vpYnLPj+SInCAESAQEaIPffzxYCAbdKN6vwBzoCzWg1t5X1Gp+5TFmnED+hRhfe"
		// cmd: atomoned q block 3284732|jq '.block.header.app_hash'
		appHash = "416D75EE392246A41BEA5FBD350C13A5EA54DD9F57F75DB5991C3EB3D4BBACF0"
	)

	var merkleProof commitmenttypes.MerkleProof
	bz, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		panic(err)
	}
	err = merkleProof.Unmarshal(bz)
	if err != nil {
		panic(err)
	}
	appHashBz, err := hex.DecodeString(appHash)
	if err != nil {
		panic(err)
	}
	merkleRoot := commitmenttypes.NewMerkleRoot(appHashBz)
	specs := commitmenttypes.GetSDKSpecs()
	key := host.PacketReceiptKey("transfer", "channel-2", 5)
	path := commitmenttypes.NewMerklePath([]byte("ibc"), key)
	value := []byte{byte(1)} // value of packet receipt is always 1
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, value)
	fmt.Println("VERIFY A1 PACKET RECEIPT", err)
}
