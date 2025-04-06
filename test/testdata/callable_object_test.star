load("@plugin", "os", t = "test")

def test_basic():
    task = _create_task(prefix = "[prefix-1]")
    t.equal("[prefix-1]: greeting-1", task("greeting-1"))
    t.equal("[prefix-1]: greeting-2", task("greeting-2"))

    task = task.with_overrides(prefix = "[prefix-2]")
    t.equal("[prefix-2]: greeting-1", task("greeting-1"))
    t.equal("[prefix-2]: greeting-2", task("greeting-2"))

    task = task.with_overrides(prefix = "[prefix-3]")
    t.equal("[prefix-3]: greeting-1", task("greeting-1"))
    t.equal("[prefix-3]: greeting-2", task("greeting-2"))

def _create_task(prefix):
    def task(greeting):
        return prefix + ": " + greeting

    def with_overrides(prefix = prefix):
        return _create_task(prefix)

    task = callable_object(task)
    task.with_overrides = with_overrides

    return task
