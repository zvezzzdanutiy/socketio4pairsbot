package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"socketiobot/internal/models"
	"strings"
	"time"

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
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/auth/login", bytes.NewBuffer(reqBody))
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
func FetchAndParseSocketResponse(baseURL string, client http.Client) (*models.SocketResponse, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("неверный URL: %v", err)
	}
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "polling")
	q.Set("t", fmt.Sprintf("%d-0", time.Now().UnixNano()/int64(time.Millisecond)))
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}
	body = RemoveNumericPrefix(body)

	var socketResp models.SocketResponse
	err = json.Unmarshal(body, &socketResp)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %v", err)
	}

	return &socketResp, nil
}

func FetchResponseTwo(baseURL string, token string, sid string, client http.Client, hnya string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("неверный URL: %v", err)
	}

	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "polling")
	q.Set("t", fmt.Sprintf("%d-0", time.Now().UnixNano()/int64(time.Millisecond)))
	q.Set("sid", sid)
	u.RawQuery = q.Encode()
	data := map[string]string{"token": token}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при кодировании данных: %v", err)
	}
	requestBody := hnya + string(jsonData)
	fmt.Println("requestbody", requestBody)
	resp, err := client.Post(u.String(), "application/json", strings.NewReader(requestBody))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	return string(body), nil
}

func FetchResponseThree(baseURL string, sid string, client http.Client) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("неверный URL: %v", err)
	}

	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "polling")
	q.Set("t", fmt.Sprintf("%d-0", time.Now().UnixNano()/int64(time.Millisecond)))
	q.Set("sid", sid)
	u.RawQuery = q.Encode()
	fmt.Println("Запросик-", u.String())
	resp, err := client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	cleanBody := RemoveNumericPrefix(body)

	fmt.Println("Тело ответа", string(cleanBody))
	return string(cleanBody), nil
}

func ConnectWebSocket(baseURL, sid string) (*websocket.Conn, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("неверный URL: %v", err)
	}

	u.Scheme = "ws"
	u.Path = "/ws/game/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	q.Set("sid", sid)
	u.RawQuery = q.Encode()

	fmt.Println("Url sockets:", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при установке WebSocket соединения: %v", err)
	}

	return c, nil
}
func SendSocketRequest(conn *websocket.Conn, event string, data interface{}) error {
	payload, err := json.Marshal([]interface{}{event, data})
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}
	message := fmt.Sprintf("42%s", string(payload))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
func SendSocketRequest2(conn *websocket.Conn, event string) error {
	payload, err := json.Marshal([]interface{}{event})
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}
	message := fmt.Sprintf("42%s", string(payload))
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func ReadSocketResponse(conn *websocket.Conn) (string, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(message), nil
}
