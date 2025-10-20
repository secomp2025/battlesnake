#pragma once
#include <jansson.h>
#include <stddef.h>

// -------------------- Data structures --------------------

typedef struct {
    char *color;
    char *head;
    char *tail;
} SnakeInfo;

typedef struct {
    int x;
    int y;
} Coord;

typedef struct {
    Coord *body;
    size_t length;
    char *id;
    char *name;
    int health;
    Coord head;
    char *color;
    char *head_type;
    char *tail_type;
} Snake;

typedef struct {
    int width;
    int height;
    Coord *food;
    size_t food_count;
    Coord *hazards;
    size_t hazard_count;
    Snake *snakes;
    size_t snake_count;
    Snake you;
    int turn;
} GameState;

typedef struct {
    char *move;
    char *taunt;
} MoveResult;



