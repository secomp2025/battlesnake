package controllers

import (
	"sync"

	"github.com/secomp2025/localsnake/game"
)

type gameManager struct {
	lock  sync.RWMutex
	games map[string]GameInfo
}

type GameInfo struct {
	ID     string
	Server *game.BoardServer
	Snakes []game.Snake
}

type GameController struct {
	manager *gameManager
}

var globalGameManager *gameManager

func init() {
	globalGameManager = &gameManager{
		games: make(map[string]GameInfo),
	}
}

func NewGameController() *GameController {
	return &GameController{
		manager: globalGameManager,
	}
}

func (c *GameController) CreateGame(snakes []game.Snake) GameInfo {
	c.manager.lock.Lock()
	defer c.manager.lock.Unlock()

	gameID, boardServer := game.CreateGame(snakes)
	// go func() {
	// 	defer boardServer.Shutdown()
	// }()

	gameInfo := GameInfo{
		ID:     gameID,
		Server: boardServer,
		Snakes: snakes,
	}
	c.manager.games[gameID] = gameInfo

	return gameInfo
}
