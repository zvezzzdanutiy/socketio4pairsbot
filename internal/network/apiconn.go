package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
