package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"socketiobot/internal/models"
	"time"

	"github.com/gorilla/websocket"
)

var (
	knownCards map[int][]int
	knownLens  map[int]int
)

func init() {
	knownCards = make(map[int][]int)
	knownLens = make(map[int]int)
}

func ParseGameResponse(jsonData string) (models.GameResponse, error) {
	var response []json.RawMessage
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		return models.GameResponse{}, fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	if len(response) < 2 {
		return models.GameResponse{}, fmt.Errorf("некорректный формат JSON")
	}

	var gameResponse models.GameResponse
	err = json.Unmarshal(response[1], &gameResponse.GameSession)
	if err != nil {
		return models.GameResponse{}, fmt.Errorf("ошибка при разборе GameSession: %v", err)
	}

	return gameResponse, nil
}

func UpdateKnownCards(board [][]int) {
	for i, stack := range board {
		currentLen := len(stack)
		if prevLen, exists := knownLens[i]; exists && currentLen != prevLen {
			for card, positions := range knownCards {
				for j, pos := range positions {
					if pos == i {
						knownCards[card] = append(positions[:j], positions[j+1:]...)
						break
					}
				}
				if len(knownCards[card]) == 0 {
					delete(knownCards, card)
				}
			}
		}
		knownLens[i] = currentLen

		if currentLen > 0 && stack[0] != -1 {
			card := stack[0]
			if _, ok := knownCards[card]; !ok {
				knownCards[card] = []int{i}
			} else if !Contains(knownCards[card], i) {
				knownCards[card] = append(knownCards[card], i)
			}
		}
	}
}

func Contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func FindBestMove(board [][]int) (int, int) {
	for _, positions := range knownCards {
		if len(positions) == 2 {
			return positions[0], positions[1]
		}
	}

	availablePositions := make([]int, 0)
	for i, stack := range board {
		if len(stack) > 0 {
			availablePositions = append(availablePositions, i)
		}
	}

	if len(availablePositions) > 0 {
		randomIndex := rand.Intn(len(availablePositions))
		return availablePositions[randomIndex], -1
	}

	return -1, -1
}

func MakeMove(conn *websocket.Conn, gameID string, board [][]int) error {
	firstCard, secondCard := FindBestMove(board)

	if firstCard == -1 {
		return fmt.Errorf("нет доступных ходов")
	}

	if err := SendTurn(conn, gameID, firstCard); err != nil {
		return err
	}

	if secondCard != -1 {
		time.Sleep(750 * time.Millisecond)
		return SendTurn(conn, gameID, secondCard)
	}

	return nil
}
func SendTurn(conn *websocket.Conn, gameID string, position int) error {
	turnData := map[string]interface{}{
		"columnId":      position,
		"gameSessionId": gameID,
	}
	return SendSocketRequest(conn, "turn", turnData)
}
func SendSocketRequest(conn *websocket.Conn, event string, data interface{}) error {
	payload, err := json.Marshal([]interface{}{event, data})
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}
	message := fmt.Sprintf("42%s", string(payload))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
