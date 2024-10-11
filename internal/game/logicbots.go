package game

import (
	"fmt"
	"math/rand"
	"socketio4/internal/models"

	socketio_v5_client "github.com/maldikhan/go.socket.io/socket.io/v5/client"
	"github.com/maldikhan/go.socket.io/socket.io/v5/client/emit"
)

type Manager struct {
	gameManager *models.Game
}

func NewManager(gm *models.Game) *Manager {
	return &Manager{
		gameManager: gm,
	}
}

const (
	COLUMNS = 7
	CONNECT = 4
	EMPTY   = -1
	ENEMY   = 0
	BOT     = 1
)

func (m *Manager) MakeTurn(client *socketio_v5_client.Client, game models.Game) {
	board := game.Game.Board
	columnId := getBestMove(board)

	turn := models.Turn{
		ColumnID:      columnId,
		GameSessionID: game.ID,
	}

	client.Emit("turn", turn, emit.WithAck(func(response interface{}) {
		fmt.Printf("Turn response: %v\n", response)
	}))
}

func getBestMove(board [][]int) int {
	for col := 0; col < COLUMNS; col++ {
		if isValidMove(board, col) {
			row := len(board[col])
			board[col] = append(board[col], BOT)
			if checkWin(board, row, col, BOT) {
				board[col] = board[col][:len(board[col])-1]
				return col
			}
			board[col] = board[col][:len(board[col])-1]
		}
	}

	for col := 0; col < COLUMNS; col++ {
		if isValidMove(board, col) {
			row := len(board[col])
			board[col] = append(board[col], ENEMY)
			if checkWin(board, row, col, ENEMY) {
				board[col] = board[col][:len(board[col])-1]
				return col
			}
			board[col] = board[col][:len(board[col])-1]
		}
	}

	if isValidMove(board, 3) {
		return 3
	}

	validMoves := []int{}
	for col := 0; col < COLUMNS; col++ {
		if isValidMove(board, col) {
			validMoves = append(validMoves, col)
		}
	}
	if len(validMoves) > 0 {
		return validMoves[rand.Intn(len(validMoves))]
	}

	return -1
}

func isValidMove(board [][]int, col int) bool {
	return col >= 0 && col < COLUMNS && len(board[col]) < 7
}

func checkWin(board [][]int, row, col, player int) bool {

	// Вертикаль
	if len(board[col]) >= CONNECT {
		count := 0
		for i := len(board[col]) - 1; i >= 0; i-- {
			if board[col][i] == player {
				count++
				if count == CONNECT {
					return true
				}
			} else {
				break
			}
		}
	}

	// Горизонталь
	count := 0
	for c := 0; c < COLUMNS; c++ {
		if len(board[c]) > row && board[c][row] == player {
			count++
			if count == CONNECT {
				return true
			}
		} else {
			count = 0
		}
	}

	// Диагональ (слева направо)
	count = 0
	for r, c := row-min(row, col), col-min(row, col); c < COLUMNS; r, c = r+1, c+1 {
		if r < 0 || c < 0 || len(board[c]) <= r {
			break
		}
		if board[c][r] == player {
			count++
			if count == CONNECT {
				return true
			}
		} else {
			count = 0
		}
	}

	// Диагональ (справа налево)
	count = 0
	for r, c := row-min(row, COLUMNS-1-col), col+min(row, COLUMNS-1-col); c >= 0; r, c = r+1, c-1 {
		if r < 0 || c < 0 || len(board[c]) <= r {
			break
		}
		if board[c][r] == player {
			count++
			if count == CONNECT {
				return true
			}
		} else {
			count = 0
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
