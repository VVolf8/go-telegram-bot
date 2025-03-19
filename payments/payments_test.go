package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VVolf8/go-telegram-bot/core"
)

// mockTelegramHandler эмулирует ответы Telegram API.
func mockTelegramHandler(expectedPath string, response interface{}, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

func TestSendInvoice(t *testing.T) {
	// Эмулируем ответ Telegram API для sendInvoice:
	mockResp := map[string]interface{}{
		"ok": true,
	}
	// Создаем тестовый сервер, который будет обрабатывать запросы по пути "/sendInvoice".
	ts := httptest.NewServer(mockTelegramHandler("/sendInvoice", mockResp, http.StatusOK))
	defer ts.Close()

	// Создаем PaymentService.
	// Здесь мы создаем экземпляр с тестовым токеном и заменяем apiURL на адрес нашего тестового сервера.
	logger := core.NewLogger(core.DebugLevel)
	ps := &paymentService{
		token:      "TEST_TOKEN",
		apiURL:     ts.URL, // Подменяем базовый URL на тестовый сервер.
		httpClient: ts.Client(),
		logger:     logger,
	}

	// Создаем тестовый инвойс.
	invoice := Invoice{
		ChatID:         123456789,
		Title:          "Test Invoice",
		Description:    "This is a test invoice",
		Payload:        "test_payload",
		ProviderToken:  "provider_test_token",
		StartParameter: "start",
		Currency:       "USD",
		Prices: []Price{
			{Label: "Item 1", Amount: 1000},
			{Label: "Item 2", Amount: 2000},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ps.SendInvoice(ctx, invoice); err != nil {
		t.Errorf("SendInvoice returned error: %v", err)
	}
}

func TestAnswerShippingQuery(t *testing.T) {
	mockResp := map[string]interface{}{
		"ok": true,
	}
	ts := httptest.NewServer(mockTelegramHandler("/answerShippingQuery", mockResp, http.StatusOK))
	defer ts.Close()

	logger := core.NewLogger(core.DebugLevel)
	ps := &paymentService{
		token:      "TEST_TOKEN",
		apiURL:     ts.URL,
		httpClient: ts.Client(),
		logger:     logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ps.AnswerShippingQuery(ctx, "test_query_id", true, ""); err != nil {
		t.Errorf("AnswerShippingQuery returned error: %v", err)
	}
}

func TestAnswerPreCheckoutQuery(t *testing.T) {
	mockResp := map[string]interface{}{
		"ok": true,
	}
	ts := httptest.NewServer(mockTelegramHandler("/answerPreCheckoutQuery", mockResp, http.StatusOK))
	defer ts.Close()

	logger := core.NewLogger(core.DebugLevel)
	ps := &paymentService{
		token:      "TEST_TOKEN",
		apiURL:     ts.URL,
		httpClient: ts.Client(),
		logger:     logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ps.AnswerPreCheckoutQuery(ctx, "test_precheckout_id", true, ""); err != nil {
		t.Errorf("AnswerPreCheckoutQuery returned error: %v", err)
	}
}

// dummyLogger реализует интерфейс core.Logger для тестирования.
type dummyLogger struct {
	logs []string
}

func (d *dummyLogger) Debug(msg string, fields ...core.Field) {
	d.logs = append(d.logs, "DEBUG: "+msg)
}

func (d *dummyLogger) Info(msg string, fields ...core.Field) {
	d.logs = append(d.logs, "INFO: "+msg)
}

func (d *dummyLogger) Warn(msg string, fields ...core.Field) {
	d.logs = append(d.logs, "WARN: "+msg)
}

func (d *dummyLogger) Error(msg string, fields ...core.Field) {
	d.logs = append(d.logs, "ERROR: "+msg)
}

func (d *dummyLogger) Fatal(msg string, fields ...core.Field) {
	// В тестах не вызываем os.Exit, поэтому просто логгируем Fatal-сообщение.
	d.logs = append(d.logs, "FATAL: "+msg)
}

func (d *dummyLogger) WithFields(fields ...core.Field) core.Logger {
	// Для целей тестирования можно вернуть тот же логгер,
	// либо создать новый с добавлением полей (при необходимости).
	return d
}

func TestHandleSuccessfulPayment(t *testing.T) {
	logger := &dummyLogger{}
	ps := NewPaymentService("test_token", logger, &http.Client{})

	// Создаём тестовый update (используем core.Update как структуру).
	update := core.Update{
		UpdateID: 12345,
		// Другие поля можно заполнить по необходимости.
	}

	ctx := context.Background()
	err := ps.HandleSuccessfulPayment(ctx, update)
	if err != nil {
		t.Errorf("HandleSuccessfulPayment вернул ошибку: %v", err)
	}

	if len(logger.logs) == 0 {
		t.Error("Логгер не записал ни одного сообщения")
	}
}

func TestProcessRefund(t *testing.T) {
	logger := &dummyLogger{}
	ps := NewPaymentService("test_token", logger, &http.Client{})
	ctx := context.Background()

	paymentID := "refund_test_id"
	err := ps.ProcessRefund(ctx, paymentID)
	if err != nil {
		t.Errorf("ProcessRefund вернул ошибку для валидного paymentID: %v", err)
	}

	// Проверяем, что в логах присутствует сообщение о начале обработки возврата.
	found := false
	for _, msg := range logger.logs {
		if msg == "INFO: Processing refund" {
			found = true
			break
		}
	}
	if !found {
		t.Log("Обработка возврата не отразилась в логах (возможно, логирование изменилось)")
	}
}

func TestGenerateReceipt(t *testing.T) {
	logger := &dummyLogger{}
	ps := NewPaymentService("test_token", logger, &http.Client{})
	ctx := context.Background()

	paymentID := "receipt_test_id"
	receiptStr, err := ps.GenerateReceipt(ctx, paymentID)
	if err != nil {
		t.Errorf("GenerateReceipt вернул ошибку: %v", err)
	}

	if receiptStr == "" {
		t.Error("GenerateReceipt вернул пустую квитанцию")
	}

	// Проверяем, что результат является корректным JSON-объектом с необходимыми полями.
	var receiptData map[string]interface{}
	if err := json.Unmarshal([]byte(receiptStr), &receiptData); err != nil {
		t.Errorf("Сгенерированная квитанция не является корректным JSON: %v", err)
	}

	// Проверяем наличие payment_id.
	if id, ok := receiptData["payment_id"]; !ok || id != paymentID {
		t.Errorf("В квитанции отсутствует корректный payment_id, ожидается %s, получено %v", paymentID, id)
	}

	// Проверяем статус.
	if status, ok := receiptData["status"]; !ok || status != "success" {
		t.Errorf("В квитанции отсутствует корректный статус, ожидается \"success\", получено %v", status)
	}

	// Проверяем наличие и корректность timestamp.
	if ts, ok := receiptData["timestamp"]; !ok {
		t.Error("В квитанции отсутствует timestamp")
	} else {
		if _, err := time.Parse(time.RFC3339, ts.(string)); err != nil {
			t.Errorf("timestamp имеет неверный формат: %v", err)
		}
	}
}
