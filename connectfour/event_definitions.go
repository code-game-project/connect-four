package connectfour

import "github.com/code-game-project/go-server/cg"

type GameConfig struct {
	// The width of the game grid. min = 3, default = 7
	Width int `json:"width"`
	// The height of the game grid. min = 3, default = 6
	Height int `json:"height"`
	// The number of discs, which form a winning line. min = 2, default = 4
	WinLength int `json:"win_length"`
	// The rule variation to use. default: original
	Variation Variation `json:"variation"`
}

// The `drop_disc` command can be sent to drop a disc into the game grid. Only allowed when it is the current player's turn.
const DropDiscCmd cg.CommandName = "drop_disc"

type DropDiscCmdData struct {
	// 0 <= column < config.width
	Column int `json:"column"`
}

// The `pop_out` command can be sent to remove a disc of your color from the bottom of the grid. Only available if config.variation = pop_out.
const PopOutCmd cg.CommandName = "pop_out"

type PopOutCmdData struct {
	// 0 <= column < config.width
	Column int `json:"column"`
}

// The `start` event is sent to all players when the game begins.
const StartEvent cg.EventName = "start"

type StartEventData struct {
	// A map of player IDs mapped to their respective disc colors.
	Colors map[string]Color `json:"colors"`
}

// The game grid.
const GridEvent cg.EventName = "grid"

type GridEventData struct {
	// The cells of the grid as columns (left to right) inside of rows (top to bottom).
	Cells [][]Cell `json:"cells"`
}

// The 'invalid_action' event notifies the player that their action was not allowed.
const InvalidActionEvent cg.EventName = "invalid_action"

type InvalidActionEventData struct {
	// The message containing details on what the player did wrong.
	Message string `json:"message"`
}

// The `turn` event is sent to all players when it is the next player's turn.
const TurnEvent cg.EventName = "turn"

type TurnEventData struct {
	// The sign of the player whose turn it is now.
	Color Color `json:"color"`
}

const GameOverEvent cg.EventName = "game_over"

type GameOverEventData struct {
	// The color of the winner.
	WinnerColor Color `json:"winner_color"`
	// The four cells which form a line.
	WinningLine []Cell `json:"winning_line"`
}

// A cell on the game grid.
type Cell struct {
	Row    int   `json:"row"`
	Column int   `json:"column"`
	Color  Color `json:"color"`
}

// A disc color.
type Color string

const (
	// No color. Used for empty cells.
	ColorNone Color = "none"
	// A drops the first disc.
	ColorA Color = "a"
	// B drops the second disc.
	ColorB Color = "b"
)

type Variation string

const (
	// The original Connect 4 game.
	VariationOriginal Variation = "original"
	// Instead of dropping a disc into a grid a player may choose to remove a disc of their own color from the bottom.
	VariationPopOut Variation = "pop_out"
)
