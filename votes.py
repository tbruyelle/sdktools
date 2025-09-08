#!/usr/bin/env python3

import requests
import urllib.parse


def get_validators():
    return paginate(f"{BASE_URL}/cosmos/staking/v1beta1/validators", "validators")


def paginate(baseurl, items_key):
    next_key = ""
    items = []
    print(f"Fetching {baseurl}...\n")
    while True:
        url = f"{baseurl}?pagination.key={next_key}"
        response = requests.get(url)
        if response.status_code != 200:
            print(f"Error fetching {url}: {response.status_code}")
        data = response.json()
        items += data.get(items_key, [])
        print(f"{len(items)} fetched items")
        next_key = data.get("pagination").get("next_key")
        if next_key is None:
            return items
        next_key = urllib.parse.quote_plus(next_key)


def get_votes():
    return paginate(f"{BASE_URL}/atomone/gov/v1/proposals/15/votes", "votes")


def get_delegator_delegations(addr):
    return paginate(
        f"{BASE_URL}/cosmos/staking/v1beta1/delegations/{addr}",
        "delegation_responses",
    )


def main():
    global BASE_URL
    BASE_URL = "https://atomone-api.allinbits.services"
    print("Fetching all votes...")
    votes = get_votes()
    print(f"Fetched {len(votes)} votes.")

    if not votes:
        print("No votes found.")
        return

    m = {}
    for vote in votes:
        addr = vote["voter"]
        print(f"Fetching delegations for voter {addr}...")
        dels = get_delegator_delegations(addr)
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
