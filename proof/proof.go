package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"

	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
)

var (
	// ------------
	// AtomOne proof
	// Proof of gov params key existence
	// using https://atomone-rpc.allinbits.services/abci_query?path=%22store/gov/key%22&data=0x30&prove=true&height=3228573
	atomOneTmProofBz = []byte(`{
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
      }`)
	value = []byte("ChMKBnVhdG9uZRIJNTEyMDAwMDAwEgQIgOpJGgQIgN9uIhQwLjI1MDAwMDAwMDAwMDAwMDAwMCoUMC42NjcwMDAwMDAwMDAwMDAwMDA6FDAuMTAwMDAwMDAwMDAwMDAwMDAwcAF6FDAuMDEwMDAwMDAwMDAwMDAwMDAwggEUMC4yNTAwMDAwMDAwMDAwMDAwMDCKARQwLjkwMDAwMDAwMDAwMDAwMDAwMJIBFDAuMjUwMDAwMDAwMDAwMDAwMDAwmgEUMC45MDAwMDAwMDAwMDAwMDAwMDCiAQQIgLxpqgEECICjBQ==")
	// app hash from https://atomone-rpc.allinbits.services/block?height=3228573
	atomOneAppHash = "39F2564ECF16A07C62283773ED9CD6A990EC8EAA449379F469FF02277EE7B579"

	//----------
	// Gno proof
	gnoTm2ProofBz = []byte(`{
        "ops": [
          {
            "type": "iavl:v",
            "key": "Z2FzUHJpY2U=",
            "data": "pwUKpAUKLAgeEJDdARiisTwiIJWXfrnf4FN11EzJUs/6etgXUtWdGsS7ardTCnIbmLDRCisIHBCmXBiisTwiIPh5hAhi46NB5CxiOoHBjjpN3CG66NSNGLGb6oqpMB3kCisIGhCQPBiisTwiIF358abbBoePxEvMbPbwiX8x419bHX5lt9aco986oGFgCisIGBD+GxiisTwqICn/rYj2gVU3sSIaMFVW4U57rqJwxbRyxnhkRqGQa/SnCisIFhCeEBiisTwiIIoG+h/yyV5OcSmzs5+LqxfTk7hOSGA5Jssball5+2d3CisIFBCKCBiisTwiIHfzys8A4eP67H0WmmcCJ3uXdCdhVslydAVnejEv2x+gCisIEBCABBiisTwiIJzzUhruzOjIGxqy3XHv6szhYnmpqoWncoaq04HSuRdRCisIDhCAAhiisTwqIPR/pcvgNdgUICxaC5+EfSWajv3PuNAveq78jgolqdWTCisIDBCAARiisTwqIHSIu9wK0pjwS6BJC2LXNU+OBruzNQ4AhFxxmZcXno1MCioIChBAGKKxPCogkBOFQ4fze4rfpyRjQkC6yY975kUeP7QSjd0XHkO7x4wKKggIECAYorE8KiD+MHdW+K+TJIi5f3Z0HKKb2txg0FiOMhDJ57rGpPH+rAoqCAYQEBiisTwiIJY5g9TlmAlMFUhJweBxm7AMVjEMiRa5kd5AYwqUEM/vCioIBBAIGKKxPCoggakESPyw+BlnlLm5FWghAtvkFneYMvYRdAAEkerb8ncKKggCEAQYorE8IiBbt5DisYgvM78Isi7Xg/zg2n+5r0X8wNZa10TAzNZcsRowCghnYXNQcmljZRIgoQVt4nMhH3LZH6D1JMQXjvzN5DX++k+Wa6m4XcEwejEYorE8"
          },
          {
            "type": "multistore",
            "key": "bWFpbg==",
            "data": "PAo6CjAKBG1haW4SKAomCKKxPBIgAedJzU9avKtZY6tjz2jbxXweb9xK9wVNFXgGmpHiQKwKBgoEYmFzZQ=="
          }
        ]
      }`)
)

func main() {
	// Unmarshal tm proofs
	var (
		atomOneTmProof crypto.ProofOps
		gnoTm2Proof    crypto.ProofOps
	)
	err := json.Unmarshal([]byte(atomOneTmProofBz), &atomOneTmProof)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(gnoTm2ProofBz), &gnoTm2Proof)
	if err != nil {
		panic(err)
	}

	// Turn AtomOne tm proof into ics23 commitment proof used by tm light client
	proofs := make([]*ics23.CommitmentProof, len(atomOneTmProof.Ops))
	for i, op := range atomOneTmProof.Ops {
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

	// NOTE error bc "value" does not match the value of the first
	// key "gov" from the proof, which is normal considering that "value" is the
	// value of the other key 0x30 (the gov params)...
	// "path" seems correct since VerifyMembership comments that it should be
	// composed of the module key then the key itself.
	// Are we supposed to pass 2 values for each different keys? There's no parameter for that...
	// TODO check the proofs in mintscan RecvPacket, try to unmarshal them to see if there's
	// multiple proof for the differents merke tree like we have here.
	err = merkleProof.VerifyMembership(specs, merkleRoot, path, value)
	fmt.Println(err)
}
