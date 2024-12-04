load("@plugin", "stringify_activity")

def plus(a, b):
    return a + b

def stringify(*args):
    return stringify_activity(*args)
