#!/usr/bin/env python3

import bech32
import node
import sys

BASE_URLS = {
    "cosmos": "https://cosmos-api.polkachu.com",
    "osmo": "https://osmosis-api.polkachu.com",
    "atomone": "https://atomone-api.polkachu.com",
}
BECHS = {
    "cosmos": "cosmos",
    "osmo": "osmo",
    "atomone": "atone",
}


def convert_valoper_to_account(bech, valoper_address):
    """
    Converts a validator operator address to a basic account address.
    """
    # Decode the valoper address
    hrp, data = bech32.bech32_decode(valoper_address)

    # Encode the data with the new prefix
    account_address = bech32.bech32_encode(bech, data)

    return account_address


def main():
    # check there is at least one parameter that match one of the BASE_URLS
    if len(sys.argv) < 2 or sys.argv[1] not in BASE_URLS:
        print("Usage:")
        for url in BASE_URLS:
            print(f"\t{sys.argv[0]} {url}")
        return

    # Set the base URL to use based on command line parameter
    n = node.Node(BASE_URLS[sys.argv[1]])
    bech = BECHS[sys.argv[1]]
    # Step 1: Get all validators
    print("Fetching all validators...")
    validators = n.get_validators()

    if not validators:
        print("No validators found.")
        return

    # Step 2: Query accounts for each validator
    for validator in validators:
        account = convert_valoper_to_account(bech, validator["operator_address"])
        # print(
        #     f"Fetching accounts for validator: {validator['operator_address']} {account}"
        # )
        info = n.get_account(account)

        acc = info[0]["account"]

        # pubkey = f"{acc['@type']} - {acc['pub_key']['@type']}"
        if "pub_key" in acc:
            pubkey = acc["pub_key"]
        else:
            pubkey = "XXXX"

        print(f"Accounts {pubkey} {validator['description']['moniker']}")


main()
