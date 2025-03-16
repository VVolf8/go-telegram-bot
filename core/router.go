package core

// HandlerFunc – функция-обработчик для обновления.
// Возвращает ошибку, если обработка обновления завершилась неудачно.
type HandlerFunc func(update Update) error

// Router – интерфейс для маршрутизации обновлений.
type Router interface {
	// HandleCommand регистрирует обработчик для команд (например, "/start").
	HandleCommand(command string, handler HandlerFunc)
	// HandleCallback регистрирует обработчик для колбэков.
	HandleCallback(callbackData string, handler HandlerFunc)
	// Route определяет, какой обработчик должен обработать переданное обновление.
	Route(update Update) error
}

// simpleRouter – простая реализация роутера.
type simpleRouter struct {
	commandHandlers  map[string]HandlerFunc
	callbackHandlers map[string]HandlerFunc
	logger           Logger
}

// NewRouter создаёт новый экземпляр роутера с использованием переданного логгера.
func NewRouter(logger Logger) Router {
	return &simpleRouter{
		commandHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]HandlerFunc),
		logger:           logger,
	}
}

// HandleCommand регистрирует обработчик для указанной команды.
func (r *simpleRouter) HandleCommand(command string, handler HandlerFunc) {
	r.commandHandlers[command] = handler
	r.logger.Debug("Registered command handler", Field{"command", command})
}

// HandleCallback регистрирует обработчик для указанного callback-данных.
func (r *simpleRouter) HandleCallback(callbackData string, handler HandlerFunc) {
	r.callbackHandlers[callbackData] = handler
	r.logger.Debug("Registered callback handler", Field{"callback_data", callbackData})
}

// Route выполняет маршрутизацию обновления.
// Если обновление содержит сообщение с командой, ищется соответствующий обработчик.
// Обработчик вызывается в блоке с механизмом перехвата паники.
func (r *simpleRouter) Route(update Update) error {
	var err error

	// Обработка текстового сообщения (например, команды, начинающиеся с '/')
	if update.Message != nil {
		text := update.Message.Text
		if len(text) > 0 && text[0] == '/' {
			if handler, exists := r.commandHandlers[text]; exists {
				// Вызов обработчика в защищённом блоке с перехватом паники.
				WithRecovery(r.logger, func() {
					err = handler(update)
				})
				if err != nil {
					r.logger.Error("Error handling command",
						Field{"command", text},
						Field{"error", err},
					)
					return err
				}
				r.logger.Info("Handled command successfully", Field{"command", text})
			} else {
				r.logger.Warn("No handler registered for command", Field{"command", text})
			}
		} else {
			r.logger.Debug("Received message without command", Field{"text", text})
		}
	}

	// Если в будущем понадобится обрабатывать callback'и, можно добавить обработку аналогичным образом:
	/*
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			if handler, exists := r.callbackHandlers[data]; exists {
				WithRecovery(r.logger, func() {
					err = handler(update)
				})
				if err != nil {
					r.logger.Error("Error handling callback",
						Field{"callback_data", data},
						Field{"error", err},
					)
					return err
				}
				r.logger.Info("Handled callback successfully", Field{"callback_data", data})
			} else {
				r.logger.Warn("No handler registered for callback", Field{"callback_data", data})
			}
		}
	*/

	return nil
}
