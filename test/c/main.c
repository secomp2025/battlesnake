#include "battlesnake.h"
#include <stdio.h>

SnakeInfo info(void) {
    return (SnakeInfo) {
        .color = "#754927",
        .head = "alligator",
        .tail = "rocket"
    };
}

void gameStart(GameState *state) {
    printf("[START] Turn: %d | You: %s\n", state->turn, state->you.name);
}
void gameEnd(GameState *state) {
    printf("[END] Game finished\n");
}

MoveResult move(GameState *state) {
    MoveResult res = {0};

    printf("[MOVE] Turn %d, Health %d\n", state->turn, state->you.health);

    res.move = "up";
    return res;
}

