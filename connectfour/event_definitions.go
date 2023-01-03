package connectfour

import "github.com/code-game-project/go-server/cg"

type GameConfig struct {
	// The width of the game grid. default = 7
	Width int `json:"width"`
	// The height of the game grid. default = 6
	Height int `json:"height"`
}

// The `drop_token` command can be sent to drop a token into the game grid. Only allowed when it is the current player's turn.
const DropTokenCmd cg.CommandName = "drop_token"

type DropTokenCmdData struct {
	// 0 <= column < config.width
	Column int `json:"column"`
}

// The `start` event is sent to all players when the game begins.
const StartEvent cg.EventName = "start"

type StartEventData struct {
	// A map of player IDs mapped to their respective token colors.
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

// A token color.
type Color string

const (
	// No color. Used for empty cells.
	ColorNone Color = "none"
	// Yellow drops the first token.
	ColorYellow Color = "yellow"
	// Red drops the second token.
	ColorRed Color = "red"
)
