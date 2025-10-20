#include "battlesnake.h"
#include <stdio.h>
#include <stdlib.h>

// info é chamado quando você cria seu Battlesnake em play.battlesnake.com
// e controla a aparência do seu Battlesnake
SnakeInfo info(void) {
    return (SnakeInfo) {
        .color = "#F30303",  // Escolha uma cor
        .head = "default",   // Escolha uma cabeça
        .tail = "default"    // Escolha uma cauda
    };
}

// start é chamado quando seu Battlesnake inicia uma partida
void gameStart(GameState *state) {
    printf("INÍCIO DO JOGO\n");
}

// end é chamado quando seu Battlesnake termina uma partida
void gameEnd(GameState *state) {
    printf("FIM DO JOGO\n");
}

// move é chamado a cada turno e retorna o seu próximo movimento
// Movimentos válidos são "up", "down", "left" ou "right"
MoveResult move(GameState *state) {
    MoveResult result = {0};

    int is_move_safe_up = 1;
    int is_move_safe_down = 1;
    int is_move_safe_left = 1;
    int is_move_safe_right = 1;

    // Incluímos código para impedir que sua Snake se mova para trás
    Coord my_head = state->you.body[0];  // Coordenadas da cabeça
    Coord my_neck = state->you.body[1];  // Coordenadas do "pescoço"

    if (my_neck.x < my_head.x) {          // Pescoço está à esquerda da cabeça
        is_move_safe_left = 0;
    } else if (my_neck.x > my_head.x) {   // Pescoço está à direita
        is_move_safe_right = 0;
    } else if (my_neck.y < my_head.y) {   // Pescoço está abaixo
        is_move_safe_down = 0;
    } else if (my_neck.y > my_head.y) {   // Pescoço está acima
        is_move_safe_up = 0;
    }

    // TODO: Etapa 1 - Impedir que o Battlesnake saia dos limites do tabuleiro
    // int board_width = state->width;
    // int board_height = state->height;

    // TODO: Etapa 2 - Impedir que o Battlesnake colida com ele mesmo
    // Coord *my_body = state->you.body;
    // size_t my_length = state->you.length;

    // TODO: Etapa 3 - Impedir que o Battlesnake colida com outros Battlesnakes
    // Snake *opponents = state->snakes;
    // size_t opponent_count = state->snake_count;

    // Existem movimentos seguros disponíveis?
    const char *safe_moves[4];
    int safe_count = 0;

    if (is_move_safe_up)    safe_moves[safe_count++] = "up";
    if (is_move_safe_down)  safe_moves[safe_count++] = "down";
    if (is_move_safe_left)  safe_moves[safe_count++] = "left";
    if (is_move_safe_right) safe_moves[safe_count++] = "right";

    if (safe_count == 0) {
        printf("MOVIMENTO %d: Nenhum movimento seguro detectado! Indo para baixo\n", state->turn);
        result.move = "down";
        return result;
    }

    // Escolhe um movimento aleatório entre os seguros
    int idx = rand() % safe_count;
    const char *next_move = safe_moves[idx];

    // TODO: Etapa 4 - Ir em direção à comida em vez de aleatoriamente
    // Coord *food = state->food;
    // size_t food_count = state->food_count;

    printf("MOVIMENTO %d: %s\n", state->turn, next_move);
    result.move = (char *)next_move;
    return result;
}
