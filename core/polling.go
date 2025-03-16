package core

import (
	"context"
	"time"
)

// Poller – интерфейс для получения обновлений с поддержкой контекста.
type Poller interface {
	Start(ctx context.Context) error
	Stop() error
}

// pollingImpl – реализация поллинга, использующая контекст для корректного завершения.
type pollingImpl struct {
	api      BotAPI
	router   Router
	offset   int
	logger   Logger
	cancel   context.CancelFunc
}

// NewPoller создаёт новый экземпляр Poller с заданными API, роутером и логгером.
func NewPoller(api BotAPI, router Router, logger Logger) Poller {
	return &pollingImpl{
		api:    api,
		router: router,
		logger: logger,
	}
}

// Start запускает процесс поллинга с использованием переданного контекста.
// При отмене контекста цикл завершится корректно.
func (p *pollingImpl) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				p.logger.Info("Polling stopped due to context cancellation")
				return
			case <-ticker.C:
				updates, err := p.api.GetUpdates(ctx, p.offset, 100, 60)
				if err != nil {
					p.logger.Error("Error fetching updates", Field{"error", err})
					continue
				}
				for _, update := range updates {
					if err := p.router.Route(update); err != nil {
						p.logger.Error("Error routing update", Field{"error", err})
					}
					p.offset = update.UpdateID + 1
				}
			}
		}
	}()

	return nil
}

// Stop отменяет выполнение поллинга.
func (p *pollingImpl) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}
