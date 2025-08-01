load("@plugin", "os", "time", t = "test")

def test_sleep():
    start_ts = time.time_ns()
    time.sleep(seconds = 5)
    total = time.time_ns() - start_ts
    t.equal(5000000000, total)

def test_utc_format_seconds():
    t.equal("1970-01-01T00:00:00", time.utc_format_seconds("%Y-%m-%dT%H:%M:%S", 0.0))
    t.equal("2024-05-14T21:59:29", time.utc_format_seconds("%Y-%m-%dT%H:%M:%S", 1715723969.5412211))

def test_time():
    seconds = time.time()
    t.equal("float", type(seconds))
    datestr = time.utc_format_seconds("%Y-%m-%d", seconds)
    t.equal("string", type(datestr))

def effective_time_test():
    t.equal("unix:1753361232", os.environ["STARLARK_TIME"])
    t.equal(float(1753361232), time.time())

    sleep_seconds = 5
    time.sleep(seconds = sleep_seconds)
    t.equal(float(1753361232 + sleep_seconds), time.time())
