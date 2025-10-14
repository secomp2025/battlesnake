package controllers

import (
	"context"

	"github.com/secomp2025/localsnake/database"
)

type SnakeController struct {
	queries *database.Queries
}

func NewSnakeController(db database.DBTX) SnakeController {
	return SnakeController{queries: database.New(db)}
}

func (c *SnakeController) ListSnakes(ctx context.Context) ([]database.Snake, error) {
	return c.queries.ListSnakes(ctx)
}

func (c *SnakeController) GetSnake(ctx context.Context, id int64) (*database.Snake, error) {
	snake_model, err := c.queries.GetSnake(ctx, id)
	if err != nil {
		return nil, err
	}
	return &snake_model, nil
}

func (c *SnakeController) CreateSnake(ctx context.Context, snake *database.Snake) (*database.Snake, error) {
	snake_model, err := c.queries.CreateSnake(ctx, database.CreateSnakeParams{Path: snake.Path, Lang: snake.Lang, TeamID: snake.TeamID})
	if err != nil {
		return nil, err
	}
	return &snake_model, nil
}

func (c *SnakeController) UpdateSnake(ctx context.Context, snake *database.Snake) (*database.Snake, error) {
	err := c.queries.UpdateSnake(ctx, database.UpdateSnakeParams{ID: snake.ID, Path: snake.Path, Lang: snake.Lang})
	if err != nil {
		return nil, err
	}
	return snake, nil
}

func (c *SnakeController) DeleteSnake(ctx context.Context, id int64) error {
	return c.queries.DeleteSnake(ctx, id)
}

func (c *SnakeController) ListTeamSnakes(ctx context.Context, team_id int64) ([]database.Snake, error) {
	return c.queries.ListTeamSnakes(ctx, team_id)
}
