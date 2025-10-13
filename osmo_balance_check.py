#!/usr/bin/env python3

import sys
import node


denoms = {
    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2": "atom",
    "ibc/BC26A7A805ECD6822719472BCB7842A48EF09DF206182F8F259B2593EB5D23FB": "atone",
    "ibc/D6E02C5AE8A37FC2E3AB1FC8AC168878ADB870549383DFFEA9FD020C234520A7": "photon",
}


def main():
    # Put args 1 in ADDR
    addr = sys.argv[1]
    n = node.Node("https://rest.cosmos.directory/osmosis")
    print("Fetching balances...")
    balances = n.get_balances(addr)

    if not balances:
        print("No balances found.")
        return

    for balance in balances:
        amount = float(balance["amount"]) / 1000000
        denom = balance["denom"]
        # if denom present in denoms
        if denom in denoms:
            denom = denoms[denom]

        print(f"{denom}\t{amount}")


main()
