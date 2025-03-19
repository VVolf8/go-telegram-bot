package core

import (
	"context"
 "crypto/rand"
 "encoding/hex"
	"time"
)

// WithTimeoutAndCorrelation создает контекст с заданным таймаутом и добавляет correlation ID в контекст.
func WithTimeoutAndCorrelation(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// Можно расширить этот функционал, добавив генерацию correlation ID, если он еще не задан.
	ctx, cancel := context.WithTimeout(parent, timeout)
	// Например, если не установлен, можно установить в контекст:
	if ctx.Value("correlation_id") == nil {
		ctx = context.WithValue(ctx, "correlation_id", generateCorrelationID())
	}
	return ctx, cancel
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