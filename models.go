package core

// Update представляет обновление от Telegram.
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
	// Можно добавить и другие поля, например, CallbackQuery, если требуется.
}

// Message представляет сообщение Telegram.
type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
	// Дополнительные поля, если необходимо.
}

// Chat представляет чат Telegram.
type Chat struct {
	ID int64 `json:"id"`
	// Дополнительные поля, если необходимо.
}
