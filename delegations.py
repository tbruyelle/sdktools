#!/usr/bin/env python3

import bech32
import requests
import sys


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
    else:
        print(f"Error fetching validators: {response.status_code}")
        return []


def get_delegations(val):
    url = f"{BASE_URL}/cosmos/staking/v1beta1/validators/{val}/delegations"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        dels = data.get("delegation_responses", [])
        return dels
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
            addr = del_["delegation"]["delegator_address"]
            if addr not in m:
                m[addr] = int(del_["balance"]["amount"])
            else:
                m[addr] += int(del_["balance"]["amount"])
    print("Done")
    m = dict(sorted(m.items(), key=lambda item: item[1], reverse=True))
    for delAddr, amount in m.items():
        print(delAddr, amount)


main()
