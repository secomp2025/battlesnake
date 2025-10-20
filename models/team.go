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

type Code struct {
	ID   int64
	Code string
	Used bool
}
