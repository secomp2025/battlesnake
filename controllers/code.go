package controllers

import (
	"context"
	"database/sql"

	"github.com/secomp2025/localsnake/database"
)

type CodeController struct {
	queries *database.Queries
}

func NewCodeController(db database.DBTX) *CodeController {
	return &CodeController{queries: database.New(db)}
}

func (c *CodeController) GetCode(ctx context.Context, id int64) (*database.Code, error) {
	code_model, err := c.queries.GetCode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &code_model, nil
}

func (c *CodeController) FindCode(ctx context.Context, code string) (*database.Code, error) {
	code_model, err := c.queries.FindCode(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &code_model, nil
}

func (c *CodeController) ListCodes(ctx context.Context) ([]database.Code, error) {
	codes, err := c.queries.ListCodes(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return codes, nil
}
