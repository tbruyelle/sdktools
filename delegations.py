#!/usr/bin/env python3

import node


def main():
    n = node.Node("https://atomone-testnet-1-api.allinbits.services")
    print("Fetching all validators...")
    validators = n.get_validators()

    if not validators:
        print("No validators found.")
        return

    m = {}
    for validator in validators:
        valAddr = validator["operator_address"]
        print(f"Fetching delegations for validator {valAddr}...")
        dels = n.get_delegations(valAddr)
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
