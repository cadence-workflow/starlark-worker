load("@plugin", "testplugin")

def plus(a, b):
    return a + b

def stringify(*args):
    return testplugin.stringify_activity(*args)
