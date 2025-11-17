package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveBech32(t *testing.T) {
	mnemonic := "burden junk salon cabbage energy damp view camp pole endorse isolate arrange struggle reflect easy hawk chat social finish prepare wagon utility drive input"

	t.Run("cosmos", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "", hdPaths["cosmos"], "atone")

		assert.Equal(t, "atone1rku58s0axgpex6e2uuarxpcrzu3gyur2wkhyqd", bech)
	})

	t.Run("cosmos+passphrase", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "passphrase", hdPaths["cosmos"], "atone")

		assert.Equal(t, "atone159k2tt0ruh8jlyz5q4fjmjxuxpg4pvdp333avs", bech)
	})
}
