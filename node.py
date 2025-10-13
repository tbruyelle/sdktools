#!/usr/bin/env python3


import requests
import urllib.parse


class Node:
    def __init__(self, baseurl):
        self.url = baseurl

    def get_validators(self):
        return self.__paginate("/cosmos/staking/v1beta1/validators", "validators")

    def get_delegations(self, val):
        return self.__paginate(
            "/cosmos/staking/v1beta1/validators/{val}/delegations",
            "delegation_responses",
        )

    def get_votes(self):
        return self.__paginate("/atomone/gov/v1/proposals/15/votes", "votes")

    def get_delegator_delegations(self, addr):
        return self.__paginate(
            "/cosmos/staking/v1beta1/delegations/{addr}",
            "delegation_responses",
        )

    def get_block(self, height):
        url = f"{self.url}/cosmos/base/tendermint/v1beta1/blocks/{height}"
        response = requests.get(url)
        if response.status_code == 200:
            data = response.json()
            block = data.get("block", [])
            return block
        else:
            print(f"Error fetching block: {response.status_code}")
            return []

    def get_balances(self, addr):
        return self.__paginate(f"/cosmos/bank/v1beta1/balances/{addr}", "balances")

    def get_account(self, address):
        url = f"{self.url}/cosmos/auth/v1beta1/accounts/{address}"
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

    def get_tx(self, hash):
        url = f"{self.url}/cosmos/tx/v1beta1/txs/{hash}"
        response = requests.get(url)
        if response.status_code == 200:
            data = response.json()
            tx = data.get("tx_response", [])
            return tx
        else:
            print(f"Error fetching tx: {response.status_code}")
            return []

    def __paginate(self, uri, items_key):
        next_key = ""
        items = []
        url = self.url + uri
        print(f"Fetching {url}...\n")
        while True:
            url = f"{url}?pagination.key={next_key}"
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
