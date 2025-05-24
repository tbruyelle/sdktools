package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/davecgh/go-spew/spew"

	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"

	abci "github.com/cometbft/cometbft/abci/types"
)

var (
	// ------------
	// AtomOne proof
	// Proof of gov params key existence
	// using https://atomone-rpc.allinbits.services/abci_query?path=%22store/gov/key%22&data=0x30&prove=true&height=3228573
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
            "data": "CvYFCgEwEuUBChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBRoLCAEYASABKgMAAgIiKwgBEgQCBAIgGiEgjkBfizMZnHPckgMhtvluJjWvXT3288+1zy6yhhSxJ/QiLAgBEigEBo7UgAMgC3FGm4XkJ940SMUTV88+GaktknbWwZELhuQitp0ostggIiwIARIoCBC0sIgDIObhZQlpgV7lAAkE4EVPzyRQ78D/Wga7zuRVOCLn8bPCICIsCAESKAoi+vqIAyCT4SY5DUfwYFEvLUveiyVhWajDqx5lxsBhH3/1Mjl3FyAiLAgBEigMPPr6iAMg717IfgAGlJB7r0r1V68XEUp5QWySWU6oc8NpVuyYxPkgIiwIARIoDmr2uokDIDm8WtAcrIvwHxyX4zhefYsD2H/hFnCtD9I27+o8WzB2ICItCAESKRDMAeiLigMglcFJ6jojUIaUR9GFRNo85CXFq3Um+zj2NLC1ig0twN4gIi0IARIpEqAD6IuKAyD56BAuYPzSpTlKdqqbdu00F9w6e9FQcppO3IggIp9T8yAiLQgBEikW2gfoi4oDIF7g4dXbu+Ees0dbcgGn2R+HKVKlrQAB3zC3tHAv1sDmICItCAESKRjQDuiLigMgFT9ye2vtlOWLJvtrWA8rcgUntHbT6fphhhsaQvQbWWsgIi0IARIpGvAZ6IuKAyArGw3giBvl2G14KcWWIQmB/eHVddTIqB7QynZk8l6R2iA="
          },
          {
            "type": "ics23:simple",
            "key": "Z292",
            "data": "CvsBCgNnb3YSIHYu5lhY1MjSOFkTVzDuGbpuYuRlTi1L/gkZWEWs1vezGgkIARgBIAEqAQAiJQgBEiEBEaKNvFesWykCp7xJPv9ZB24/kB+z/T6qrqUsnrdZCikiJwgBEgEBGiB22GWXyqA0nUDuhJ6l5g2/RerAXbNZW5p6O3Ppq+CLUiInCAESAQEaIAMIj7i/ftScmosPiftJUwyblliaESAz5r7MuVXbCQzFIiUIARIhAc5KUUKxy7uK/US0dMeb2QO6qWpBpL/B4HNv3nIQidGxIicIARIBARogOLHYCrCpD4iHEru1fNQ26Xv852/wl6g6BcdRabD05U0="
          }
        ]
      },
      "height": "3228573",
      "codespace": ""
    }`)
	// app hash from https://atomone-rpc.allinbits.services/block?height=3228573
	atomOneAppHash = "39F2564ECF16A07C62283773ED9CD6A990EC8EAA449379F469FF02277EE7B579"

	//----------
	// Gno proof
	// TODO
)

func main() {
	spew.Config.DisableMethods = true
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
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, res.Value)
	fmt.Println(err)
	spew.Dump(proofs[0].GetExist().Value)
	spew.Dump(res.Value)
}
