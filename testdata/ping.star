load("@plugin", "cad", "os", "time", "uuid")
load("//testdata/lib/math.star", "PI")

def main(verbose = True, **keywords):
    result = {
        "message": "PI is {}".format(PI),
        "execution_id": cad.execution_id,
        "verbose": verbose,
        "keywords": keywords,
    }
    if not verbose:
        return result

    rand_uuid = uuid.uuid4()
    extra = {
        "execution_run_id": cad.execution_run_id,
        "epoch_nanosec": time.time_ns(),
        "epoch_sec": time.time(),
        "environ": os.environ,
        "uuid": {
            "urn": rand_uuid.urn,
            "hex": rand_uuid.hex,
        },
    }
    result.update(extra)
    return result
