import bip39
import hashlib

entropy = hashlib.sha256(b"my custom entropy").digest()
mnemonic = bip39.encode_bytes(entropy)
print(mnemonic)
