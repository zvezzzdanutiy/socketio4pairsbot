package main

import (
	"context"
	"fmt"
	"math/rand"
	"socketio4/internal/network"
	"time"

	socketio_v5_client "github.com/maldikhan/go.socket.io/socket.io/v5/client"
	"github.com/maldikhan/go.socket.io/socket.io/v5/client/emit"
	"github.com/maldikhan/go.socket.io/utils"
)

type Participant struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Credits  int    `json:"credits"`
	IsActive bool   `json:"isActive"`
	Score    int    `json:"score"`
}

type GameProperties struct {
	Columns  int `json:"columns"`
	MaxChips int `json:"maxChips"`
}

type GameDetails struct {
	Properties GameProperties `json:"properties"`
	Board      [][]int        `json:"board"`
	Stroke     [][]int        `json:"stroke"`
	Type       string         `json:"type"`
}

type Game struct {
	ID                string        `json:"id"`
	Participants      []Participant `json:"participants"`
	MaxParticipants   int           `json:"maxParticipants"`
	State             string        `json:"state"`
	LastTurnTimestamp int64         `json:"lastTurnTimestamp"`
	TurnTimeout       int           `json:"turnTimeout"`
	Game              GameDetails   `json:"game"`
}

type Turn struct {
	ColumnID      int    `json:"columnId"`
	GameSessionID string `json:"gameSessionId"`
}

func main() {
	ctx := context.Background()
	token, err := network.LoginAndGetToken("bot123", "pass123")
	if err != nil {
		fmt.Println(err)
	}
	client, err := socketio_v5_client.NewClient(
		socketio_v5_client.WithLogger(&utils.DefaultLogger{Level: utils.INFO}),
		socketio_v5_client.WithRawURL("https://bun-gs-bungamecenter.amvera.io/ws/game/"),
	)

	if err != nil {
		panic(err)
	}
	client.SetHandshakeData(map[string]interface{}{"token": token})
	gameIdChannel := make(chan string, 1)

	client.On("error", func(err string) {
		fmt.Println("ERROR", err)
	})

	client.On("jgs", func(gameId string) {
		fmt.Println("Server required me to join the game", gameId)
		gameIdChannel <- gameId
	})

	client.On("gs", func(game Game) {
		fmt.Println("New Game State", game)
		if game.State == "FINISHED" {
			gameIdChannel <- ""
			fmt.Errorf("ИГРА ЗАКОНЧЕНА", game)
		} else if game.State == "GAME" {
			for _, participant := range game.Participants {
				if participant.IsActive {
					makeTurn(client, game)
					break
				}
			}
		}
	})

	client.Connect(context.Background())
	activeGame := ""
	ticker := time.NewTicker(time.Millisecond * 500)

	for {
		select {
		case gameId := <-gameIdChannel:
			activeGame = gameId
			if gameId == "" {
				continue
			}
			client.Emit("joinGameSession", gameId)
		case <-ticker.C:
			if activeGame != "" {
				continue
			}
			fmt.Println("No game received, find the new One")
			client.Emit("searchGameSession", "firdrop", emit.WithAck(func(gameId string) {
				fmt.Println("Found new game", gameId)
				gameIdChannel <- gameId
			}))
		case <-ctx.Done():
			return
		}
	}
}

const (
	COLUMNS = 7
	CONNECT = 4
	EMPTY   = -1
	ENEMY   = 0
	BOT     = 1
)

func makeTurn(client *socketio_v5_client.Client, game Game) {
	board := game.Game.Board
	columnId := getBestMove(board)

	turn := Turn{
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
