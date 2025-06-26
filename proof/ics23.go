package main

import (
	"encoding/binary"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"

	gnomerkle "github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	gnoiavl "github.com/gnolang/gno/tm2/pkg/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
)

// Turn tm2 proofs into ics23 proofs so it can be used by the default
// 07-tendermint light client implementation.
// TODO wip
func convertProof(value []byte, height int64, tm2Proofs gnomerkle.ProofOperators) []*ics23.CommitmentProof {
	ics23Proofs := make([]*ics23.CommitmentProof, len(tm2Proofs))
	for i, p := range tm2Proofs {
		fmt.Println(i, p)
		switch pp := p.(type) {
		case rootmulti.MultiStoreProofOp:
			ics23Proofs[i] = &ics23.CommitmentProof{}
		case gnoiavl.IAVLValueOp:
			ics23Proofs[i] = &ics23.CommitmentProof{
				Proof: &ics23.CommitmentProof_Exist{
					Exist: &ics23.ExistenceProof{
						Key:   pp.GetKey(),
						Value: value,
						Leaf:  convertLeafOp(height - 1),
						Path:  convertInnerOps(pp.Proof.LeftPath),
					},
				},
			}
		case gnoiavl.IAVLAbsenceOp:
			ics23Proofs[i] = &ics23.CommitmentProof{
				Proof: &ics23.CommitmentProof_Nonexist{
					Nonexist: &ics23.NonExistenceProof{
						Key: pp.GetKey(),
						// TODO fill
						Left:  nil,
						Right: nil,
					},
				},
			}
		}
	}
	return ics23Proofs
}

func convertLeafOp(version int64) *ics23.LeafOp {
	var varintBuf [binary.MaxVarintLen64]byte
	// this is adapted from iavl/proof.go:proofLeafNode.Hash()
	prefix := convertVarIntToBytes(0, varintBuf)
	prefix = append(prefix, convertVarIntToBytes(1, varintBuf)...)
	prefix = append(prefix, convertVarIntToBytes(version, varintBuf)...)

	return &ics23.LeafOp{
		Hash:         ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
		Prefix:       prefix,
	}
}

// we cannot get the proofInnerNode type, so we need to do the whole path in one function
func convertInnerOps(path gnoiavl.PathToLeaf) []*ics23.InnerOp {
	steps := make([]*ics23.InnerOp, 0, len(path))

	// lengthByte is the length prefix prepended to each of the sha256 sub-hashes
	var lengthByte byte = 0x20

	var varintBuf [binary.MaxVarintLen64]byte

	// we need to go in reverse order, iavl starts from root to leaf,
	// we want to go up from the leaf to the root
	for i := len(path) - 1; i >= 0; i-- {
		// this is adapted from iavl/proof.go:proofInnerNode.Hash()
		prefix := convertVarIntToBytes(int64(path[i].Height), varintBuf)
		prefix = append(prefix, convertVarIntToBytes(path[i].Size, varintBuf)...)
		prefix = append(prefix, convertVarIntToBytes(path[i].Version, varintBuf)...)

		var suffix []byte
		if len(path[i].Left) > 0 {
			// length prefixed left side
			prefix = append(prefix, lengthByte)
			prefix = append(prefix, path[i].Left...)
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
		} else {
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
			// length-prefixed right side
			suffix = []byte{lengthByte}
			suffix = append(suffix, path[i].Right...)
		}

		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: prefix,
			Suffix: suffix,
		}
		steps = append(steps, op)
	}
	return steps
}

func convertVarIntToBytes(orig int64, buf [binary.MaxVarintLen64]byte) []byte {
	n := binary.PutVarint(buf[:], orig)
	return buf[:n]
}
