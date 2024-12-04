""""""

load("@plugin", "atexit", "concurrent", "os", "request", "time", "uuid", t = "test")

TEST_SERVER_URL = os.environ["TEST_SERVER_URL"]

def test_concurrent_run():
    futures = [
        concurrent.run(task, 1),
        concurrent.run(task, 2),
        concurrent.run(task, 3),
    ]
    res = [
        f.result()
        for f in futures
    ]
    t.equal([1, 4, 9], res)

# This function doesn't start with test_ prefix for a reason. It's tested separately from other test_ functions.
# See integration_test.Suite.TestAtExit
def injected_error_test():
    futures = [
        concurrent.run(task, 4),
        concurrent.run(task, 3, injected_error = "injected error"),
        concurrent.run(task, 2),
        concurrent.run(task, 1),
        concurrent.run(task, 2, injected_error = "injected error"),
        concurrent.run(task, 3),
        concurrent.run(task, 4),
    ]
    return [
        f.result()
        for f in futures
    ]

def task(delay_seconds, injected_error = None, exit_hook = True):
    # create resource

    resource_url = "{}/at/{}.json".format(TEST_SERVER_URL, uuid.uuid4().hex)
    res = request.do("PUT", resource_url, body = bytes([delay_seconds]))
    t.equal(200, res.status_code)

    # define and register resource clean up function

    def delete_resource():
        res = request.do("DELETE", resource_url)
        t.equal(200, res.status_code)

    if exit_hook:
        atexit.register(delete_resource)

    # simulate resource work

    res = request.do("GET", resource_url)
    t.equal(200, res.status_code)

    time.sleep(delay_seconds)

    if injected_error:
        fail(injected_error)

    # close resource and unregister the handler

    delete_resource()
    atexit.unregister(delete_resource)

    return delay_seconds * delay_seconds
