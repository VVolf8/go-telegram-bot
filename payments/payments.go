package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/VVolf8/go-telegram-bot/core"
)

// Price represents a labeled price for the invoice.
type Price struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

// Invoice represents the invoice details for sending payments.
type Invoice struct {
	ChatID         int64   `json:"chat_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Payload        string  `json:"payload"`
	ProviderToken  string  `json:"provider_token"`
	StartParameter string  `json:"start_parameter"`
	Currency       string  `json:"currency"`
	Prices         []Price `json:"prices"`
	// Можно добавить дополнительные опциональные поля по необходимости.
}

// PaymentService defines the interface for payment-related methods.
type PaymentService interface {
	// SendInvoice sends an invoice to a chat.
	SendInvoice(ctx context.Context, invoice Invoice) error
	// AnswerShippingQuery responds to a shipping query.
	AnswerShippingQuery(ctx context.Context, shippingQueryID string, ok bool, errorMessage string) error
	// AnswerPreCheckoutQuery responds to a pre-checkout query.
	AnswerPreCheckoutQuery(ctx context.Context, preCheckoutQueryID string, ok bool, errorMessage string) error
}

// PaymentServiceExtended расширяет базовый PaymentService дополнительными методами.
type PaymentServiceExtended interface {
	PaymentService
	// HandleSuccessfulPayment обрабатывает обновление об успешной оплате.
	HandleSuccessfulPayment(ctx context.Context, update core.Update) error
	// ProcessRefund обрабатывает возврат платежа.
	ProcessRefund(ctx context.Context, paymentID string) error
	// GenerateReceipt генерирует квитанцию для платежа.
	GenerateReceipt(ctx context.Context, paymentID string) (string, error)
}

// paymentService is the concrete implementation of PaymentService.
type paymentService struct {
	token      string
	apiURL     string
	httpClient *http.Client
	logger     core.Logger
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(token string, logger core.Logger, httpClient *http.Client) PaymentService {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &paymentService{
		token:      token,
		apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", token),
		httpClient: httpClient,
		logger:     logger,
	}
}

// SendInvoice sends an invoice via Telegram.
func (ps *paymentService) SendInvoice(ctx context.Context, invoice Invoice) error {
	endpoint := fmt.Sprintf("%s/sendInvoice", ps.apiURL)
	// Создаем JSON-пейлоад на основе структуры Invoice.
	payloadBytes, err := json.Marshal(invoice)
	if err != nil {
		ps.logger.Error("Failed to marshal SendInvoice payload", core.Field{"error", err})
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		ps.logger.Error("Failed to create SendInvoice request", core.Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	core.WithRecovery(ps.logger, func() {
		resp, err = ps.httpClient.Do(req)
	})
	if err != nil {
		ps.logger.Error("Error sending invoice", core.Field{"error", err})
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ps.logger.Error("Failed to read SendInvoice response", core.Field{"error", err})
		return err
	}
	if resp.StatusCode != http.StatusOK {
		ps.logger.Error("Non-OK response from SendInvoice",
			core.Field{"status", resp.Status},
			core.Field{"body", string(bodyBytes)},
		)
		return fmt.Errorf("SendInvoice failed with status: %s", resp.Status)
	}

	ps.logger.Info("Invoice sent successfully", core.Field{"chat_id", invoice.ChatID}, core.Field{"title", invoice.Title})
	return nil
}

// AnswerShippingQuery responds to a shipping query.
func (ps *paymentService) AnswerShippingQuery(ctx context.Context, shippingQueryID string, ok bool, errorMessage string) error {
	endpoint := fmt.Sprintf("%s/answerShippingQuery", ps.apiURL)
	payload := map[string]interface{}{
		"shipping_query_id": shippingQueryID,
		"ok":                ok,
	}
	if !ok {
		payload["error_message"] = errorMessage
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ps.logger.Error("Failed to marshal AnswerShippingQuery payload", core.Field{"error", err})
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		ps.logger.Error("Failed to create AnswerShippingQuery request", core.Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	core.WithRecovery(ps.logger, func() {
		resp, err = ps.httpClient.Do(req)
	})
	if err != nil {
		ps.logger.Error("Error answering shipping query", core.Field{"error", err})
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ps.logger.Error("Failed to read AnswerShippingQuery response", core.Field{"error", err})
		return err
	}
	if resp.StatusCode != http.StatusOK {
		ps.logger.Error("Non-OK response from AnswerShippingQuery",
			core.Field{"status", resp.Status},
			core.Field{"body", string(bodyBytes)},
		)
		return fmt.Errorf("AnswerShippingQuery failed with status: %s", resp.Status)
	}

	ps.logger.Info("Shipping query answered successfully", core.Field{"shipping_query_id", shippingQueryID})
	return nil
}

// AnswerPreCheckoutQuery responds to a pre-checkout query.
func (ps *paymentService) AnswerPreCheckoutQuery(ctx context.Context, preCheckoutQueryID string, ok bool, errorMessage string) error {
	endpoint := fmt.Sprintf("%s/answerPreCheckoutQuery", ps.apiURL)
	payload := map[string]interface{}{
		"pre_checkout_query_id": preCheckoutQueryID,
		"ok":                    ok,
	}
	if !ok {
		payload["error_message"] = errorMessage
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ps.logger.Error("Failed to marshal AnswerPreCheckoutQuery payload", core.Field{"error", err})
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		ps.logger.Error("Failed to create AnswerPreCheckoutQuery request", core.Field{"error", err})
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	core.WithRecovery(ps.logger, func() {
		resp, err = ps.httpClient.Do(req)
	})
	if err != nil {
		ps.logger.Error("Error answering pre-checkout query", core.Field{"error", err})
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ps.logger.Error("Failed to read AnswerPreCheckoutQuery response", core.Field{"error", err})
		return err
	}
	if resp.StatusCode != http.StatusOK {
		ps.logger.Error("Non-OK response from AnswerPreCheckoutQuery",
			core.Field{"status", resp.Status},
			core.Field{"body", string(bodyBytes)},
		)
		return fmt.Errorf("AnswerPreCheckoutQuery failed with status: %s", resp.Status)
	}

	ps.logger.Info("Pre-checkout query answered successfully", core.Field{"pre_checkout_query_id", preCheckoutQueryID})
	return nil
}

// HandleSuccessfulPayment обрабатывает обновление об успешной оплате.
// Обычно Telegram отправляет обновление, содержащее информацию об успешной оплате.
// Здесь можно, например, логировать событие и выполнять дополнительные действия (уведомлять, сохранять данные и т.д.).
func (ps *paymentService) HandleSuccessfulPayment(ctx context.Context, update core.Update) error {
	// Пример: логируем успешную оплату. 
	// Для реальной обработки нужно извлечь необходимые данные из update (например, через update.Message или специальное поле).
	ps.logger.Info("Payment processed successfully", core.Field{"update_id", update.UpdateID})
	// Здесь можно вызвать GenerateReceipt и, например, отправить квитанцию пользователю.
	receipt, err := ps.GenerateReceipt(ctx, "payment_id_example")
	if err != nil {
		ps.logger.Error("Error generating receipt", core.Field{"error", err})
		return err
	}
	ps.logger.Info("Receipt generated", core.Field{"receipt", receipt})
	return nil
}

// ProcessRefund обрабатывает возврат платежа.
// Здесь можно реализовать логику взаимодействия с платежным провайдером для проведения возврата.
func (ps *paymentService) ProcessRefund(ctx context.Context, paymentID string) error {
	// Пример: логируем запрос на возврат и возвращаем ошибку, если что-то пошло не так.
	ps.logger.Info("Processing refund", core.Field{"payment_id", paymentID})
	// Здесь должна быть реальная логика работы с API платежного провайдера.
	// Для примера возвращаем nil.
	return nil
}

// GenerateReceipt генерирует квитанцию для платежа и возвращает строковое представление квитанции.
func (ps *paymentService) GenerateReceipt(ctx context.Context, paymentID string) (string, error) {
	// Пример: генерация квитанции (например, в виде JSON или строки).
	receiptData := map[string]interface{}{
		"payment_id": paymentID,
		"status":     "success",
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	receiptBytes, err := json.Marshal(receiptData)
	if err != nil {
		ps.logger.Error("Failed to marshal receipt", core.Field{"error", err})
		return "", err
	}
	receipt := string(receiptBytes)
	ps.logger.Info("Receipt generated successfully", core.Field{"payment_id", paymentID})
	return receipt, nil
}
