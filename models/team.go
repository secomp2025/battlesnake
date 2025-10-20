package models

import "time"

type Team struct {
	ID   int64
	Name string
	Code string

	Snake *Snake
}

type Snake struct {
	ID        int64
	Lang      string
	UpdatedAt time.Time
	Status    string
}
