#!/usr/bin/env python3

import node


def main():
    n = node.Node("https://rest.cosmos.directory/osmosis")
    print("Fetching all denoms...")
    denoms = n.get_ibc_denoms()
    print(f"Fetched {len(denoms)} denoms.")

    if not denoms:
        print("No denoms found.")
        return

    for d in denoms:
        if d["base_denom"] == "uatone":
            print(d)


main()
