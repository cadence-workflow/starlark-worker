load("@plugin", "uuid", t = "test")

def test_uuid4():
    rand_uuid = uuid.uuid4()
    rand_uuid_str = str(rand_uuid)
    hex = rand_uuid.hex
    t.equal("string", type(hex), hex)
    t.equal(36, len(rand_uuid_str), rand_uuid_str)
    t.equal(rand_uuid_str, hex[0:8] + "-" + hex[8:12] + "-" + hex[12:16] + "-" + hex[16:20] + "-" + hex[20:], rand_uuid_str)

def test_uuid4_equal():
    uuid1 = uuid.uuid4()
    uuid2 = uuid.uuid4()

    t.equal(uuid1, uuid1)
    t.not_equal(uuid1, uuid2)
