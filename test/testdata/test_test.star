load("@plugin", t = "test")

def test_equal():
    t.equal(101, int("101"))

def test_not_equal():
    t.not_equal("foo", "bar")
    t.not_equal(0, False)

def test_true():
    t.true(True)
    t.true(1 == 1)
    t.true([0])
    t.true({"id": 0})
    t.true("foo")
    t.true(8)

def test_false():
    t.false(False)
    t.false(1 == 2)
    t.false([])
    t.false({})
    t.false("")
    t.false(0)
