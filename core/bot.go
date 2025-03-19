package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Расширенный интерфейс BotAPI с дополнительными методами.
type BotAPI interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithMarkup(ctx context.Context, chatID int64, text string, replyMarkup interface{}) error
	GetUpdates(ctx context.Context, offset, limit, timeout int) ([]Update, error)
	SendPhoto(ctx context.Context, chatID int64, photo interface{}, caption string, replyMarkup interface{}) error
	SendDocument(ctx context.Context, chatID int64, document interface{}, caption string, replyMarkup interface{}) error
	EditMessageText(ctx context.Context, chatID int64, messageID int, text string, replyMarkup interface{}) error
	EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int, replyMarkup interface{}) error
	AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error
	ForwardMessage(ctx context.Context, chatID int64, fromChatID int64, messageID int) error
	GetChat(ctx context.Context, chatID int64) (Chat, error)
	GetChatMembersCount(ctx context.Context, chatID int64) (int, error)
	GetChatAdministrators(ctx context.Context, chatID int64) ([]Chat, error)
	GetMe(ctx context.Context) (User, error)
	// Другие методы можно добавить при необходимости.
}

// botClient – реализация интерфейса BotAPI.
type botClient struct {
	token      string
	apiURL     string
	httpClient *http.Client
	logger     Logger
}

// NewBotClient возвращает новый экземпляр BotAPI, инициализированный токеном, логгером и HTTP-клиентом.
func NewBotClient(token string, logger Logger, httpClient *http.Client) BotAPI {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &botClient{
		token:      token,
		apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", token),
		httpClient: httpClient,
		logger:     logger,
	}
}

// SendMessage отправляет текстовое сообщение в указанный чат.
func (b *botClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	endpoint := fmt.Sprintf("%s/sendMessage", b.apiURL)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal sendMessage payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create sendMessage request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error sending message", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from sendMessage", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("sendMessage failed with status: %s", resp.Status)
	}
	b.logger.Info("Message sent successfully", Field{"chat_id", chatID}, Field{"text", text})
	return nil
}

// SendMessageWithMarkup отправляет сообщение с дополнительной разметкой (например, inline-клавиатурой).
func (b *botClient) SendMessageWithMarkup(ctx context.Context, chatID int64, text string, replyMarkup interface{}) error {
	endpoint := fmt.Sprintf("%s/sendMessage", b.apiURL)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": replyMarkup,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal sendMessageWithMarkup payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create sendMessageWithMarkup request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error sending message with markup", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from sendMessageWithMarkup", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("sendMessageWithMarkup failed with status: %s", resp.Status)
	}
	b.logger.Info("Message with markup sent successfully", Field{"chat_id", chatID}, Field{"text", text})
	return nil
}

// GetUpdates получает обновления от Telegram API с использованием контекста.
func (b *botClient) GetUpdates(ctx context.Context, offset, limit, timeout int) ([]Update, error) {
	endpoint := fmt.Sprintf("%s/getUpdates", b.apiURL)
	params := url.Values{}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("timeout", strconv.Itoa(timeout))
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		b.logger.Error("Failed to create getUpdates request", Field{"error", err})
		return nil, err
	}
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing getUpdates request", Field{"error", err})
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from getUpdates", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return nil, fmt.Errorf("getUpdates failed with status: %s", resp.Status)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading getUpdates response", Field{"error", err})
		return nil, err
	}
	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		b.logger.Error("Error unmarshalling getUpdates response", Field{"error", err})
		return nil, err
	}
	if !result.OK {
		b.logger.Error("Telegram API returned not OK for getUpdates", Field{"response", string(respBody)})
		return nil, fmt.Errorf("telegram API error: %s", string(respBody))
	}
	b.logger.Info("Fetched updates", Field{"updates_count", len(result.Result)})
	return result.Result, nil
}

// SendPhoto отправляет фото в указанный чат.
// Параметр photo может быть либо строкой (URL или file_id), либо файлом (но для файлов нужна дополнительная обработка).
func (b *botClient) SendPhoto(ctx context.Context, chatID int64, photo interface{}, caption string, replyMarkup interface{}) error {
	endpoint := fmt.Sprintf("%s/sendPhoto", b.apiURL)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"photo":   photo,
		"caption": caption,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal sendPhoto payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create sendPhoto request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error sending photo", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from sendPhoto", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("sendPhoto failed with status: %s", resp.Status)
	}
	b.logger.Info("Photo sent successfully", Field{"chat_id", chatID})
	return nil
}

// SendDocument отправляет документ в указанный чат.
func (b *botClient) SendDocument(ctx context.Context, chatID int64, document interface{}, caption string, replyMarkup interface{}) error {
	endpoint := fmt.Sprintf("%s/sendDocument", b.apiURL)
	payload := map[string]interface{}{
		"chat_id":  chatID,
		"document": document,
		"caption":  caption,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal sendDocument payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create sendDocument request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error sending document", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from sendDocument", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("sendDocument failed with status: %s", resp.Status)
	}
	b.logger.Info("Document sent successfully", Field{"chat_id", chatID})
	return nil
}

// EditMessageText редактирует текст ранее отправленного сообщения.
func (b *botClient) EditMessageText(ctx context.Context, chatID int64, messageID int, text string, replyMarkup interface{}) error {
	endpoint := fmt.Sprintf("%s/editMessageText", b.apiURL)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal editMessageText payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create editMessageText request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error editing message text", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from editMessageText", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("editMessageText failed with status: %s", resp.Status)
	}
	b.logger.Info("Message text edited successfully", Field{"chat_id", chatID}, Field{"message_id", messageID})
	return nil
}

// EditMessageReplyMarkup обновляет reply_markup для сообщения.
func (b *botClient) EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int, replyMarkup interface{}) error {
	endpoint := fmt.Sprintf("%s/editMessageReplyMarkup", b.apiURL)
	payload := map[string]interface{}{
		"chat_id":     chatID,
		"message_id":  messageID,
		"reply_markup": replyMarkup,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal editMessageReplyMarkup payload", Field{"error", err})
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create editMessageReplyMarkup request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing editMessageReplyMarkup request", Field{"error", err})
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading editMessageReplyMarkup response", Field{"error", err})
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b.logger.Error("Non-OK response from editMessageReplyMarkup",
		 Field{"status", resp.Status},
			Field{"body", string(respBody)},
		)
		return fmt.Errorf("editMessageReplyMarkup failed with status: %s", resp.Status)
	}

	b.logger.Info("Message reply markup edited successfully", core.Field{"chat_id", chatID}, Field{"message_id", messageID})
	return nil
}


// AnswerCallbackQuery отвечает на callback-запрос.
func (b *botClient) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error {
	endpoint := fmt.Sprintf("%s/answerCallbackQuery", b.apiURL)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        showAlert,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal answerCallbackQuery payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create answerCallbackQuery request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error answering callback query", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from answerCallbackQuery", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("answerCallbackQuery failed with status: %s", resp.Status)
	}
	b.logger.Info("Callback query answered successfully", Field{"callback_query_id", callbackQueryID})
	return nil
}

// ForwardMessage пересылает сообщение из одного чата в другой.
func (b *botClient) ForwardMessage(ctx context.Context, chatID int64, fromChatID int64, messageID int) error {
	endpoint := fmt.Sprintf("%s/forwardMessage", b.apiURL)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"from_chat_id": fromChatID,
		"message_id":   messageID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		b.logger.Error("Failed to marshal forwardMessage payload", Field{"error", err})
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		b.logger.Error("Failed to create forwardMessage request", Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error forwarding message", Field{"error", err})
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		b.logger.Error("Non-OK response from forwardMessage", Field{"status", resp.Status}, Field{"body", string(respBody)})
		return fmt.Errorf("forwardMessage failed with status: %s", resp.Status)
	}
	b.logger.Info("Message forwarded successfully", Field{"chat_id", chatID}, Field{"from_chat_id", fromChatID}, Field{"message_id", messageID})
	return nil
}

// Пример реализации метода GetChat.
func (b *botClient) GetChat(ctx context.Context, chatID int64) (Chat, error) {
	endpoint := fmt.Sprintf("%s/getChat?chat_id=%d", b.apiURL, chatID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		b.logger.Error("Failed to create getChat request", Field{"error", err})
		return Chat{}, err
	}
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing getChat request", Field{"error", err})
		return Chat{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading getChat response", Field{"error", err})
		return Chat{}, err
	}
	var result struct {
		OK     bool `json:"ok"`
		Result Chat `json:"result"`
	}
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		b.logger.Error("Error unmarshalling getChat response", Field{"error", err})
		return Chat{}, err
	}
	if !result.OK {
		b.logger.Error("Telegram API returned not OK for getChat", Field{"response", string(bodyBytes)})
		return Chat{}, fmt.Errorf("getChat failed with response: %s", string(bodyBytes))
	}
	b.logger.Info("Chat retrieved successfully", Field{"chat_id", chatID})
	return result.Result, nil
}

// Пример реализации метода GetChatMembersCount.
func (b *botClient) GetChatMembersCount(ctx context.Context, chatID int64) (int, error) {
	endpoint := fmt.Sprintf("%s/getChatMembersCount?chat_id=%d", b.apiURL, chatID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		b.logger.Error("Failed to create getChatMembersCount request", Field{"error", err})
		return 0, err
	}
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing getChatMembersCount request", Field{"error", err})
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading getChatMembersCount response", Field{"error", err})
		return 0, err
	}
	var result struct {
		OK     bool `json:"ok"`
		Result int  `json:"result"`
	}
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		b.logger.Error("Error unmarshalling getChatMembersCount response", Field{"error", err})
		return 0, err
	}
	if !result.OK {
		b.logger.Error("Telegram API returned not OK for getChatMembersCount", Field{"response", string(bodyBytes)})
		return 0, fmt.Errorf("getChatMembersCount failed with response: %s", string(bodyBytes))
	}
	b.logger.Info("Chat members count retrieved", Field{"chat_id", chatID}, Field{"count", result.Result})
	return result.Result, nil
}

// Пример реализации метода GetChatAdministrators.
func (b *botClient) GetChatAdministrators(ctx context.Context, chatID int64) ([]Chat, error) {
	endpoint := fmt.Sprintf("%s/getChatAdministrators?chat_id=%d", b.apiURL, chatID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		b.logger.Error("Failed to create getChatAdministrators request", Field{"error", err})
		return nil, err
	}
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing getChatAdministrators request", Field{"error", err})
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading getChatAdministrators response", Field{"error", err})
		return nil, err
	}
	var result struct {
		OK     bool   `json:"ok"`
		Result []Chat `json:"result"`
	}
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		b.logger.Error("Error unmarshalling getChatAdministrators response", Field{"error", err})
		return nil, err
	}
	if !result.OK {
		b.logger.Error("Telegram API returned not OK for getChatAdministrators", Field{"response", string(bodyBytes)})
		return nil, fmt.Errorf("getChatAdministrators failed with response: %s", string(bodyBytes))
	}
	b.logger.Info("Chat administrators retrieved", Field{"chat_id", chatID}, Field{"count", len(result.Result)})
	return result.Result, nil
}

func (b *botClient) GetMe(ctx context.Context) (User, error) {
	endpoint := fmt.Sprintf("%s/getMe", b.apiURL)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		b.logger.Error("Failed to create getMe request", Field{"error", err})
		return User{}, err
	}
	var resp *http.Response
	WithRecovery(b.logger, func() {
		resp, err = b.httpClient.Do(req)
	})
	if err != nil {
		b.logger.Error("Error executing getMe request", Field{"error", err})
		return User{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading getMe response", Field{"error", err})
		return User{}, err
	}
	var result struct {
		OK     bool `json:"ok"`
		Result User `json:"result"`
	}
	if err = json.Unmarshal(bodyBytes, &result); err != nil {
		b.logger.Error("Error unmarshalling getMe response", Field{"error", err})
		return User{}, err
	}
	if !result.OK {
		b.logger.Error("Telegram API returned not OK for getMe", Field{"response", string(bodyBytes)})
		return User{}, fmt.Errorf("getMe failed with response: %s", string(bodyBytes))
	}
	b.logger.Info("GetMe executed successfully")
	return result.Result, nil
}
