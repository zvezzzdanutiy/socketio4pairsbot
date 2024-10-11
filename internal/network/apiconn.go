package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"socketio4/internal/models"

	"github.com/gorilla/websocket"
)

func RemoveNumericPrefix(data []byte) []byte {
	for i, b := range data {
		if b < '0' || b > '9' {
			return data[i:]
		}
	}
	return data
}
func LoginAndGetToken(username, password string) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, "https://bun-gs-bungamecenter.amvera.io/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("сервер вернул ошибку: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var respBody map[string]string
	if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
		return "", err
	}
	return respBody["access_token"], nil
}
func ConnectWebSocket(baseURL string) (*websocket.Conn, *models.SocketResponse, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing URL: %w", err)
	}

	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()

	fmt.Println("Connecting to:", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error establishing WebSocket connection: %w", err)
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		c.Close()
		return nil, nil, fmt.Errorf("error reading message: %w", err)
	}

	message = RemoveNumericPrefix(message)

	var response models.SocketResponse
	err = json.Unmarshal(message, &response)
	if err != nil {
		c.Close()
		return nil, nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return c, &response, nil
}
func SendSocketRequest(conn *websocket.Conn, event string) error {
	data := map[string]string{"token": event}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка при кодировании данных: %v", err)
	}
	message := fmt.Sprintf("40%s", string(jsonData))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
func SendSocketRequest2(conn *websocket.Conn, event string, event2 string, number string) error {
	payload, err := json.Marshal([]string{event, event2})
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}
	message := number + string(payload)
	fmt.Println("Отправляемый запрос:", string(message))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func ReadSocketResponse(conn *websocket.Conn) (string, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(message), nil
}
