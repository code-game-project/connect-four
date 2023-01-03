package connectfour

import (
	"fmt"
	"math/rand"

	"github.com/code-game-project/go-server/cg"
)

type Game struct {
	cg     *cg.Game
	config GameConfig

	yellowPlayer *cg.Player
	redPlayer    *cg.Player
	currentTurn  Color

	grid [][]Cell
}

func NewGame(cgGame *cg.Game, config GameConfig) *Game {
	game := &Game{
		cg:          cgGame,
		config:      config,
		currentTurn: ColorRed, // Switched before the game starts. The first token is dropped by yellow.
	}

	cgGame.OnPlayerLeft = func(_ *cg.Player) {
		game.cg.Close()
	}
	cgGame.OnPlayerJoined = game.onPlayerJoined
	cgGame.OnPlayerSocketConnected = game.onPlayerSocketConnected

	game.grid = make([][]Cell, config.Width)
	for col := range game.grid {
		game.grid[col] = make([]Cell, config.Height)
		for row := range game.grid[col] {
			game.grid[col][row] = Cell{
				Row:    row,
				Column: col,
				Color:  ColorNone,
			}
		}
	}

	return game
}

func (g *Game) onPlayerSocketConnected(player *cg.Player, socket *cg.GameSocket) {
	if g.yellowPlayer == nil || g.redPlayer == nil {
		return
	}

	g.cg.Send(StartEvent, StartEventData{
		Colors: map[string]Color{
			g.yellowPlayer.Id: ColorYellow,
			g.redPlayer.Id:    ColorRed,
		},
	})

	socket.Send(GridEvent, GridEventData{
		Cells: g.grid,
	})

	socket.Send(TurnEvent, TurnEventData{
		Color: g.currentTurn,
	})
}

func (g *Game) onPlayerJoined(player *cg.Player) {
	if g.yellowPlayer != nil {
		g.redPlayer = player
	} else if g.redPlayer != nil {
		g.yellowPlayer = player
	} else {
		if rand.Intn(2) == 1 {
			g.yellowPlayer = player
		} else {
			g.redPlayer = player
		}
	}

	if g.yellowPlayer != nil && g.redPlayer != nil {
		g.start()
	}
}

func (g *Game) start() {
	g.cg.Send(StartEvent, StartEventData{
		Colors: map[string]Color{
			g.yellowPlayer.Id: ColorYellow,
			g.redPlayer.Id:    ColorRed,
		},
	})
	g.sendBoard()
	g.turn()
}

func (g *Game) turn() {
	if g.currentTurn == ColorYellow {
		g.currentTurn = ColorRed
	} else {
		g.currentTurn = ColorYellow
	}
	g.cg.Send(TurnEvent, TurnEventData{
		Color: g.currentTurn,
	})
}

func (g *Game) sendBoard() {
	g.cg.Send(GridEvent, GridEventData{
		Cells: g.grid,
	})
}

func (g *Game) Run() {
	for g.cg.Running() {
		cmd, ok := g.cg.WaitForNextCommand()
		if !ok {
			break
		}
		g.handleCommand(cmd.Origin, cmd.Cmd)
	}
}

func (g *Game) handleCommand(origin *cg.Player, cmd cg.Command) {
	switch cmd.Name {
	default:
		origin.Log.ErrorData(cmd, fmt.Sprintf("unexpected command: %s", cmd.Name))
	}
}