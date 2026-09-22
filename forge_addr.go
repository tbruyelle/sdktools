package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"flag"
	"fmt"
	"hash"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosmos/go-bip39"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160" //nolint:staticcheck // cosmos addresses are hash160
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

// NOTE: similar to https://github.com/atomone-hub/atonevanity
//
// Deriving a brand new mnemonic for every candidate is dominated by the BIP39
// PBKDF2 (2048 HMAC-SHA512 rounds, ~700µs), which alone caps throughput at
// ~1.2k addr/s/core. So each worker draws a single mnemonic, derives the
// account chain m/44'/<coin>'/0'/0 once, and then scans the address index.
// Since the last path element is not hardened, a candidate costs one
// HMAC-SHA512 plus one EC base multiplication (~15µs). The hit is still a
// plain BIP39 mnemonic, just at index i instead of 0.

const (
	charset  = "qpzry9x8gf2tvdw0s3jn54khce6mua7l" // bech32
	hardened = uint32(0x80000000)
	// Points are produced in Jacobian form and converted to affine a whole
	// batch at a time, so the modular inversion is paid once per batch
	// instead of once per candidate (~8µs of the 23µs).
	batchSize = 256
	// Indices stay well inside the non-hardened range; rotating the mnemonic
	// costs one PBKDF2 amortized over 16M candidates.
	indicesPerMnemonic = 1 << 24
)

type target struct {
	hrp      string
	prefix   string
	suffix   string
	coinType uint32

	// The bech32 data part encodes the address 5 bits at a time, most
	// significant first, so a prefix of n chars pins the first 5n bits of the
	// 20 byte address. Matching those bits directly skips the bech32 encoding
	// for all but a vanishing fraction of the candidates.
	want, mask [20]byte
	maskLen    int
}

type result struct {
	addr     string
	mnemonic string
	path     string
}

func main() {
	if err := run(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		hrp      = flag.String("hrp", "g", "bech32 human readable prefix")
		prefix   = flag.String("prefix", "th0mas", "wanted prefix, right after the '1' separator")
		suffix   = flag.String("suffix", "", "wanted suffix at the very end of the address")
		coinType = flag.Uint("cointype", uint(types.CoinType), "BIP44 coin type")
		cpus     = flag.Int("cpus", runtime.NumCPU(), "number of workers")
	)
	flag.Parse()

	t, err := newTarget(*hrp, *prefix, *suffix, uint32(*coinType))
	if err != nil {
		return err
	}
	fmt.Printf("Looking for %s1%s…%s — 1 in %s addresses, %d workers\n",
		t.hrp, t.prefix, t.suffix, humanize(t.odds()), *cpus)

	var (
		tries atomic.Uint64
		once  sync.Once
		res   result
		start = time.Now()
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < *cpus; i++ {
		g.Go(func() error {
			return search(ctx, t, &tries, func(r result) {
				once.Do(func() { res = r; cancel() })
			})
		})
	}
	go report(ctx, t, &tries, start)

	if err := g.Wait(); err != nil {
		return err
	}
	if res.addr == "" {
		return fmt.Errorf("stopped after %d tries without a match", tries.Load())
	}
	// The search bypasses the sdk to go fast, so confirm the hit really is
	// what the sdk derives before trusting it.
	if err := verify(t, res); err != nil {
		return fmt.Errorf("found %s but it does not check out: %w", res.addr, err)
	}
	fmt.Printf("\nFOUND %s after %d tries in %s\n", res.addr, tries.Load(), time.Since(start).Round(time.Second))
	fmt.Println("mnemonic:", res.mnemonic)
	fmt.Println("hd path: ", res.path)
	return nil
}

// search scans mnemonics until ctx is cancelled, reporting any match to found.
func search(ctx context.Context, t *target, tries *atomic.Uint64, found func(result)) error {
	var (
		privs [batchSize]secp.ModNScalar
		pts   [batchSize]secp.JacobianPoint
		acc   [batchSize]secp.FieldVal
		pub   [33]byte
		data  [37]byte // parent compressed pubkey || big endian index
		addr  = make([]byte, 0, ripemd160.Size)
		rip   = ripemd160.New()
	)
	for {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return err
		}
		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			return err
		}
		// m/44'/<coin>'/0'/0
		priv, chainCode := hd.ComputeMastersFromSeed(bip39.NewSeed(mnemonic, ""))
		for _, step := range []struct {
			index  uint32
			harden bool
		}{{44, true}, {t.coinType, true}, {0, true}, {0, false}} {
			priv, chainCode = deriveChild(priv, chainCode, step.index, step.harden)
		}
		var parent secp.ModNScalar
		parent.SetBytes(&priv)
		copy(data[:33], compressed(&parent, &pub))
		mac := hmac.New(sha512.New, chainCode[:])

		for base := uint32(0); base < indicesPerMnemonic; base += batchSize {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			for j := range privs {
				binary.BigEndian.PutUint32(data[33:], base+uint32(j))
				privs[j] = childScalar(mac, data[:], &parent)
				secp.ScalarBaseMultNonConst(&privs[j], &pts[j])
			}
			toAffine(pts[:], acc[:])
			for j := range pts {
				pub[0] = byte(2 + pts[j].Y.IsOddBit())
				pts[j].X.PutBytesUnchecked(pub[1:])
				sum := sha256.Sum256(pub[:])
				rip.Reset()
				rip.Write(sum[:]) //nolint:errcheck // never errors
				addr = rip.Sum(addr[:0])
				if t.matches(addr) {
					found(result{
						addr:     t.bech32(addr),
						mnemonic: mnemonic,
						path:     t.path(base + uint32(j)),
					})
					return nil
				}
			}
			tries.Add(batchSize)
		}
	}
}

// childScalar returns the non-hardened BIP32 child of parent for the index
// already written into data. It mirrors the sdk's derivePrivateKey, which
// reduces mod n instead of rejecting IL >= n like BIP32 says; that only
// differs with probability 2^-128 and the sdk is what will import the key.
func childScalar(mac hash.Hash, data []byte, parent *secp.ModNScalar) secp.ModNScalar {
	mac.Reset()
	mac.Write(data) //nolint:errcheck // never errors
	var (
		i  [sha512.Size]byte
		il [32]byte
		s  secp.ModNScalar
	)
	mac.Sum(i[:0])
	copy(il[:], i[:32])
	s.SetBytes(&il)
	s.Add(parent)
	if s.IsZero() {
		// Not a valid key; keep the point on the curve so the batch inversion
		// stays well defined, the address simply will not match.
		s.SetInt(1)
	}
	return s
}

// deriveChild is the sdk's hd.derivePrivateKey, kept here to also return the
// chain code.
func deriveChild(priv, chainCode [32]byte, index uint32, harden bool) ([32]byte, [32]byte) {
	var data []byte
	if harden {
		index |= hardened
		data = append([]byte{0}, priv[:]...)
	} else {
		var (
			k   secp.ModNScalar
			pub [33]byte
		)
		k.SetBytes(&priv)
		data = append(data, compressed(&k, &pub)...)
	}
	data = binary.BigEndian.AppendUint32(data, index)

	mac := hmac.New(sha512.New, chainCode[:])
	mac.Write(data) //nolint:errcheck // never errors
	var (
		i  [sha512.Size]byte
		il [32]byte
		s  secp.ModNScalar
	)
	mac.Sum(i[:0])
	copy(il[:], i[:32])
	copy(chainCode[:], i[32:])

	s.SetBytes(&il)
	var p secp.ModNScalar
	p.SetBytes(&priv)
	s.Add(&p)
	return s.Bytes(), chainCode
}

// compressed writes the SEC1 compressed public key of k into pub.
func compressed(k *secp.ModNScalar, pub *[33]byte) []byte {
	var p secp.JacobianPoint
	secp.ScalarBaseMultNonConst(k, &p)
	p.ToAffine()
	pub[0] = byte(2 + p.Y.IsOddBit())
	p.X.PutBytesUnchecked(pub[1:])
	return pub[:]
}

// toAffine normalizes a batch of Jacobian points using Montgomery's trick, so
// the whole batch shares a single field inversion.
func toAffine(pts []secp.JacobianPoint, acc []secp.FieldVal) {
	var run secp.FieldVal
	run.SetInt(1)
	for j := range pts {
		acc[j] = run
		run.Mul(&pts[j].Z).Normalize()
	}
	inv := run
	inv.Inverse()
	for j := len(pts) - 1; j >= 0; j-- {
		var zInv, z2, z3 secp.FieldVal
		zInv.Set(&inv).Mul(&acc[j]).Normalize()
		inv.Mul(&pts[j].Z).Normalize()
		z2.SquareVal(&zInv).Normalize()
		z3.Mul2(&z2, &zInv).Normalize()
		pts[j].X.Mul(&z2).Normalize()
		pts[j].Y.Mul(&z3).Normalize()
		pts[j].Z.SetInt(1)
	}
}

func newTarget(hrp, prefix, suffix string, coinType uint32) (*target, error) {
	t := &target{hrp: hrp, prefix: prefix, suffix: suffix, coinType: coinType}
	if prefix == "" && suffix == "" {
		return nil, fmt.Errorf("need at least one of -prefix or -suffix")
	}
	if len(prefix) > 32 {
		return nil, fmt.Errorf("prefix %q is longer than the 32 char data part", prefix)
	}
	for _, s := range []string{prefix, suffix} {
		for _, c := range s {
			if !strings.ContainsRune(charset, c) {
				return nil, fmt.Errorf("%q is not a bech32 char, pick from %s", string(c), charset)
			}
		}
	}
	for i := 0; i < len(prefix); i++ {
		v := strings.IndexByte(charset, prefix[i])
		for b := 0; b < 5; b++ {
			bit := i*5 + b
			if v&(1<<(4-b)) != 0 {
				t.want[bit/8] |= 0x80 >> (bit % 8)
			}
			t.mask[bit/8] |= 0x80 >> (bit % 8)
		}
	}
	t.maskLen = (len(prefix)*5 + 7) / 8
	return t, nil
}

func (t *target) matches(addr []byte) bool {
	for i := 0; i < t.maskLen; i++ {
		if addr[i]&t.mask[i] != t.want[i] {
			return false
		}
	}
	// The suffix lands on the bech32 checksum, which depends on the whole
	// address, so it can only be checked on the encoded string.
	return t.suffix == "" || strings.HasSuffix(t.bech32(addr), t.suffix)
}

func (t *target) bech32(addr []byte) string {
	s, err := types.Bech32ifyAddressBytes(t.hrp, addr)
	if err != nil {
		panic(err)
	}
	return s
}

func (t *target) path(index uint32) string {
	return fmt.Sprintf("m/44'/%d'/0'/0/%d", t.coinType, index)
}

// odds returns the expected number of candidates per hit.
func (t *target) odds() float64 {
	return math.Pow(32, float64(len(t.prefix)+len(t.suffix)))
}

// verify re-derives the hit through the sdk, the slow way.
func verify(t *target, res result) error {
	priv, err := hd.Secp256k1.Derive()(res.mnemonic, "", res.path)
	if err != nil {
		return err
	}
	addr, err := types.Bech32ifyAddressBytes(t.hrp, hd.Secp256k1.Generate()(priv).PubKey().Address())
	if err != nil {
		return err
	}
	if addr != res.addr {
		return fmt.Errorf("sdk derives %s", addr)
	}
	return nil
}

func report(ctx context.Context, t *target, tries *atomic.Uint64, start time.Time) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n := tries.Load()
			rate := float64(n) / time.Since(start).Seconds()
			eta := time.Duration(t.odds()/rate) * time.Second
			fmt.Printf("\r%s tries, %s addr/s, ~%s expected  ",
				humanize(float64(n)), humanize(rate), eta.Round(time.Second))
		}
	}
}

func humanize(f float64) string {
	for _, unit := range []string{"", "k", "M", "G", "T"} {
		if f < 1000 {
			return strings.TrimSuffix(fmt.Sprintf("%.1f", f), ".0") + unit
		}
		f /= 1000
	}
	return fmt.Sprintf("%.1fP", f)
}
