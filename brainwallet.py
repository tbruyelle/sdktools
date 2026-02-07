import bip39
import hashlib
import sys

entropy = hashlib.sha256(" ".join(sys.argv[1:]).encode()).digest()
mnemonic = bip39.encode_bytes(entropy)
print(mnemonic)
