load("@plugin", "cad", "os", "time", "uuid", "concurrent", "progress")

def wf(n = 3, sleep_duration_sec = 2):
    def task(i):
        progress.report("task %d started" % i)
        time.sleep(seconds = sleep_duration_sec)
        progress.report("task %d returning" % i)
        return "Hello from task %d" % i

    print("concurrent_hello is called with n: %d, sleep_duration_sec: %d. run id is %s" % (n, sleep_duration_sec, cad.execution_run_id))

    start_ns = time.time_ns()

    futures = []
    for i in range(0, n):
        f = concurrent.run(task, i)
        futures.append(f)

    res = [
        f.result()
        for f in futures
    ]

    return res
