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

def test_batch_run_with_limited_max_concurrency_less_than_task_list_length():
    """Test batch_run with limited concurrency"""

    def slow_task(sleep_time, id):
        current_time = time.time_ns()
        time.sleep(seconds=sleep_time)
        return sleep_time
        
    callables = [
        concurrent.new_callable(slow_task, 5, 0),
        concurrent.new_callable(slow_task, 10, 1),
        concurrent.new_callable(slow_task, 5, 2),
        concurrent.new_callable(slow_task, 5, 3),
        concurrent.new_callable(slow_task, 10, 4)
    ]
    
    workflow_start_time = time.time_ns()
    batch_future = concurrent.batch_run(callables, max_concurrency=2)
    res = batch_future.get()
    workflow_duration = time.time_ns() - workflow_start_time
    
    t.equal([5, 10, 5, 5, 10], res)
    t.equal(20000000000, workflow_duration)  # 20 seconds

def test_batch_run_with_max_concurrency_equal_to_task_list_length():
    """Test batch_run with concurrency equal to task list length"""

    def slow_task(sleep_time, id):
        current_time = time.time_ns()
        time.sleep(seconds=sleep_time)
        return sleep_time
        
    callables = [
        concurrent.new_callable(slow_task, 5, 0),
        concurrent.new_callable(slow_task, 10, 1),
        concurrent.new_callable(slow_task, 5, 2),
        concurrent.new_callable(slow_task, 5, 3),
        concurrent.new_callable(slow_task, 10, 4)
    ]
    
    workflow_start_time = time.time_ns()
    batch_future = concurrent.batch_run(callables)
    res = batch_future.get()
    workflow_duration = time.time_ns() - workflow_start_time
    
    t.equal([5, 10, 5, 5, 10], res)
    t.equal(10000000000, workflow_duration)  # 10 seconds
