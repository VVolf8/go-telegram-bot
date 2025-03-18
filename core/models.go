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
	Video    *Video    `json:"video,omitempty"`
	Audio    *Audio    `json:"audio,omitempty"`
	Contact  *Contact  `json:"contact,omitempty"`
	Location *Location `json:"location,omitempty"`
	// Можно добавить Document, Animation и т.д.
	Document *Document `json:"document,omitempty"`
	Animation *Animation `json:"animation,omitempty"`
}

// Chat представляет чат Telegram.
type Chat struct {
	ID int64 `json:"id"`
	// Дополнительные поля, если необходимо.
	Title     string `json:"title,omitempty"`
	Type      string `json:"type,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	// Дополнительные поля можно добавить по необходимости.
}

// Video представляет видео-сообщение.
type Video struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// Audio представляет аудиосообщение.
type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	Performer string `json:"performer,omitempty"`
	Title    string `json:"title,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// Contact представляет контакт пользователя.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name,omitempty"`
	UserID      int    `json:"user_id,omitempty"`
}

// Location представляет географическое местоположение.
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Document represents a document.
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// Animation represents an animation (e.g., GIF).
type Animation struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// User represents a Telegram user (including the bot itself).
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}
