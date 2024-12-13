load("@plugin", "random", t = "test")

def test_run():
    test_randint()
    test_random()

def test_randint():
    # Generate 1000 random numbers from 1 to 10,
    # then perform a simple validation by checking frequency of numbers are close to expected frequency
    # An ideal validation could use Chi-Square Goodness-of-Fit Test

    numbers = {i:0 for i in range(1, 11)}
    for i in range(0, 1000):
        n = random.randint(1, 10)
        numbers[n] = numbers[n] + 1

    expected_freq = 1000 / 10
    tolerance = 0.25 * expected_freq  # 25% tolerance

    # Validate distribution
    for number, count in numbers.items():
        t.true(abs(count - expected_freq) < tolerance, "expected ~%d occurrences of %d but found %d" % (expected_freq, number, count))

def test_random():
    # Generate 1000 random float numbers between 0 to 1, take average and validate it's close to 0.5

    total = 0
    for i in range(0, 1000):
        n = random.random()
        total = total + n

    got_avg = total / 1000.0
    expected_avg = 0.5
    tolerance = 0.1

    t.true(abs(got_avg - expected_avg) < tolerance, "expected average to be in [%d, %d] range but found %d" % (expected_avg-tolerance, expected_avg+tolerance, got_avg))
