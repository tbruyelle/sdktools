#!/usr/bin/env python3

import bech32
import node

PROPOSAL_ID = 21
ACCOUNT_PREFIX = "atone"


def convert_valoper_to_account(valoper_address):
    """Convert a validator operator address (atonevaloper1...) to its
    account address (atone1...)."""
    _, data = bech32.bech32_decode(valoper_address)
    return bech32.bech32_encode(ACCOUNT_PREFIX, data)


def format_vote(vote):
    """Render a vote's options as a short human-readable string."""
    parts = []
    for opt in vote.get("options", []):
        name = opt["option"].replace("VOTE_OPTION_", "")
        weight = opt.get("weight", "1")
        if float(weight) == 1:
            parts.append(name)
        else:
            parts.append(f"{name}({weight})")
    return "+".join(parts) if parts else "-"


def main():
    n = node.Node("https://atomone-api.allinbits.services")

    print("Fetching all validators...")
    validators = n.get_validators()
    if not validators:
        print("No validators found.")
        return

    # account address -> validator moniker, so we can tell if a delegator
    # is itself a validator.
    val_by_account = {}
    for validator in validators:
        account = convert_valoper_to_account(validator["operator_address"])
        val_by_account[account] = validator["description"]["moniker"]

    # voter address -> vote on the proposal.
    print(f"Fetching votes for proposal {PROPOSAL_ID}...")
    vote_by_addr = {}
    for vote in n.get_votes(PROPOSAL_ID):
        vote_by_addr[vote["voter"]] = format_vote(vote)

    # delegator address -> total staked amount across all validators.
    m = {}
    for validator in validators:
        valAddr = validator["operator_address"]
        print(f"Fetching delegations for validator {valAddr}...")
        dels = n.get_delegations(valAddr)
        for del_ in dels:
            addr = del_["delegation"]["delegator_address"]
            amount = int(del_["balance"]["amount"])
            m[addr] = m.get(addr, 0) + amount
    print("Done")

    top = sorted(m.items(), key=lambda item: item[1], reverse=True)[:100]

    print()
    print(f"{'Delegator':<44} {'Amount':>20} {'Validator':<24} Prop {PROPOSAL_ID}")
    print("-" * 100)
    for delAddr, amount in top:
        validator = val_by_account.get(delAddr, "-")
        vote = vote_by_addr.get(delAddr, "did not vote")
        print(f"{delAddr:<44} {amount:>20,} {validator:<24} {vote}")


main()
