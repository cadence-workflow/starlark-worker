load("@plugin", "concurrent", "time")

def simple_task():
    return 0

def task_with_arg(value):
    return value

def slow_task(sleep_time):
    time.sleep(seconds=sleep_time)
    return sleep_time

def test_batch_run_simple():
    """Test basic batch_run functionality with simple tasks"""
    callables = [
        concurrent.new_callable(simple_task),
        concurrent.new_callable(task_with_arg, 5),
        concurrent.new_callable(slow_task, 10)
    ]
    
    batch_future = concurrent.batch_run(callables)
    return batch_future.get()

def test_batch_run_with_max_concurrency_less_than_task_list_length():
    """Test batch_run with max concurrency less than task list length"""
    callables = [
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10)
    ]
    workflow_start_time = time.time_ns()
    batch_future = concurrent.batch_run(callables, max_concurrency=2)
    got_futures = batch_future.get()
    workflow_duration = time.time_ns() - workflow_start_time
    return workflow_duration

def test_batch_run_with_max_concurrency_greater_than_task_list_length():
    """Test batch_run with max concurrency greater than task list length"""
    callables = [
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10)
    ]
    workflow_start_time = time.time_ns()
    batch_future = concurrent.batch_run(callables, max_concurrency=10)
    got_futures = batch_future.get()
    workflow_duration = time.time_ns() - workflow_start_time
    return workflow_duration

def test_batch_run_with_max_concurrency_equal_to_task_list_length():
    """Test batch_run with max concurrency equal to task list length"""
    callables = [
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 5),
        concurrent.new_callable(slow_task, 10)
    ]
    workflow_start_time = time.time_ns()
    batch_future = concurrent.batch_run(callables, max_concurrency=len(callables))
    got_futures = batch_future.get()
    workflow_duration = time.time_ns() - workflow_start_time
    return workflow_duration

def test_batch_future_is_ready_happy_path():
    """Test batch_future.is_ready() functionality happy path"""
    callables = [
        concurrent.new_callable(simple_task)
    ]
    
    batch_future = concurrent.batch_run(callables)
    
    batch_future.get()
    return batch_future.is_ready()

def test_batch_future_is_ready_before_get():
    """Test batch_future.is_ready() functionality when it is called before get"""
    callables = [
        concurrent.new_callable(simple_task)
    ]
    batch_future = concurrent.batch_run(callables)
    return batch_future.is_ready()

def test_batch_future_get_futures():
    """Test batch_future.get_futures() functionality"""
    callables = [
        concurrent.new_callable(simple_task),
        concurrent.new_callable(task_with_arg, 10)
    ]
    
    batch_future = concurrent.batch_run(callables)
    futures = batch_future.get_futures()
    
    # verifying actual futures is hard, so we will call result() on each of the individual futures
    # and verify the end result instead
    results = []
    for i in range(len(futures)):
        results.append(futures[i].result())
    return results

