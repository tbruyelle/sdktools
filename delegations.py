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


def get_delegations(val):
    """
    Fetches the list of all validators from the RPC endpoint.
    Returns:
        List of validators' operator addresses.
    """
    url = f"{BASE_URL}/cosmos/staking/v1beta1/validators/{val}/delegations"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        dels = data.get("delegation_responses", [])
        return dels
        # return [validator["operator_address"] for validator in validators]
    else:
        print(f"Error fetching validator delegations: {response.status_code}")
        return []


def main():
    global BASE_URL
    BASE_URL = "https://atomone-testnet-1-api.allinbits.services"
    print("Fetching all validators...")
    validators = get_validators()

    if not validators:
        print("No validators found.")
        return

    m = {}
    for validator in validators:
        valAddr = validator["operator_address"]
        print(f"Fetching delegations for validator {valAddr}...")
        dels = get_delegations(valAddr)
        for del_ in dels:
            m[del_["delegation"]["delegator_address"]] = int(del_["balance"]["amount"])
    print("Done")
    m = dict(sorted(m.items(), key=lambda item: item[1], reverse=True))
    for delAddr, amount in m.items():
        print(delAddr, amount)


main()
