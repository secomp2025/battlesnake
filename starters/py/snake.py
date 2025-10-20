# Bem-vindo ao
# __________         __    __  .__                               __
# \______   \_____ _/  |__/  |_|  |   ____   ______ ____ _____  |  | __ ____
#  |    |  _/\__  \\   __\   __\  | _/ __ \ /  ___//    \\__  \ |  |/ // __ \
#  |    |   \ / __ \|  |  |  | |  |_\  ___/ \___ \|   |  \/ __ \|    <\  ___/
#  |________/(______/__|  |__| |____/\_____>______>___|__(______/__|__\\_____>
#
# Este arquivo pode ser um bom local para a lógica e funções auxiliares do seu Battlesnake.
#
# Para começar, incluímos um código para impedir que seu Battlesnake se mova para trás.
# Para mais informações, veja docs.battlesnake.com

import random
import typing


# info é chamado quando você cria seu Battlesnake em play.battlesnake.com
# e controla a aparência do seu Battlesnake
def info() -> typing.Dict:
    return {
        "color": "#F30303",  # Escolha uma cor
        "head": "default",  # Escolha uma cabeça
        "tail": "default",  # Escolha uma cauda
    }


# start é chamado quando seu Battlesnake inicia uma partida
def start(game_state: typing.Dict):
    print("INÍCIO DO JOGO")


# end é chamado quando seu Battlesnake termina uma partida
def end(game_state: typing.Dict):
    print("FIM DO JOGO")


# move é chamado a cada turno e retorna o seu próximo movimento
# Movimentos válidos são "up", "down", "left" ou "right"
# Acesse https://docs.battlesnake.com/api/example-move para ver os dados disponíveis
def move(game_state: typing.Dict) -> typing.Dict:
    is_move_safe = {"up": True, "down": True, "left": True, "right": True}

    # Incluímos código para impedir que sua Snake se mova para trás
    my_head = game_state["you"]["body"][0]  # Coordenadas da cabeça
    my_neck = game_state["you"]["body"][1]  # Coordenadas do "pescoço"

    if (
        my_neck["x"] < my_head["x"]
    ):  # Pescoço está à esquerda da cabeça, não vá para a esquerda
        is_move_safe["left"] = False

    elif (
        my_neck["x"] > my_head["x"]
    ):  # Pescoço está à direita da cabeça, não vá para a direita
        is_move_safe["right"] = False

    elif (
        my_neck["y"] < my_head["y"]
    ):  # Pescoço está abaixo da cabeça, não vá para baixo
        is_move_safe["down"] = False

    elif my_neck["y"] > my_head["y"]:  # Pescoço está acima da cabeça, não vá para cima
        is_move_safe["up"] = False

    # TODO: Etapa 1 - Impedir que o Battlesnake saia dos limites do tabuleiro
    # board_width = game_state['board']['width']
    # board_height = game_state['board']['height']

    # TODO: Etapa 2 - Impedir que o Battlesnake colida com ele mesmo
    # my_body = game_state['you']['body']

    # TODO: Etapa 3 - Impedir que o Battlesnake colida com outros Battlesnakes
    # opponents = game_state['board']['snakes']

    # Existem movimentos seguros disponíveis?
    safe_moves = []
    for move, isSafe in is_move_safe.items():
        if isSafe:
            safe_moves.append(move)

    if len(safe_moves) == 0:
        print(
            f"MOVIMENTO {game_state['turn']}: Nenhum movimento seguro detectado! Indo para baixo"
        )
        return {"move": "down"}

    # Escolhe um movimento aleatório entre os seguros
    next_move = random.choice(safe_moves)

    # TODO: Etapa 4 - Ir em direção à comida em vez de aleatoriamente, para recuperar saúde e
    # sobreviver mais
    # food = game_state['board']['food']

    print(f"MOVIMENTO {game_state['turn']}: {next_move}")
    return {"move": next_move}
