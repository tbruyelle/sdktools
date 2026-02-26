#!/usr/bin/env python3
"""Find an AtomOne RPC node that has unpruned state at a given block height.

Usage:
    ./find-archive-node.py <height>
    ./find-archive-node.py 6965425
"""

import json
import sys
import urllib.request
import urllib.parse
import concurrent.futures

CHAIN = "atomone"
REGISTRY_URL = f"https://chains.cosmos.directory/{CHAIN}"
TIMEOUT = 15

ABCI_PATH = "/cosmos.bank.v1beta1.Query/TotalSupply"


def get_rpc_nodes():
    """Fetch RPC node list from cosmos.directory chain registry."""
    req = urllib.request.Request(REGISTRY_URL, headers={"User-Agent": "govbox/1.0"})
    with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
        data = json.load(resp)
    apis = data.get("chain", {}).get("apis", {})
    return [r["address"] for r in apis.get("rpc", []) if "address" in r]


def check_node(rpc_url, height):
    """Check if a node has unpruned state at the given height.
    Returns (rpc_url, True/False, detail_string).
    """
    rpc_url = rpc_url.rstrip("/")
    query = urllib.parse.urlencode({"path": f'"{ABCI_PATH}"', "height": height})
    url = f"{rpc_url}/abci_query?{query}"
    try:
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
            data = json.load(resp)
        code = data["result"]["response"]["code"]
        if code == 0:
            return (rpc_url, True, "state available")
        log = data["result"]["response"].get("log", "")
        if "pruned" in log.lower():
            return (rpc_url, False, "pruned")
        return (rpc_url, False, log[:120])
    except Exception as e:
        return (rpc_url, False, f"error: {e}")


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <height>")
        sys.exit(1)

    height = int(sys.argv[1])
    print(f"Fetching RPC node list from {REGISTRY_URL} ...")
    nodes = get_rpc_nodes()
    print(f"Found {len(nodes)} nodes. Checking state at height {height} ...\n")

    found = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as pool:
        futures = {pool.submit(check_node, url, height): url for url in nodes}
        for f in concurrent.futures.as_completed(futures):
            url, ok, detail = f.result()
            status = "OK" if ok else "NO"
            print(f"  [{status:2s}] {url}  ({detail})")
            if ok:
                found.append(url)

    print()
    if found:
        print(f"Nodes with state at height {height}:")
        for url in found:
            print(f"  {url}")
    else:
        print(f"No node found with unpruned state at height {height}.")


if __name__ == "__main__":
    main()
