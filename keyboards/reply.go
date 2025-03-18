package keyboards

import (
	"context"
	"encoding/json"

	"github.com/VVolf8/go-telegram-bot/core"
)

// ReplyKeyboardButton представляет собой кнопку для reply‑клавиатуры.
type ReplyKeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact,omitempty"`
	RequestLocation bool   `json:"request_location,omitempty"`
}

// ReplyKeyboardMarkup описывает разметку reply‑клавиатуры.
type ReplyKeyboardMarkup struct {
	Keyboard        [][]ReplyKeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool                    `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool                    `json:"one_time_keyboard,omitempty"`
	Selective       bool                    `json:"selective,omitempty"`
}

// ReplyKeyboardBuilder предоставляет удобный способ создания reply‑клавиатуры.
type ReplyKeyboardBuilder struct {
	markup *ReplyKeyboardMarkup
	logger core.Logger
}

// NewReplyKeyboardBuilder создаёт новый билдера reply‑клавиатуры.
func NewReplyKeyboardBuilder(logger core.Logger) *ReplyKeyboardBuilder {
	if logger == nil {
		logger = core.NewDefaultLogger()
	}
	return &ReplyKeyboardBuilder{
		markup: &ReplyKeyboardMarkup{
			Keyboard: make([][]ReplyKeyboardButton, 0),
		},
		logger: logger,
	}
}

// AddRow добавляет ряд кнопок в клавиатуру.
func (b *ReplyKeyboardBuilder) AddRow(buttons ...ReplyKeyboardButton) *ReplyKeyboardBuilder {
	if len(buttons) == 0 {
		b.logger.Warn("Attempted to add an empty row to reply keyboard")
		return b
	}
	b.markup.Keyboard = append(b.markup.Keyboard, buttons)
	b.logger.Debug("Added row to reply keyboard", core.Field{"buttons_count", len(buttons)})
	return b
}

// SetResizeKeyboard задаёт опцию ResizeKeyboard.
func (b *ReplyKeyboardBuilder) SetResizeKeyboard(resize bool) *ReplyKeyboardBuilder {
	b.markup.ResizeKeyboard = resize
	return b
}

// SetOneTimeKeyboard задаёт опцию OneTimeKeyboard.
func (b *ReplyKeyboardBuilder) SetOneTimeKeyboard(oneTime bool) *ReplyKeyboardBuilder {
	b.markup.OneTimeKeyboard = oneTime
	return b
}

// SetSelective задаёт опцию Selective.
func (b *ReplyKeyboardBuilder) SetSelective(selective bool) *ReplyKeyboardBuilder {
	b.markup.Selective = selective
	return b
}

// Build возвращает собранную разметку клавиатуры.
func (b *ReplyKeyboardBuilder) Build() *ReplyKeyboardMarkup {
	b.logger.Info("Reply keyboard built", core.Field{"rows_count", len(b.markup.Keyboard)})
	return b.markup
}

// ToJSON преобразует разметку в JSON с поддержкой контекста.
func (rk *ReplyKeyboardMarkup) ToJSON(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return json.Marshal(rk)
}
