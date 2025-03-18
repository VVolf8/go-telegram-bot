package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/VVolf8/go-telegram-bot/core"
)

// MiddlewareFunc определяет функцию middleware, принимающую и возвращающую HandlerFunc.
type MiddlewareFunc func(core.HandlerFunc) core.HandlerFunc

// =======================
// SecurityMiddleware
// =======================
// SecurityMiddleware проверяет, что ID чата (или пользователя) входит в список разрешённых.
// Если обновление пришло от неразрешённого источника, обработчик не вызывается.
func SecurityMiddleware(allowedIDs []int64, logger core.Logger) MiddlewareFunc {
	allowedMap := make(map[int64]struct{})
	for _, id := range allowedIDs {
		allowedMap[id] = struct{}{}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			// Проверяем, что обновление содержит сообщение и его чат ID присутствует в списке allowedIDs.
			if update.Message == nil {
				logger.Warn("SecurityMiddleware: update has no message")
				return fmt.Errorf("security: no message in update")
			}
			chatID := update.Message.Chat.ID
			if _, ok := allowedMap[chatID]; !ok {
				logger.Warn("SecurityMiddleware: access denied", core.Field{"chat_id", chatID})
				return fmt.Errorf("security: access denied for chat %d", chatID)
			}
			return next(update)
		}
	}
}

// =======================
// TracingMiddleware
// =======================
// TracingMiddleware генерирует уникальный correlation ID для каждого обновления,
// добавляет его в контекст и логирует его. Это позволяет отслеживать цепочки запросов.
func TracingMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			correlationID := generateCorrelationID()
			// Добавляем correlation ID в контекст (если понадобится передавать его дальше)
			ctx := context.WithValue(context.Background(), "correlation_id", correlationID)
			// Можно также добавить его в базовые поля логгера:
			loggerWithCorr := logger.WithFields(core.Field{"correlation_id", correlationID})
			loggerWithCorr.Debug("TracingMiddleware: generated correlation ID", core.Field{"correlation_id", correlationID})
			// Здесь можно передать новый контекст в обработчик, если HandlerFunc поддерживает передачу контекста.
			// В нашем случае HandlerFunc принимает только update, поэтому просто логируем.
			return next(update)
		}
	}
}

// generateCorrelationID генерирует уникальную строку с использованием crypto/rand.
func generateCorrelationID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// На случай ошибки возвращаем фиксированное значение
		return "unknown-corr-id"
	}
	return hex.EncodeToString(b)
}

// =======================
// RequestLoggingMiddleware
// =======================
// RequestLoggingMiddleware логирует входящее обновление до и после вызова обработчика.
// Это помогает в отладке и мониторинге.
func RequestLoggingMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			// Логируем входящее обновление
			logger.Info("RequestLoggingMiddleware: received update", core.Field{"update_id", update.UpdateID})
			err := next(update)
			if err != nil {
				logger.Error("RequestLoggingMiddleware: handler returned error", core.Field{"update_id", update.UpdateID}, core.Field{"error", err})
			} else {
				logger.Info("RequestLoggingMiddleware: handler executed successfully", core.Field{"update_id", update.UpdateID})
			}
			return err
		}
	}
}
