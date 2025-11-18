package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveBech32(t *testing.T) {
	mnemonic := "burden junk salon cabbage energy damp view camp pole endorse isolate arrange struggle reflect easy hawk chat social finish prepare wagon utility drive input"

	t.Run("cosmos", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "", configs["cosmos"])

		assert.Equal(t, "atone1rku58s0axgpex6e2uuarxpcrzu3gyur2wkhyqd", bech)
	})

	t.Run("cosmos+passphrase", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "passphrase", configs["cosmos"])

		assert.Equal(t, "atone159k2tt0ruh8jlyz5q4fjmjxuxpg4pvdp333avs", bech)
	})
	t.Run("cosmos+customHRP", func(t *testing.T) {
		cfg := configs["cosmos"]
		cfg.hrp = "osmo"
		bech := deriveBech32(mnemonic, "", cfg)

		assert.Equal(t, "osmo1rku58s0axgpex6e2uuarxpcrzu3gyur2gdcnq8", bech)
	})

	t.Run("btc", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "", configs["segwit"])

		assert.Equal(t, "bc1qnqqu45huquxzz6sysr7denuxmg0mh09mq2usc6", bech)
	})
	t.Run("btc+passphrase", func(t *testing.T) {
		bech := deriveBech32(mnemonic, "passphrase", configs["segwit"])

		assert.Equal(t, "bc1qam7vrj9ycmm3p380jpe0u57hsdz6nc04qn5e9t", bech)
	})
}
