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
