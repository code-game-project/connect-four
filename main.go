package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/code-game-project/go-server/cg"
	"github.com/spf13/pflag"

	"github.com/code-game-project/connect-four/connectfour"
)

func main() {
	rand.Seed(time.Now().UnixMilli())

	var port int
	pflag.IntVarP(&port, "port", "p", 0, "The network port of the game server.")
	pflag.Parse()

	if port == 0 {
		portStr, ok := os.LookupEnv("CG_PORT")
		if ok {
			port, _ = strconv.Atoi(portStr)
		}
	}

	if port == 0 {
		port = 8080
	}

	server := cg.NewServer("connect-four", cg.ServerConfig{
		DisplayName:             "Connect 4",
		Version:                 "0.1",
		Description:             "Drop colored tokens into a grid. You win when you manage to form a horizontal, vertical or diagonal line of four tokens.",
		RepositoryURL:           "https://github.com/code-game-project/connect-four",
		Port:                    port,
		CGEFilepath:             "events.cge",
		DeleteInactiveGameDelay: 1 * time.Hour,
		KickInactivePlayerDelay: 30 * time.Minute,
		MaxPlayersPerGame:       2,
	})

	server.Run(func(cgGame *cg.Game, config json.RawMessage) {
		var gameConfig connectfour.GameConfig
		err := json.Unmarshal(config, &gameConfig)
		if err != nil {
			cgGame.Log.Error("Failed to unmarshal game config: %s", err)
		}

		if gameConfig.Width < 3 {
			gameConfig.Width = 7
		}
		if gameConfig.Height < 3 {
			gameConfig.Height = 6
		}
		if gameConfig.WinLength < 2 {
			gameConfig.WinLength = 4
		}

		cgGame.SetConfig(gameConfig)

		connectfour.NewGame(cgGame, gameConfig).Run()
	})
}
