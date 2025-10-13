#!/usr/bin/env python3

import node
import base64
import hashlib


def main():
    n = node.Node("https://atomone-testnet-1-api.allinbits.services")
    height = 3563389
    block = n.get_block(height)

    gas_used = 0
    for tx in block["data"]["txs"]:
        # base64 decode tx
        bz = base64.b64decode(tx)
        # hash bz to get tx hash
        tx_hash = hashlib.sha256(bz).hexdigest()
        # get tx by hash
        tx_resp = n.get_tx(tx_hash)

        gas_used += int(tx_resp["gas_used"])
        # print(f"gas used {tx_resp['gas_used']}/{tx_resp['gas_wanted']}")

    print(f"Total gas used for block {height}: {gas_used:,}")


main()
