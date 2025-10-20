#define _GNU_SOURCE
#include "battlesnake.h"
#include <microhttpd.h>
#include <jansson.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Weak function declarations — expected to be provided by LD_PRELOAD
__attribute__((weak)) SnakeInfo info(void);
__attribute__((weak)) void gameStart(GameState *state);
__attribute__((weak)) MoveResult move(GameState *state);
__attribute__((weak)) void gameEnd(GameState *state);


static Coord parse_coord(json_t *obj) {
    Coord c = {0};
    c.x = json_integer_value(json_object_get(obj, "x"));
    c.y = json_integer_value(json_object_get(obj, "y"));
    return c;
}

static char *dup_json_string(json_t *obj, const char *key) {
    json_t *val = json_object_get(obj, key);
    if (!val) return NULL;
    const char *str = json_string_value(val);
    return str ? strdup(str) : NULL;
}

static Snake parse_snake(json_t *snake_json) {
    Snake s = {0};

    s.id = dup_json_string(snake_json, "id");
    s.name = dup_json_string(snake_json, "name");
    s.health = json_integer_value(json_object_get(snake_json, "health"));
    s.length = json_integer_value(json_object_get(snake_json, "length"));

    json_t *head = json_object_get(snake_json, "head");
    s.head = parse_coord(head);

    json_t *body = json_object_get(snake_json, "body");
    s.length = json_array_size(body);
    s.body = calloc(s.length, sizeof(Coord));
    if (s.body)
        for (size_t i = 0; i < s.length; i++)
            s.body[i] = parse_coord(json_array_get(body, i));

    json_t *cust = json_object_get(snake_json, "customizations");
    if (cust) {
        s.color = dup_json_string(cust, "color");
        s.head_type = dup_json_string(cust, "head");
        s.tail_type = dup_json_string(cust, "tail");
    }

    return s;
}

GameState parse_game_state(json_t *root) {
    GameState gs = {0};
    gs.turn = json_integer_value(json_object_get(root, "turn"));

    json_t *board = json_object_get(root, "board");
    if (board) {
        gs.width = json_integer_value(json_object_get(board, "width"));
        gs.height = json_integer_value(json_object_get(board, "height"));

        json_t *food = json_object_get(board, "food");
        if (food) {
            gs.food_count = json_array_size(food);
            gs.food = calloc(gs.food_count, sizeof(Coord));
            for (size_t i = 0; i < gs.food_count; i++)
                gs.food[i] = parse_coord(json_array_get(food, i));
        }

        json_t *haz = json_object_get(board, "hazards");
        if (haz) {
            gs.hazard_count = json_array_size(haz);
            gs.hazards = calloc(gs.hazard_count, sizeof(Coord));
            for (size_t i = 0; i < gs.hazard_count; i++)
                gs.hazards[i] = parse_coord(json_array_get(haz, i));
        }

        json_t *snakes = json_object_get(board, "snakes");
        if (snakes) {
            gs.snake_count = json_array_size(snakes);
            gs.snakes = calloc(gs.snake_count, sizeof(Snake));
            if (gs.snakes)
                for (size_t i = 0; i < gs.snake_count; i++)
                    gs.snakes[i] = parse_snake(json_array_get(snakes, i));
        }
    }

    json_t *you = json_object_get(root, "you");
    gs.you = parse_snake(you);

    return gs;
}

void free_game_state(GameState *gs) {
    for (size_t i = 0; i < gs->snake_count; i++) {
        if (gs->snakes[i].id) free(gs->snakes[i].id);
        if (gs->snakes[i].name) free(gs->snakes[i].name);
        if (gs->snakes[i].body) free(gs->snakes[i].body);
        if (gs->snakes[i].color) free(gs->snakes[i].color);
        if (gs->snakes[i].head_type) free(gs->snakes[i].head_type);
        if (gs->snakes[i].tail_type) free(gs->snakes[i].tail_type);
    }
    if (gs->snakes) free(gs->snakes);

    if (gs->food) free(gs->food);
    if (gs->hazards) free(gs->hazards);

    if (gs->you.id) free(gs->you.id);
    if (gs->you.name) free(gs->you.name);
    if (gs->you.body) free(gs->you.body);
    if (gs->you.color) free(gs->you.color);
    if (gs->you.head_type) free(gs->you.head_type);
    if (gs->you.tail_type) free(gs->you.tail_type);
}


struct ConnectionInfo {
    char *data;
    size_t size;
};

static enum MHD_Result request_handler(void *cls, struct MHD_Connection *connection,
                           const char *url, const char *method,
                           const char *version, const char *upload_data,
                           size_t *upload_data_size, void **con_cls) {
    (void)cls;
    (void)version;
    struct ConnectionInfo *con_info = *con_cls;

    if (!con_info) {
        con_info = calloc(1, sizeof(*con_info));
        *con_cls = con_info;
        return MHD_YES;
    }

    if (*upload_data_size > 0) {
        con_info->data = realloc(con_info->data, con_info->size + *upload_data_size + 1);
        memcpy(con_info->data + con_info->size, upload_data, *upload_data_size);
        con_info->size += *upload_data_size;
        con_info->data[con_info->size] = '\0';
        *upload_data_size = 0;
        return MHD_YES;
    }

    const char *response_str = "{\"error\": \"no handler\"}";
    json_t *json_body = NULL;


    if (strcmp(method, "GET") == 0 && strcmp(url, "/") == 0) {
        SnakeInfo info_struct = info();
        json_t *info_json = json_object();
        json_object_set_new(info_json, "color", json_string(info_struct.color));
        json_object_set_new(info_json, "head", json_string(info_struct.head));
        json_object_set_new(info_json, "tail", json_string(info_struct.tail));
        response_str = json_dumps(info_json, 0);
        json_decref(info_json);
    }
    else if (strcmp(method, "POST") == 0 && con_info->data && con_info->size > 0) {
        json_error_t error;
        json_body = json_loads(con_info->data, 0, &error);
        if (!json_body) {
            response_str = "{\"error\":\"invalid json\"}";
        } else if (strcmp(url, "/start") == 0) {
            GameState gs = parse_game_state(json_body);
            gameStart(&gs);
            free_game_state(&gs);
            response_str = "{\"status\":\"ok\"}";
        } else if (strcmp(url, "/move") == 0) {
            GameState gs = parse_game_state(json_body);
            MoveResult res = move(&gs);
            json_t *move_json = json_object();
            json_object_set_new(move_json, "move", json_string(res.move));
            if(res.taunt) json_object_set_new(move_json, "taunt", json_string(res.taunt));
            response_str = json_dumps(move_json, 0);
            json_decref(move_json);
            free_game_state(&gs);
        } else if (strcmp(url, "/end") == 0) {
            printf("[END] Dump: %s\n", json_dumps(json_body, 0));
            GameState gs = parse_game_state(json_body);
            gameEnd(&gs);
            free_game_state(&gs);

            response_str = "{\"status\":\"ok\"}";
        }
    }

    struct MHD_Response *response = MHD_create_response_from_buffer(strlen(response_str),
                                                                    (void*)response_str,
                                                                    MHD_RESPMEM_MUST_COPY);
    MHD_add_response_header(response, "Content-Type", "application/json");
    MHD_add_response_header(response, "Server", "battlesnake/c");
    enum MHD_Result ret = MHD_queue_response(connection, MHD_HTTP_OK, response);
    MHD_destroy_response(response);

    free(con_info->data);
    con_info->data = NULL;
    con_info->size = 0;
    json_decref(json_body);

    return ret;
}

int main(int argc, char **argv) {
    setvbuf(stdout, NULL, _IOLBF, 0);
    setvbuf(stderr, NULL, _IOLBF, 0);

    if (argc < 2) {
        fprintf(stderr, "Usage: %s <port>\n", argv[0]);
        return 1;
    }
    
    struct MHD_Daemon *daemon;

    int port = atoi(argv[1]);

    daemon = MHD_start_daemon(MHD_USE_SELECT_INTERNALLY, port, NULL, NULL,
                              &request_handler, NULL, MHD_OPTION_END);
    if (daemon == NULL) return 1;

    printf("Battlesnake server running on port %d\n", port);
    // getchar(); // keep running until Enter pressed

    while (1) {
        sleep(1);
    }

    MHD_stop_daemon(daemon);
    return 0;
}
