""""""

load("@plugin", "os", "request", "uuid", t = "test")

TEST_SERVER_URL = os.environ["TEST_SERVER_URL"]

def test_json():
    headers = {
        "accept": ["application/json"],
    }

    resource_url = "{}/{}{}".format(TEST_SERVER_URL, uuid.uuid4().hex, "/data/test.json")
    res = request.do("GET", resource_url, headers = headers)
    t.equal(404, res.status_code)

    res = request.do("PUT", resource_url, body = bytes('{"data":["foo","bar"]}'))
    t.equal(200, res.status_code)

    res = request.do("GET", resource_url)
    t.equal(200, res.status_code)
    t.equal({"data": ["foo", "bar"]}, res.json())
