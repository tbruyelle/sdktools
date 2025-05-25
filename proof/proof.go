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
	// app hash from https://atomone-rpc.allinbits.services/block?height=3228573
	atomOneAppHash = "FE07EE6F56AC82346A67D0ECABCDFDF7513D7748714D843AA8A26389EBA88548"

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
}
