package middleware

import (
	"time"

	"gotgbot/core"
)

// MiddlewareFunc определяет функцию middleware, которая принимает и возвращает HandlerFunc.
type MiddlewareFunc func(core.HandlerFunc) core.HandlerFunc

// ComposeMiddleware применяет цепочку middleware к базовому обработчику.
func ComposeMiddleware(handler core.HandlerFunc, mws ...MiddlewareFunc) core.HandlerFunc {
	// Применяем middleware в обратном порядке, чтобы первый в списке оборачивал последующие.
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// LoggingMiddleware регистрирует информацию о входящем обновлении и об ошибках выполнения обработчика.
func LoggingMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			logger.Info("Middleware: Received update", core.Field{"update_id", update.UpdateID})
			err := next(update)
			if err != nil {
				logger.Error("Middleware: Error in handler", core.Field{"update_id", update.UpdateID}, core.Field{"error", err})
			}
			return err
		}
	}
}

// AuthMiddleware – пример middleware для аутентификации (пока реализован как заглушка).
// Здесь можно добавить проверку, например, по userID или другим признакам.
func AuthMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			// Пример: пропускаем все обновления. В будущем можно добавить реальные проверки.
			logger.Debug("Middleware: Auth check passed", core.Field{"update_id", update.UpdateID})
			return next(update)
		}
	}
}

// TimingMiddleware измеряет время выполнения следующего обработчика.
func TimingMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) error {
			start := time.Now()
			err := next(update)
			duration := time.Since(start)
			logger.Info("Middleware: Handler execution time", core.Field{"duration", duration}, core.Field{"update_id", update.UpdateID})
			return err
		}
	}
}

// RecoveryMiddleware перехватывает панику в цепочке middleware, используя нашу функцию WithRecovery.
func RecoveryMiddleware(logger core.Logger) MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(update core.Update) (err error) {
			core.WithRecovery(logger, func() {
				err = next(update)
			})
			return err
		}
	}
}
