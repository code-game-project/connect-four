package connectfour

import (
	"fmt"
	"math/rand"

	"github.com/code-game-project/go-server/cg"
)

type Game struct {
	cg     *cg.Game
	config GameConfig

	playerA     *cg.Player
	playerB     *cg.Player
	currentTurn Color

	running bool

	grid [][]Cell
}

func NewGame(cgGame *cg.Game, config GameConfig) *Game {
	game := &Game{
		cg:          cgGame,
		config:      config,
		currentTurn: ColorB, // Switched before the game starts. The first disc is dropped by A.
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
	if g.playerA == nil || g.playerB == nil {
		return
	}
	if player.SocketCount() == 1 {
		return
	}

	socket.Send(StartEvent, StartEventData{
		Colors: map[string]Color{
			g.playerA.Id: ColorA,
			g.playerB.Id: ColorB,
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
	if g.playerA != nil {
		g.playerB = player
	} else if g.playerB != nil {
		g.playerA = player
	} else {
		if rand.Intn(2) == 1 {
			g.playerA = player
		} else {
			g.playerB = player
		}
	}

	if g.playerA != nil && g.playerB != nil {
		g.start()
	}
}

func (g *Game) start() {
	g.cg.Send(StartEvent, StartEventData{
		Colors: map[string]Color{
			g.playerA.Id: ColorA,
			g.playerB.Id: ColorB,
		},
	})
	g.running = true
	g.sendBoard()
	g.turn()
}

func (g *Game) turn() {
	if g.currentTurn == ColorA {
		g.currentTurn = ColorB
	} else {
		g.currentTurn = ColorA
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

func (g *Game) dropDisc(player *cg.Player, data DropDiscCmdData) {
	if (g.playerA == player && g.currentTurn != ColorA) || (g.playerB == player && g.currentTurn != ColorB) {
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

func (g *Game) popOutDisc(player *cg.Player, data PopOutCmdData) {
	if g.config.Variation != VariationPopOut {
		player.Send(InvalidActionEvent, InvalidActionEventData{
			Message: fmt.Sprintf("`pop_out` is not allowed in rule variation `%s`.", g.config.Variation),
		})
		return
	}

	if (g.playerA == player && g.currentTurn != ColorA) || (g.playerB == player && g.currentTurn != ColorB) {
		player.Send(InvalidActionEvent, InvalidActionEventData{
			Message: "It is not your turn.",
		})
		return
	}

	err := g.popOutInColumn(data.Column)
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

func (g *Game) popOutInColumn(column int) error {
	if column < 0 || column >= g.config.Width {
		return fmt.Errorf("Column out of range. The grid only consists of %d columns.", g.config.Width)
	}

	if g.grid[g.config.Height-1][column].Color != g.currentTurn {
		return fmt.Errorf("You can only pop out your own discs.")
	}

	for row := g.config.Height - 1; row > 0; row-- {
		g.grid[row][column].Color = g.grid[row-1][column].Color
	}

	g.grid[0][column].Color = ColorNone
	return nil
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

	if row+(g.config.WinLength-1)*dr < 0 || row+(g.config.WinLength-1)*dr >= g.config.Height || col+(g.config.WinLength-1)*dc < 0 || col+(g.config.WinLength-1)*dc >= g.config.Width {
		return false
	}

	winningLine := make([]Cell, g.config.WinLength)
	winningLine[0] = g.grid[row][col]

	for i := 1; i < g.config.WinLength; i++ {
		if g.grid[row+i*dr][col+i*dc].Color != color {
			return false
		}
		winningLine[i] = g.grid[row+i*dr][col+i*dc]
	}

	g.gameOver(winningLine)
	return true
}

func (g *Game) gameOver(cells []Cell) {
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
	case DropDiscCmd:
		var data DropDiscCmdData
		err := cmd.UnmarshalData(&data)
		if err != nil {
			origin.Log.ErrorData(cmd, "invalid command data")
			return
		}
		g.dropDisc(origin, data)
	case PopOutCmd:
		var data PopOutCmdData
		err := cmd.UnmarshalData(&data)
		if err != nil {
			origin.Log.ErrorData(cmd, "invalid command data")
			return
		}
		g.popOutDisc(origin, data)
	default:
		origin.Log.ErrorData(cmd, fmt.Sprintf("unexpected command: %s", cmd.Name))
	}
}
