"""
Test progress plugin
"""

load("@plugin", "progress")

def test_progress_report():
    task_progress = {
        "task_name": "test_name",
        "task_path": "uberai.test_name",
        "task_log": "test-log-url",
    }
    progress.report(str(task_progress))
