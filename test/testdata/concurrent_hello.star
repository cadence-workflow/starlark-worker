load("@plugin", "concurrent", "progress", "random", "time", "workflow")

def wf(n = 10, sleep_duration_sec = 3):
    def task(i):
        jittered = sleep_duration_sec + random.random()

        # reported progress can be queried via Cadence UI or QueryWorkflow API
        progress.report("task %d started. will sleep for %f" % (i, jittered))
        time.sleep(seconds = jittered)
        progress.report("task %d returning" % i)
        return "Hello from task %d" % i

    # print outputs will be in worker logs as if they were logged via workflow.GetLogger(ctx).Info()
    # print outputs are also queriable
    print("concurrent_hello is called with n: %d, sleep_duration_sec: %d. run id is %s" % (n, sleep_duration_sec, workflow.execution_run_id))

    random.seed(1234)
    futures = []
    for i in range(0, n):
        f = concurrent.run(task, i)
        futures.append(f)

    res = [
        f.result()
        for f in futures
    ]

    return res
