package core

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// LogLevel задаёт уровень логирования
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String возвращает строковое представление LogLevel
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field представляет собой пару ключ-значение для структурированных логов
type Field struct {
	Key   string
	Value interface{}
}

// Logger – интерфейс нашего логгера с методами для разных уровней логирования
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	// WithFields возвращает новый логгер с добавлением базовых полей
	WithFields(fields ...Field) Logger
}

// defaultLogger – реализация Logger, которая выводит логи в JSON-формате
type defaultLogger struct {
	mu         sync.Mutex
	level      LogLevel
	baseFields []Field
	out        *os.File
}

// NewLogger создаёт новый логгер с заданным уровнем логирования (например, DebugLevel или InfoLevel)
func NewLogger(level LogLevel) Logger {
	return &defaultLogger{
		level: level,
		out:   os.Stdout,
	}
}

// WithFields возвращает новый логгер с добавлением указанных полей к базовым
func (l *defaultLogger) WithFields(fields ...Field) Logger {
	newBaseFields := make([]Field, len(l.baseFields))
	copy(newBaseFields, l.baseFields)
	newBaseFields = append(newBaseFields, fields...)
	return &defaultLogger{
		level:      l.level,
		baseFields: newBaseFields,
		out:        l.out,
	}
}

// logf выполняет форматирование и вывод лог-сообщения с указанным уровнем и полями
func (l *defaultLogger) logf(level LogLevel, msg string, fields ...Field) {
	// Если текущий уровень меньше требуемого, лог не выводится
	if level < l.level {
		return
	}
	entry := make(map[string]interface{})
	entry["time"] = time.Now().Format(time.RFC3339)
	entry["level"] = level.String()
	entry["message"] = msg

	// Объединяем базовые поля и поля, переданные в вызове
	for _, field := range l.baseFields {
		entry[field.Key] = field.Value
	}
	for _, field := range fields {
		entry[field.Key] = field.Value
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	// Преобразуем запись в JSON и выводим в заданный поток
	b, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.out, "Error marshaling log entry: %v\n", err)
	} else {
		fmt.Fprintln(l.out, string(b))
	}
	// Если уровень Fatal, завершаем выполнение программы
	if level == FatalLevel {
		os.Exit(1)
	}
}

func (l *defaultLogger) Debug(msg string, fields ...Field) {
	l.logf(DebugLevel, msg, fields...)
}

func (l *defaultLogger) Info(msg string, fields ...Field) {
	l.logf(InfoLevel, msg, fields...)
}

func (l *defaultLogger) Warn(msg string, fields ...Field) {
	l.logf(WarnLevel, msg, fields...)
}

func (l *defaultLogger) Error(msg string, fields ...Field) {
	l.logf(ErrorLevel, msg, fields...)
}

func (l *defaultLogger) Fatal(msg string, fields ...Field) {
	l.logf(FatalLevel, msg, fields...)
}

// WithRecovery выполняет переданную функцию fn и перехватывает панику, если она возникнет,
// записывая подробное сообщение с информацией о панике и стеком вызовов.
func WithRecovery(logger Logger, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			logger.Error("Panic recovered",
				Field{"panic", r},
				Field{"stack", stack},
			)
		}
	}()
	fn()
}

// NewDefaultLogger создаёт логгер с уровнем Debug по умолчанию.
func NewDefaultLogger() Logger {
	return NewLogger(DebugLevel)
}
