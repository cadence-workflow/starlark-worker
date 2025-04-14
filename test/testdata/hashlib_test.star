"""
Test hash plugin
"""

load("@plugin", "hashlib", t = "test")

def test_hash():
    data = "test-str which is used to generate hash value"
    digest_size = 64
    digest_val = hashlib.blake2b_hex(data, digest_size = digest_size)
    t.equal("string", type(digest_val))
    t.equal(digest_size * 2, len(digest_val))
