package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/maksekak/channelBot/cmd/internal/models"
)

type Client struct {
	Token  string
	APIURL string
}

func NewClient(token string) *Client {
	return &Client{
		Token:  token,
		APIURL: "https://api.telegram.org/bot" + token,
	}
}

func (c *Client) SendMessage(channelID int64, post models.Post) (int, error) {
	switch post.MediaType {
	case "text":
		return c.sendTextMessage(channelID, post)
	case "photo":
		return c.sendPhotoMessage(channelID, post)
	case "video":
		return c.sendVideoMessage(channelID, post)
	case "document":
		return c.sendDocumentMessage(channelID, post)
	default:
		return 0, fmt.Errorf("unsupported media type: %s", post.MediaType)
	}
}

func (c *Client) sendTextMessage(channelID int64, post models.Post) (int, error) {
	payload := map[string]interface{}{
		"chat_id":    strconv.FormatInt(channelID, 10),
		"text":       post.Content,
		"parse_mode": "HTML",
	}

	if len(post.Buttons) > 0 && string(post.Buttons) != "null" {
		var buttons []models.Button
		if err := json.Unmarshal(post.Buttons, &buttons); err == nil {
			payload["reply_markup"] = c.createInlineKeyboard(buttons)
		}
	}

	resp, err := c.makeRequest("sendMessage", payload)
	if err != nil {
		return 0, err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}

	if !result.OK {
		return 0, fmt.Errorf("telegram API error")
	}

	return result.Result.MessageID, nil
}

func (c *Client) sendPhotoMessage(channelID int64, post models.Post) (int, error) {
	return c.sendMediaMessage(channelID, post, "photo", "sendPhoto")
}

func (c *Client) sendVideoMessage(channelID int64, post models.Post) (int, error) {
	return c.sendMediaMessage(channelID, post, "video", "sendVideo")
}

func (c *Client) sendDocumentMessage(channelID int64, post models.Post) (int, error) {
	return c.sendMediaMessage(channelID, post, "document", "sendDocument")
}

func (c *Client) sendMediaMessage(channelID int64, post models.Post, fieldName, method string) (int, error) {
	file, err := os.Open(post.MediaPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл
	part, err := writer.CreateFormFile(fieldName, filepath.Base(post.MediaPath))
	if err != nil {
		return 0, err
	}
	io.Copy(part, file)

	// Добавляем остальные поля
	writer.WriteField("chat_id", strconv.FormatInt(channelID, 10))
	writer.WriteField("caption", post.Content)
	writer.WriteField("parse_mode", "HTML")

	if len(post.Buttons) > 0 && string(post.Buttons) != "null" {
		var buttons []models.Button
		if err := json.Unmarshal(post.Buttons, &buttons); err == nil {
			keyboard := c.createInlineKeyboard(buttons)
			keyboardJSON, _ := json.Marshal(keyboard)
			writer.WriteField("reply_markup", string(keyboardJSON))
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", c.APIURL+"/"+method, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, err
	}

	if !result.OK {
		return 0, fmt.Errorf("telegram API error: %s", string(respBody))
	}

	return result.Result.MessageID, nil
}

func (c *Client) createInlineKeyboard(buttons []models.Button) map[string]interface{} {
	var keyboard [][]map[string]string

	for _, button := range buttons {
		row := []map[string]string{
			{
				"text": button.Text,
				"url":  button.URL,
			},
		}
		keyboard = append(keyboard, row)
	}

	return map[string]interface{}{
		"inline_keyboard": keyboard,
	}
}

func (c *Client) makeRequest(method string, data map[string]interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.APIURL+"/"+method, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
