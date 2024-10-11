package main

import (
	"context"
	"fmt"
	"socketio4/internal/game"
	"socketio4/internal/models"
	"socketio4/internal/network"
	"time"

	socketio_v5_client "github.com/maldikhan/go.socket.io/socket.io/v5/client"
	"github.com/maldikhan/go.socket.io/socket.io/v5/client/emit"
	"github.com/maldikhan/go.socket.io/utils"
)

func main() {
	ctx := context.Background()
	GameManagerCreate := models.NewGameManager()
	GameManager := game.NewManager(GameManagerCreate)
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

	client.On("gs", func(game models.Game) {
		fmt.Println("New Game State", game)
		if game.State == "FINISHED" {
			gameIdChannel <- ""
			fmt.Errorf("ИГРА ЗАКОНЧЕНА", game)
		} else if game.State == "GAME" {
			for _, participant := range game.Participants {
				if participant.IsActive {
					GameManager.MakeTurn(client, game)
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
