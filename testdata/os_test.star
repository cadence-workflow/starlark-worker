load("@plugin", "os", t = "test")

def test_env():
    t.equal(1, len(os.environ), os.environ)
    t.true("TEST_SERVER_URL" in os.environ, os.environ)

    # add env var
    os.environ["FOO"] = "BAR"

    t.equal(2, len(os.environ), os.environ)
    t.equal("BAR", os.environ["FOO"])
