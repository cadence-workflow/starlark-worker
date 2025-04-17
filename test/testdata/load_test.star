load("@plugin", t = "test")
load("//test/testdata/lib/math.star", "PI")
load("//test/testdata/resources/catalog.json", catalog = "json")
load("//test/testdata/resources/data.txt", data = "txt")

expected_data = """The quick brown fox
jumps over the lazy dog
"""

expected_catalog = {
    "type": "catalog",
    "items": [
        {
            "id": 100,
            "alias": "foo",
        },
        {
            "id": 101,
            "alias": "bar",
        },
    ],
}

def test_loads():
    t.equal(expected_data, data)
    t.equal(expected_catalog, catalog)
    t.equal("3.14159265359", str(PI))
