#!/usr/bin/env python3

import bech32
import requests
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


def get_validators():
    """
    Fetches the list of all validators from the RPC endpoint.
    Returns:
        List of validators' operator addresses.
    """
    url = f"{BASE_URL}/cosmos/staking/v1beta1/validators"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        validators = data.get("validators", [])
        return validators
        # return [validator["operator_address"] for validator in validators]
    else:
        print(f"Error fetching validators: {response.status_code}")
        return []


def convert_valoper_to_account(valoper_address):
    """
    Converts a validator operator address to a basic account address.
    """
    # Decode the valoper address
    hrp, data = bech32.bech32_decode(valoper_address)

    # Encode the data with the new prefix
    account_address = bech32.bech32_encode(BECH, data)

    return account_address


def get_accounts_by_validator(address):
    url = f"{BASE_URL}/cosmos/auth/v1beta1/accounts/{address}"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        # delegations = data.get("delegation_responses", [])
        return [
            data
            # delegation["delegation"]["delegator_address"] for delegation in delegations
        ]
    else:
        print(
            f"Error fetching accounts for validator {address}: {response.status_code}"
        )
        return []


def main():
    # check there is at least one parameter that match one of the BASE_URLS
    if len(sys.argv) < 2 or sys.argv[1] not in BASE_URLS:
        print("Usage:")
        for url in BASE_URLS:
            print(f"\t{sys.argv[0]} {url}")
        return

    # Set the base URL to use based on command line parameter
    global BASE_URL
    BASE_URL = BASE_URLS[sys.argv[1]]
    global BECH
    BECH = BECHS[sys.argv[1]]
    # Step 1: Get all validators
    print("Fetching all validators...")
    validators = get_validators()

    if not validators:
        print("No validators found.")
        return

    # Step 2: Query accounts for each validator
    for validator in validators:
        account = convert_valoper_to_account(validator["operator_address"])
        # print(
        #     f"Fetching accounts for validator: {validator['operator_address']} {account}"
        # )
        info = get_accounts_by_validator(account)

        acc = info[0]["account"]

        # pubkey = f"{acc['@type']} - {acc['pub_key']['@type']}"
        if "pub_key" in acc:
            pubkey = acc["pub_key"]
        else:
            pubkey = "XXXX"

        print(f"Accounts {pubkey} {validator['description']['moniker']}")


main()
