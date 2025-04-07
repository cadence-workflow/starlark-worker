load("@plugin", "workflow")

def test_execution_id():
    return workflow.execution_id

def test_execution_run_id():
    return workflow.execution_run_id
