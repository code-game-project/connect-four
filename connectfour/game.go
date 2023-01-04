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

	running bool

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

	game.grid = make([][]Cell, config.Height)
	for row := range game.grid {
		game.grid[row] = make([]Cell, config.Width)
		for col := range game.grid[row] {
			game.grid[row][col] = Cell{
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
	if player.SocketCount() == 1 {
		return
	}

	socket.Send(StartEvent, StartEventData{
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
	g.running = true
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

func (g *Game) dropToken(player *cg.Player, data DropTokenCmdData) {
	if (g.yellowPlayer == player && g.currentTurn != ColorYellow) || (g.redPlayer == player && g.currentTurn != ColorRed) {
		player.Send(InvalidActionEvent, InvalidActionEventData{
			Message: "It is not your turn.",
		})
		return
	}

	err := g.dropInColumn(g.currentTurn, data.Column)
	if err != nil {
		player.Send(InvalidActionEvent, InvalidActionEventData{
			Message: err.Error(),
		})
		return
	}

	g.sendBoard()

	if !g.checkDone() {
		g.turn()
	}
}

func (g *Game) dropInColumn(color Color, column int) error {
	if column < 0 || column >= g.config.Width {
		return fmt.Errorf("Column out of range. The grid only consists of %d columns.", g.config.Width)
	}

	for row := g.config.Height - 1; row >= 0; row-- {
		if g.grid[row][column].Color == ColorNone {
			g.grid[row][column].Color = color
			return nil
		}
	}

	return fmt.Errorf("Column %d is already full.", column)
}

func (g *Game) checkDone() bool {
	for row := range g.grid {
		for col := range g.grid[row] {
			// horizontal
			if g.checkLine(row, col, 0, 1) {
				return true
			}

			// vertical
			if g.checkLine(row, col, 1, 0) {
				return true
			}

			// diagonals

			// right, down
			if g.checkLine(row, col, 1, 1) {
				return true
			}

			// right, up
			if g.checkLine(row, col, -1, 1) {
				return true
			}
		}
	}

	return false
}

func (g *Game) checkLine(row, col, dr, dc int) bool {
	color := g.grid[row][col].Color
	if color == ColorNone {
		return false
	}

	if row+3*dr < 0 || row+3*dr >= g.config.Height || col+3*dc < 0 || col+3*dc >= g.config.Width {
		return false
	}

	if g.grid[row+1*dr][col+1*dc].Color != color || g.grid[row+2*dr][col+2*dc].Color != color || g.grid[row+3*dr][col+3*dc].Color != color {
		return false
	}

	g.gameOver(g.grid[row][col], g.grid[row+1*dr][col+1*dc], g.grid[row+2*dr][col+2*dc], g.grid[row+3*dr][col+3*dc])
	return true
}

func (g *Game) gameOver(cells ...Cell) {
	g.cg.Send(GameOverEvent, GameOverEventData{
		WinnerColor: g.currentTurn,
		WinningLine: cells,
	})
}

func (g *Game) handleCommand(origin *cg.Player, cmd cg.Command) {
	if !g.running {
		origin.Send(InvalidActionEvent, InvalidActionEventData{
			Message: "The game is not running.",
		})
		return
	}
	switch cmd.Name {
	case DropTokenCmd:
		var data DropTokenCmdData
		err := cmd.UnmarshalData(&data)
		if err != nil {
			origin.Log.ErrorData(cmd, "invalid command data")
			return
		}
		g.dropToken(origin, data)
	default:
		origin.Log.ErrorData(cmd, fmt.Sprintf("unexpected command: %s", cmd.Name))
	}
}
