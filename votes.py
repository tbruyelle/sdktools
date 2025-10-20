#!/usr/bin/env python3

import node
import sys


def main():
    n = node.Node("https://atomone-api.allinbits.services")
    print("Fetching all votes...")
    votes = n.get_votes(sys.argv[1])
    print(f"Fetched {len(votes)} votes.")

    if not votes:
        print("No votes found.")
        return

    m = {}
    for vote in votes:
        addr = vote["voter"]
        print(f"Fetching delegations for voter {addr}...")
        dels = n.get_delegator_delegations(addr)
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
