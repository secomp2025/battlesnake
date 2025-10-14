package game

import (
	stdlog "log"
	"os"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/board"
	log "github.com/spf13/jwalterweatherman"
)

var verbose bool

type Snake struct {
	Name string
	URL  string
}

func CreateGame(snakes []Snake) (string, *BoardServer) {
	verbose = true

	gameState := &GameState{
		Width:           11,
		Height:          11,
		Names:           make([]string, len(snakes)),
		URLs:            make([]string, len(snakes)),
		Timeout:         500,
		GameType:        "standard",
		MapName:         "standard",
		Seed:            time.Now().UTC().UnixNano(),
		TurnDelay:       0,
		TurnDuration:    0,
		FoodSpawnChance: 15,
	}

	for i, snake := range snakes {
		gameState.Names[i] = snake.Name
		gameState.URLs[i] = snake.URL
	}

	if err := gameState.Initialize(); err != nil {
		log.ERROR.Fatalf("Error initializing game: %v", err)
	}

	boardGame := board.Game{
		ID:     gameState.gameID,
		Status: "running",
		Width:  gameState.Width,
		Height: gameState.Height,
		Ruleset: map[string]string{
			rules.ParamGameType: gameState.GameType,
		},
		RulesetName: gameState.GameType,
		RulesStages: []string{},
		Map:         gameState.MapName,
	}
	boardServer := NewBoardServer(boardGame)

	go func() {
		defer boardServer.Shutdown()
		if err := gameState.Run(boardGame, boardServer); err != nil {
			log.ERROR.Fatalf("Error running game: %v", err)
		}
	}()

	return gameState.gameID, boardServer
}

func init() {
	initConfig()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Setup logging
	log.SetStdoutOutput(os.Stderr)
	log.SetFlags(stdlog.Ltime | stdlog.Lmicroseconds)
	if verbose {
		log.SetStdoutThreshold(log.LevelDebug)
	} else {
		log.SetStdoutThreshold(log.LevelDebug)
	}
}
