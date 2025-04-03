load("@plugin", t = "test")

def test_attributes():
    o1 = dataclass(
        p1 = "v1",
        p2 = "v2",
    )
    t.equal(["p1", "p2"], dir(o1))
    t.true(hasattr(o1, "p1"))
    t.true(hasattr(o1, "p2"))
    t.false(hasattr(o1, "p3"))

    t.equal(o1.p1, "v1")
    t.equal(o1.p2, "v2")

    o1.p1 = "v11"
    t.equal(o1.p1, "v11")
    t.equal(o1.p2, "v2")

def test_comparison_operators():
    o1 = dataclass(
        p1 = "v1",
        p2 = "v2",
    )
    o1_clone = dataclass(
        p1 = "v1",
        p2 = "v2",
    )
    o2 = dataclass(
        p1 = "v11",
        p2 = "v22",
    )
    o3 = dataclass(
        p11 = "v1",
        p22 = "v2",
    )

    t.equal(o1, o1)
    t.equal(o1, o1_clone)

    t.true(o1 == o1)
    t.true(o1 == o1_clone)
    t.true(o1 != o2)
    t.true(o1 != o3)
    t.true(o2 != o3)

    t.false(o1 != o1)
    t.false(o1 != o1_clone)
    t.false(o1 == o2)
    t.false(o1 == o3)
    t.false(o2 == o3)

def test_empty():
    o = dataclass()
    t.equal([], dir(o))
    t.false(bool(o))

def test_builtin_methods():
    o1 = dataclass(
        p1 = "v1",
        p2 = "v2",
    )
    t.equal("dataclass", type(o1))
    t.equal('<dataclass {p1: "v1", p2: "v2"}>', str(o1))
