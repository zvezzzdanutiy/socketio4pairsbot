package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"socketiobot/internal/game"
	"socketiobot/internal/network"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Ошибка при создании cookie jar: %v", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Jar:     jar,
	}
	baseURL := "http://127.0.0.1:3000/ws/game/"
	response, err := network.FetchAndParseSocketResponse(baseURL, *client)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	}
	fmt.Println("Первый запрос:", response)
	sid := response.SID
	token, err := network.LoginAndGetToken("bot123", "pass123")
	if err != nil {
		fmt.Println(err)
	}
	responsetwo, err := network.FetchResponseTwo(baseURL, token, sid, *client, "40")
	if err != nil {
		fmt.Printf("Ошибка2: %v\n", err)
	}
	fmt.Println("Второй запрос:", responsetwo)
	responsethree, err := network.FetchResponseThree(baseURL, sid, *client)
	if err != nil {
		fmt.Printf("Ошибка3: %v\n", err)
	}
	fmt.Println("Третий запрос:", responsethree)
	socketURL := "ws://127.0.0.1:3000/ws/game/"
	websocketconnect, err := network.ConnectWebSocket(socketURL, sid)
	if err != nil {
		fmt.Printf("Ошибка4: %v\n", err)
		return
	}
	defer websocketconnect.Close()

	err = websocketconnect.WriteMessage(websocket.TextMessage, []byte("2probe"))
	if err != nil {
		fmt.Printf("Ошибка при отправке probe-сообщения: %v\n", err)
		return
	}

	upgradeDone := make(chan bool)
	var check string
	go func() {
		for {
			_, message, err := websocketconnect.ReadMessage()
			if err != nil {
				fmt.Printf("Ошибка при чтении сообщения: %v\n", err)
				return
			}

			fmt.Printf("Получено сообщение: %s\n", string(message))
			if string(message) == "2" {
				check = ""
			}
			if len(message) > 6 {
				check = string(string(message)[4]) + string(string(message)[5])
				fmt.Println("CHECK епта:", check)
			}
			if string(message) == "3probe" {
				err = websocketconnect.WriteMessage(websocket.TextMessage, []byte("5"))
				if err != nil {
					fmt.Printf("Ошибка при отправке финального сообщения повышения: %v\n", err)
					return
				}
				upgradeDone <- true
			} else if strings.Contains(string(message), "2") {
				err = websocketconnect.WriteMessage(websocket.TextMessage, []byte("3"))
				if err != nil {
					fmt.Printf("Ошибка при отправке pong-сообщения: %v\n", err)
					return
				}
			}
			if check == "jg" {
				fmt.Println("ПРИШЛО JGS")
				trimmedResponse := strings.TrimPrefix(string(message), "42")
				var parsedData []string
				err := json.Unmarshal([]byte(trimmedResponse), &parsedData)
				gameid := parsedData[1]
				if err != nil {
					fmt.Println("Ошибка при разборе JSON:", err)
					return
				}
				err = network.SendSocketRequest(websocketconnect, "joinGameSession", gameid)
				if err != nil {
					fmt.Printf("Ошибка при отправке joinGame: %v\n", err)
					return
				}
				check = ""
				continue
			}
			if check == "gs" {
				trimmedResponse := strings.TrimPrefix(string(message), "42")
				gameResponse, err := game.ParseGameResponse(trimmedResponse)
				if err != nil {
					fmt.Printf("Ошибка: %v\n", err)
					return
				}
				board := gameResponse.GameSession.Game.Board
				game.UpdateKnownCards(board)

				for _, participant := range gameResponse.GameSession.Participants {
					if participant.IsActive {
						err := game.MakeMove(websocketconnect, gameResponse.GameSession.ID, board)
						fmt.Println("Ход отправлен")
						if err != nil {
							fmt.Printf("Ошибка при совершении хода: %v\n", err)
						}
					}
				}

				continue
			}
		}
	}()

	<-upgradeDone
	err = network.SendSocketRequest2(websocketconnect, "searchGameSession")
	if err != nil {
		fmt.Printf("Ошибка при отправке searchGame: %v\n", err)
	}

	select {}
}
