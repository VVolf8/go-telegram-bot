# go-telegram-bot 

go-telegram-bot is a modular Go library for building Telegram bots. It provides a solid foundation for rapid bot development, including:

- A **Bot API client** with context support, logging, and error handling.
- **Polling and webhook** support for receiving Telegram updates.
- A **routing system** and middleware for pre-processing updates (e.g., authentication, logging, request tracing).
- Tools for building **inline keyboards and menus** for interactive messages.
- A **file manager** for uploading and downloading files via the Telegram API.
- An **in-memory cache** for temporary data storage.
- (Optional) **Metrics** (using Prometheus/OpenTelemetry) and **proxy support**.

## Features

- **Modular Architecture:** Each functional area is implemented in its own package, making it easy to extend and modify functionality.
- **Robust Logging and Error Handling:** All errors and panics are centrally managed using a custom logger and a `WithRecovery` function (in `core/foundation.go`).
- **Context Integration:** Every method accepts a `context.Context` to manage timeouts and cancellations.
- **Extended Functionality:** Supports sending messages, photos, documents, editing messages, answering callback queries, forwarding messages, and file operations.
- **Middleware Support:** Easily add pre-processing steps such as authentication, logging, and timing measurements.

## Project Structure
```bash
go-telegram-bot/
├── cache/ 
│ └── cache.go # In-memory cache implementation 
├── cmd/ 
│ └── testbot/ # A sample bot application for demonstrating the library features 
│   └── main.go 
├── core/ 
│ ├── bot.go # Telegram Bot API client (sending, editing, forwarding messages, etc.) 
│ ├── foundation.go # Logger and error handling (panic recovery) 
│ ├── models.go # Data models (Update, Message, Chat) 
│ ├── polling.go # Update polling mechanism 
│ └── router.go # Routing system for handling commands and callback queries 
├── files/ 
│ └── file.go # File manager for uploading and downloading files 
├── keyboards/ 
│ └── keyboards.go # Tools for building inline keyboards and menus 
├── metrics/ # (Optional) Metrics collection for bot monitoring 
│ └── metrics.go
├── middleware/ # Middleware for pre-processing updates 
│ └── middleware.go
├── proxy/ # (Optional) Support for using a proxy
│ └── proxy.go
├── webhooks/
  └── webhook.go
```

## Installation

To install the library, run:

```bash
go get github.com/VVolf8/go-telegram-bot
```

## Usage

A simple test bot example is provided in the cmd/testbot directory. It demonstrates:

- **Sending messages, photos, and documents.**
- **Forwarding and editing messages.**
- **Using inline keyboards.**
- **Working with an in-memory cache.**
- **Using middleware to process updates.**

Example code snippet:
```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/VVolf8/go-telegram-bot"
)

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        fmt.Println("TELEGRAM_BOT_TOKEN is not set")
        return
    }

    logger := core.NewLogger(core.DebugLevel)
    // Pass nil for the HTTP client to use the default one; for proxy support, provide a custom client.
    bot := core.NewBotClient(token, logger, nil)
    router := core.NewRouter(logger)

    // Example: inline keyboard usage for the /start command.
    router.HandleCommand("/start", func(update core.Update) error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        kbBuilder := keyboards.NewInlineKeyboardBuilder(logger)
        kbBuilder.AddRow(
            keyboards.InlineKeyboardButton{
                Text: "Visit Website",
                URL:  "https://example.com",
            },
            keyboards.InlineKeyboardButton{
                Text:         "Click Me",
                CallbackData: "callback_test",
            },
        )
        kbMarkup := kbBuilder.Build()
        return bot.SendMessageWithMarkup(ctx, update.Message.Chat.ID, "Welcome to the bot!", kbMarkup)
    })

    // Example: using the file manager to upload a document.
    fileManager := files.NewFileManager(token, logger, nil)
    router.HandleCommand("/test_senddocument", func(update core.Update) error {
        return fileManager.UploadFile(update.Message.Chat.ID, "test.txt", "Test Document Upload")
    })

    // Example: using the in-memory cache.
    memCache := cache.NewMemoryCache(logger)
    router.HandleCommand("/test_cache", func(update core.Update) error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        memCache.Set("foo", "bar")
        val, _ := memCache.Get("foo")
        memCache.Delete("foo")
        msg := fmt.Sprintf("Cache test: set foo=bar, retrieved value: %v, and deleted key", val)
        return bot.SendMessage(ctx, update.Message.Chat.ID, msg)
    })

    // Start polling for updates.
    poller := core.NewPoller(bot, router, logger)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    if err := poller.Start(ctx); err != nil {
        logger.Error("Error starting poller", core.Field{"error", err})
        return
    }

    // Graceful shutdown on termination signal.
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    logger.Info("Termination signal received, shutting down...")
    poller.Stop()
}
```

## Important Notice

This license is non-standard and modified.
Before using this library, please carefully review the full license terms to ensure that your intended use complies with the additional restrictions.
