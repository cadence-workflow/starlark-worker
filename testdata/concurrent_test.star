load("@plugin", "concurrent", "time", t = "test")

def test_run():
    def task(n):
        time.sleep(seconds = 5)  # task takes 5 seconds
        return n * n

    start_ns = time.time_ns()

    futures = []
    for i in [1, 2, 3, 4, 5]:
        f = concurrent.run(task, i)
        futures.append(f)

    res = [
        f.result()
        for f in futures
    ]

    total_ns = time.time_ns() - start_ns

    t.equal([1, 4, 9, 16, 25], res)
    t.equal(5000000000, total_ns)  # 5 seconds
