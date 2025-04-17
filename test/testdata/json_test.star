load("@plugin", "json", t = "test")
load("../testdata/lib/math.star", "PI")

def test_json_dumps():
    t.equal('"foo"', json.dumps("foo"))
    t.equal("15", json.dumps(15))
    t.equal("3.14159265359", json.dumps(PI))
    t.equal("true", json.dumps(True))
    t.equal("false", json.dumps(False))
    t.equal("null", json.dumps(None))
    t.equal('[1,"foo",true,{"id":-101}]', json.dumps([1, "foo", True, {"id": -101}]))

def test_json_loads():
    t.equal("foo", json.loads('"foo"'))
    t.equal(15, json.loads("15"))
    t.equal(PI, json.loads("3.14159265359"))
    t.equal(True, json.loads("true"))
    t.equal(False, json.loads("false"))
    t.equal(None, json.loads("null"))
    t.equal([1, "foo", True, {"id": -101}], json.loads('[1,"foo",true,{"id":-101}]'))

def test_json_loads_dataclass():
    # load dataclass (meta) at root
    meta_json = '{"last_updated":"2023-10-01T19:32:08Z","__codec__":"dataclass"}'
    meta = json.loads(meta_json)
    t.equal("2023-10-01T19:32:08Z", meta.last_updated)

    # load dataclass (player) with nested dataclass (player.meta)
    player_1_json = '{"id":1,"alias":"jack","points":10522,"meta":' + meta_json + ',"__codec__":"dataclass"}'
    player_2_json = '{"id":2,"alias":"jill","points":25006,"meta":' + meta_json + ',"__codec__":"dataclass"}'

    player_1 = json.loads(player_1_json)
    t.equal(1, player_1.id)
    t.equal("jack", player_1.alias)
    t.equal(10522, player_1.points)
    t.equal(meta, player_1.meta)

    player_2 = json.loads(player_2_json)
    t.equal(2, player_2.id)
    t.equal("jill", player_2.alias)
    t.equal(25006, player_2.points)
    t.equal(meta, player_2.meta)

    # load json object (board) with nested dataclasses (board.players, board.meta)
    board_json = '{"type":"board","players": [' + player_1_json + "," + player_2_json + '],"meta":' + meta_json + "}"
    board = json.loads(board_json)
    t.equal("board", board["type"])

    t.equal(player_1, board["players"][0])
    t.equal(player_2, board["players"][1])
    t.equal(meta, board["meta"])
