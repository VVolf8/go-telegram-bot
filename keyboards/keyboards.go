package keyboards

import (
        "context"
        "encoding/json"

        "github.com/VVolf8/go-telegram-bot/core"
)

// InlineKeyboardButton описывает кнопку для inline клавиатуры.
type InlineKeyboardButton struct {
        Text         string `json:"text"`
        URL          string `json:"url,omitempty"`
        CallbackData string `json:"callback_data,omitempty"`
}

// InlineKeyboardMarkup описывает разметку inline клавиатуры.
type InlineKeyboardMarkup struct {
        InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// NewInlineKeyboardMarkup создаёт новую пустую разметку для клавиатуры.
func NewInlineKeyboardMarkup() *InlineKeyboardMarkup {
        return &InlineKeyboardMarkup{
                InlineKeyboard: make([][]InlineKeyboardButton, 0),
        }
}

// InlineKeyboardBuilder предоставляет функциональный подход для создания inline клавиатур.
type InlineKeyboardBuilder struct {
        markup *InlineKeyboardMarkup
        logger core.Logger
}

// NewInlineKeyboardBuilder возвращает новый билдер для создания клавиатуры с использованием переданного логгера.
func NewInlineKeyboardBuilder(logger core.Logger) *InlineKeyboardBuilder {
        if logger == nil {
                // Если логгер не предоставлен, используем дефолтный.
                logger = core.NewDefaultLogger()
        }
        return &InlineKeyboardBuilder{
                markup: NewInlineKeyboardMarkup(),
                logger: logger,
        }
}

// AddRow добавляет ряд кнопок в клавиатуру. Если ряд пустой, регистрируется предупреждение.
func (b *InlineKeyboardBuilder) AddRow(buttons ...InlineKeyboardButton) *InlineKeyboardBuilder {
        if len(buttons) == 0 {
                b.logger.Warn("Попытка добавить пустой ряд в клавиатуру")
                return b
        }
        b.markup.InlineKeyboard = append(b.markup.InlineKeyboard, buttons)
        b.logger.Debug("Добавлен ряд кнопок", core.Field{"buttons_count", len(buttons)})
        return b
}

// Build возвращает собранную разметку клавиатуры и логирует результат.
func (b *InlineKeyboardBuilder) Build() *InlineKeyboardMarkup {
        b.logger.Info("Клавиатура построена", core.Field{"rows_count", len(b.markup.InlineKeyboard)})
        return b.markup
}

// ToJSON преобразует разметку клавиатуры в JSON с поддержкой контекста.
// Это полезно при передаче клавиатуры через запросы к API Telegram.
func (ikm *InlineKeyboardMarkup) ToJSON(ctx context.Context) ([]byte, error) {
        // Проверяем, не отменён ли контекст.
        select {
        case <-ctx.Done():
                return nil, ctx.Err()
        default:
        }
        return json.Marshal(ikm)
}
