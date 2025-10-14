package controllers

import (
	"context"
	"database/sql"

	"github.com/secomp2025/localsnake/database"
)

type TeamController struct {
	queries *database.Queries
}

func NewTeamController(db database.DBTX) *TeamController {
	return &TeamController{queries: database.New(db)}
}

func (c *TeamController) GetTeamByCode(ctx context.Context, code_id int64) (*database.Team, error) {
	team_model, err := c.queries.GetTeamByCode(ctx, code_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &team_model, nil
}

func (c *TeamController) ListTeams(ctx context.Context) ([]database.Team, error) {
	teams, err := c.queries.ListTeams(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return teams, nil
}

func (c *TeamController) CreateTeam(ctx context.Context, name string, code_id int64) (*database.Team, error) {
	team_model, err := c.queries.CreateTeam(ctx, database.CreateTeamParams{Name: name, CodeID: code_id})
	if err != nil {
		return nil, err
	}
	return &team_model, nil
}

func (c *TeamController) UpdateTeam(ctx context.Context, team *database.Team) (*database.Team, error) {
	err := c.queries.UpdateTeam(ctx, database.UpdateTeamParams{ID: team.ID, Name: team.Name, CodeID: team.CodeID})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (c *TeamController) DeleteTeam(ctx context.Context, id int64) error {
	return c.queries.DeleteTeam(ctx, id)
}
