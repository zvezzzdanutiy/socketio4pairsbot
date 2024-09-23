package game

import (
	"encoding/json"
	"fmt"
	"socketiobot/internal/models"
	"time"

	"github.com/gorilla/websocket"
)

type Manager struct {
	gameManager *models.GameManager
}

func NewManager(gm *models.GameManager) *Manager {
	return &Manager{
		gameManager: gm,
	}
}

func (m *Manager) ParseGameResponse(jsonData string) (models.GameResponse, error) {
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

func (m *Manager) UpdateKnownCards(board [][]int) {
	for i, stack := range board {
		currentLen := len(stack)
		if prevLen, exists := m.gameManager.KnownLens[i]; exists && currentLen != prevLen {
			for card, positions := range m.gameManager.KnownCards {
				for j, pos := range positions {
					if pos == i {
						m.gameManager.KnownCards[card] = append(positions[:j], positions[j+1:]...)
						break
					}
				}
				if len(m.gameManager.KnownCards[card]) == 0 {
					delete(m.gameManager.KnownCards, card)
				}
			}
		}
		m.gameManager.KnownLens[i] = currentLen

		if currentLen > 0 && stack[0] != -1 {
			card := stack[0]
			if _, ok := m.gameManager.KnownCards[card]; !ok {
				m.gameManager.KnownCards[card] = []int{i}
			} else if !Contains(m.gameManager.KnownCards[card], i) {
				m.gameManager.KnownCards[card] = append(m.gameManager.KnownCards[card], i)
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

func (m *Manager) FindBestMove(board [][]int) (int, int) {
	for _, positions := range m.gameManager.KnownCards {
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
		randomIndex := m.gameManager.Random.Intn(len(availablePositions))
		return availablePositions[randomIndex], -1
	}

	return -1, -1
}

func (m *Manager) MakeMove(conn *websocket.Conn, gameID string, board [][]int) error {
	firstCard, secondCard := m.FindBestMove(board)

	if firstCard == -1 {
		return fmt.Errorf("нет доступных ходов")
	}

	if err := m.SendTurn(conn, gameID, firstCard); err != nil {
		return err
	}

	if secondCard != -1 {
		time.Sleep(750 * time.Millisecond)
		return m.SendTurn(conn, gameID, secondCard)
	}

	return nil
}

func (m *Manager) SendTurn(conn *websocket.Conn, gameID string, position int) error {
	turnData := map[string]interface{}{
		"columnId":      position,
		"gameSessionId": gameID,
	}
	return m.SendSocketRequest(conn, "turn", turnData)
}

func (m *Manager) SendSocketRequest(conn *websocket.Conn, event string, data interface{}) error {
	payload, err := json.Marshal([]interface{}{event, data})
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}
	message := fmt.Sprintf("42%s", string(payload))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
