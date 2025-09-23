#!/usr/bin/env python3

import requests
import base64
import hashlib


def get_block(height):
    url = f"{BASE_URL}/cosmos/base/tendermint/v1beta1/blocks/{height}"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        block = data.get("block", [])
        return block
    else:
        print(f"Error fetching block: {response.status_code}")
        return []


def get_tx(hash):
    url = f"{BASE_URL}/cosmos/tx/v1beta1/txs/{hash}"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        tx = data.get("tx_response", [])
        return tx
    else:
        print(f"Error fetching tx: {response.status_code}")
        return []


def main():
    global BASE_URL
    BASE_URL = "https://atomone-testnet-1-api.allinbits.services"
    height = 3563389
    block = get_block(height)

    gas_used = 0
    for tx in block["data"]["txs"]:
        # base64 decode tx
        bz = base64.b64decode(tx)
        # hash bz to get tx hash
        tx_hash = hashlib.sha256(bz).hexdigest()
        # get tx by hash
        tx_resp = get_tx(tx_hash)

        gas_used += int(tx_resp["gas_used"])
        # print(f"gas used {tx_resp['gas_used']}/{tx_resp['gas_wanted']}")

    print(f"Total gas used for block {height}: {gas_used:,}")


main()
