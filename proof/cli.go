package main

import (
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	gnoclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

var _gnocli gnoclient.Client

func gnocli() gnoclient.Client {
	if _gnocli == nil {
		remote := "http://localhost:26657"
		// remote = "https://rpc.gno.land:443"
		cli, err := gnoclient.NewHTTPClient(remote)
		if err != nil {
			panic(err)
		}
		_gnocli = cli
	}
	return _gnocli
}

var _a1cli rpcclient.Client

func a1cli() rpcclient.Client {
	if _a1cli == nil {
		remote := "https://atomone-rpc.allinbits.services:443"
		cli, err := rpchttp.New(remote, "")
		if err != nil {
			panic(err)
		}
		_a1cli = cli
	}
	return _a1cli
}
