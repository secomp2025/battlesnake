import logging
import typing
import importlib.util
import sys

from flask import Flask
from flask import request


def run_server(handlers: typing.Dict, port: int):
    app = Flask("Battlesnake")

    @app.get("/")
    def on_info():
        info = handlers["info"]()

        info["author"] = "IFSP"
        info["apiversion"]: "1"

        return info

    @app.post("/start")
    def on_start():
        game_state = request.get_json()
        handlers["start"](game_state)
        return "ok"

    @app.post("/move")
    def on_move():
        game_state = request.get_json()
        return handlers["move"](game_state)

    @app.post("/end")
    def on_end():
        game_state = request.get_json()
        handlers["end"](game_state)
        return "ok"

    @app.after_request
    def identify_server(response):
        response.headers.set("server", "battlesnake/github/starter-snake-python")
        return response

    host = "0.0.0.0"

    logging.getLogger("werkzeug").setLevel(logging.ERROR)

    print(f"\nRunning Battlesnake at http://{host}:{port}")
    app.run(host=host, port=port)


def import_dynamic(name, module_path):
    spec = importlib.util.spec_from_file_location(name, module_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


if __name__ == "__main__":
    snake_path = sys.argv[1]
    port = int(sys.argv[2])

    snake = import_dynamic("snake", snake_path)

    run_server(
        {
            "info": snake.info,
            "start": snake.start,
            "move": snake.move,
            "end": snake.end,
        },
        port,
    )
