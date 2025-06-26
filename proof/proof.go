package main

import (
	"context"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/davecgh/go-spew/spew"

	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	gnoclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	gnomerkle "github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	gnoiavl "github.com/gnolang/gno/tm2/pkg/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	gnorootmulti "github.com/gnolang/gno/tm2/pkg/store/rootmulti"
)

func main() {
	// spew.Config.DisableMethods = true

	verifyGnoGasPrice()
	verifyA1GovParams()
	verifyGnoAbsence()
	verifyA1Absence()

	// TODO find how to determine key to r/ibc packet commitment/ack
	// objectID: oid:<OBJECT_ID> where OBJECT_ID is REALM_ID:sequence
	// is it possible to fetch that from a gno sc and returns it to the
	// caller ?
	// One solution consists of using std.SetParamStrings in the gno realm, so
	// it's easy for the relayer to determine the key of the commitment packet
	// knowing his sequence number. Keys registered by `std.SetParamStrings` have
	// the following format: `vm:RLM_PATH:key`, so considering a `r/sys/ibc`
	// realm, the key of a commitment packet would look like:
	// `vm:gno.land/r/sys/ibc:07-tendermint-1142` where
	// - `vm:gno.land/r/sys/ibc:` is the prefix added by `std.SetParamStrings`
	// - `07-tendermint-1` is the client id
	// - `1` is the byte indicator of a commitment packet (2 for receipt and 3 for
	// ack)
	// - `42` is the sequence number of the commitment packet
}

func verifyGnoGasPrice() {
	var (
		path = ".store/main/key"
		key  = []byte("gasPrice")
	)
	height := int64(10)
	qres, err := gnocli().ABCIQueryWithOptions(
		path, key, gnoclient.ABCIQueryOptions{
			Height: height,
			Prove:  true,
		})
	if err != nil {
		panic(err)
	}

	// Decode tm2 proof
	prf := gnorootmulti.DefaultProofRuntime()
	tmp2Proofs := make(gnomerkle.ProofOperators, len(qres.Response.Proof.Ops))
	for i, op := range qres.Response.Proof.Ops {
		po, err := prf.Decode(op)
		if err != nil {
			panic(err)
		}
		tmp2Proofs[i] = po
	}

	// Verify proofs against app hash
	height++
	rres, err := gnocli().Block(&height)
	if err != nil {
		panic(err)
	}

	err = tmp2Proofs.VerifyValue(rres.Block.Header.AppHash, "/main/gasPrice", qres.Response.Value)
	fmt.Println("VERIFY GNO GAS PRICE", err)

	// TODO Turn tm2 proofs into ics23 proofs so it can be used by the default
	// 07-tendermint light client implementation
	ics23Proofs := make([]*ics23.CommitmentProof, len(tmp2Proofs))
	for i, p := range tmp2Proofs {
		fmt.Println(i, p)
		switch pp := p.(type) {
		case rootmulti.MultiStoreProofOp:
			ics23Proofs[i] = &ics23.CommitmentProof{}
		case gnoiavl.IAVLValueOp:
			ics23Proofs[i] = &ics23.CommitmentProof{
				Proof: &ics23.CommitmentProof_Exist{
					Exist: &ics23.ExistenceProof{
						Key:   pp.GetKey(),
						Value: qres.Response.Value,
						Leaf:  convertLeafOp(height - 1),
						Path:  convertInnerOps(pp.Proof.LeftPath),
					},
				},
			}
		case gnoiavl.IAVLAbsenceOp:
			// TODO
		}
	}
	merkleProof := commitmenttypes.MerkleProof{Proofs: ics23Proofs}
	merkleRoot := commitmenttypes.NewMerkleRoot(rres.Block.Header.AppHash)
	specs := commitmenttypes.GetSDKSpecs()
	mpath := commitmenttypes.NewMerklePath([]byte("main"), []byte("gasPrice"))
	err = merkleProof.VerifyMembership(specs, merkleRoot, mpath, qres.Response.Value)
	fmt.Println("VERIFY GNO GAS PRICE FROM TM LIGHTCLIENT CODE", err)
}

func verifyGnoAbsence() {
	var (
		path = ".store/main/key"
		key  = []byte("does_not_exist_XX")
	)
	height := int64(10)
	qres, err := gnocli().ABCIQueryWithOptions(
		path, key, gnoclient.ABCIQueryOptions{
			Height: height,
			Prove:  true,
		})
	if err != nil {
		panic(err)
	}

	// Decode tm2 proof
	prf := gnorootmulti.DefaultProofRuntime()
	proofOps := make(gnomerkle.ProofOperators, len(qres.Response.Proof.Ops))
	for i, op := range qres.Response.Proof.Ops {
		po, err := prf.Decode(op)
		if err != nil {
			panic(err)
		}
		proofOps[i] = po
	}

	// Verify proofs against app hash
	height++
	rres, err := gnocli().Block(&height)
	if err != nil {
		panic(err)
	}

	err = proofOps.Verify(rres.Block.Header.AppHash, "/main/does_not_exist_XX", nil)
	fmt.Println("VERIFY GNO ABSENCE", err)
}

func verifyA1GovParams() {
	var (
		ctx  = context.Background()
		path = "store/gov/key" // path to gov module store
		key  = []byte{0x30}    // key used to store gov params
	)
	infres, err := a1cli().ABCIInfo(ctx)
	if err != nil {
		panic(err)
	}
	// Get a recent height
	height := infres.Response.LastBlockHeight - 10

	// Get proof
	reqres, err := a1cli().ABCIQueryWithOptions(ctx, path, key,
		rpcclient.ABCIQueryOptions{
			Height: height,
			Prove:  true,
		})
	if err != nil {
		panic(err)
	}

	// Decode ics23 proof
	proofs := make([]*ics23.CommitmentProof, len(reqres.Response.ProofOps.Ops))
	spew.Dump(reqres.Response.ProofOps.Ops)
	for i, op := range reqres.Response.ProofOps.Ops {
		var p ics23.CommitmentProof
		err = p.Unmarshal(op.Data)
		if err != nil || p.Proof == nil {
			panic(fmt.Sprintf("could not unmarshal proof op into CommitmentProof at index %d: %v", i, err))
		}
		proofs[i] = &p
	}
	spew.Dump(proofs)
	merkleProof := commitmenttypes.MerkleProof{Proofs: proofs}

	// Get app hash for proof height (must use following block to get app hash)
	height++
	blockres, err := a1cli().Block(ctx, &height)
	if err != nil {
		panic(err)
	}

	var (
		merkleRoot = commitmenttypes.NewMerkleRoot(blockres.Block.Header.AppHash)
		specs      = commitmenttypes.GetSDKSpecs()
		mpath      = commitmenttypes.NewMerklePath([]byte("gov"), key)
	)
	err = merkleProof.VerifyMembership(specs, merkleRoot, mpath, reqres.Response.Value)
	fmt.Println("VERIFY A1 GOV PARAMS", err)
}

func verifyA1Absence() {
	var (
		ctx  = context.Background()
		path = "store/gov/key"             // path to gov module store
		key  = []byte("does_not_exist_XX") // unknown key in gov module store
	)
	infres, err := a1cli().ABCIInfo(ctx)
	if err != nil {
		panic(err)
	}
	// Get a recent height
	height := infres.Response.LastBlockHeight - 10

	// Get proof
	reqres, err := a1cli().ABCIQueryWithOptions(ctx, path, key,
		rpcclient.ABCIQueryOptions{
			Height: height,
			Prove:  true,
		})
	if err != nil {
		panic(err)
	}

	// Decode ics23 proof
	proofs := make([]*ics23.CommitmentProof, len(reqres.Response.ProofOps.Ops))
	for i, op := range reqres.Response.ProofOps.Ops {
		var p ics23.CommitmentProof
		err = p.Unmarshal(op.Data)
		if err != nil || p.Proof == nil {
			panic(fmt.Sprintf("could not unmarshal proof op into CommitmentProof at index %d: %v", i, err))
		}
		proofs[i] = &p
	}
	merkleProof := commitmenttypes.MerkleProof{Proofs: proofs}

	// Get app hash for proof height (must use following block to get app hash)
	height++
	blockres, err := a1cli().Block(ctx, &height)
	if err != nil {
		panic(err)
	}

	var (
		merkleRoot = commitmenttypes.NewMerkleRoot(blockres.Block.Header.AppHash)
		specs      = commitmenttypes.GetSDKSpecs()
		mpath      = commitmenttypes.NewMerklePath([]byte("gov"), key)
	)
	err = merkleProof.VerifyNonMembership(specs, merkleRoot, mpath)
	fmt.Println("VERIFY A1 ABSENCE", err)
}
