# go-telegram-bot

**go-telegram-bot** is a modular Go library for building Telegram bots that enables rapid development of sophisticated bots with rich functionality. The library provides robust support for the Telegram Bot API, along with extended features for handling payments, metrics, proxy support, and webhooks.

## Features

- **Modular Architecture:**  
  Each functional block is implemented in its own package, making it easy to extend and modify the functionality.

- **Telegram Bot API Integration:**  
  A client with context support, centralized logging, and error handling is provided. It supports sending messages, photos, documents, editing messages, answering callback queries, forwarding messages, and chat management.

- **Update Handling:**  
  Supports both polling and webhook modes. A router directs updates to handlers for commands, callback queries, and file types (documents, animations, etc.).

- **Caching:**  
  An in-memory cache is implemented for temporary data storage, with thread-safe access and logging of cache operations.

- **Keyboards:**  
  Tools for creating both inline and reply keyboards using builder patterns, which simplifies generating interactive messages.

- **Payments Module:**  
  The `payments` module offers a full suite of functions for handling payments via the Telegram API:
  - Sending invoices
  - Answering shipping queries
  - Answering pre‑checkout queries
  - Handling successful payments  
  This allows you to integrate a payment system into your bot with minimal effort.

- **Metrics:**  
  Integration with Prometheus collects metrics (number of messages sent, operation latency, error counts) to help monitor and analyze bot performance.

- **Middleware:**  
  Support for middleware chains enables pre-processing of updates: logging, timing, access control, tracing (with correlation IDs), and panic recovery.

- **Additional Capabilities:**  
  Includes support for proxy servers and webhook mode, letting you choose between polling and webhook update handling based on your infrastructure.

- **Licensing:**  
  Distributed under the Apache License 2.0 with an additional clause:  
  **This software CANNOT be used in projects that require paid access for end users without prior written permission from the author.** For commercial usage, please contact the author.

## Project Structure

```bash
go-telegram-bot/
├── cache/ 
│   └── cache.go            # In-memory cache implementation
├── cmd/ 
│   ├── testbot/            # Example bot demonstrating library features
│   │   └── main.go
│   └── testpayments/       # Test bot for the payments module
│       └── payments_main.go
├── core/ 
│   ├── bot.go              # Telegram Bot API client: sending messages, etc.
│   ├── foundation.go       # Logging, error handling, and panic recovery
│   ├── models.go           # Data models (Update, Message, Chat, etc.)
│   ├── polling.go          # Update polling mechanism
│   └── router.go           # Routing updates to handlers
├── files/ 
│   └── files.go            # File management: uploading and downloading via Telegram API
├── keyboards/ 
│   ├── keyboards.go        # Tools for building inline keyboards
│   └── reply.go            # Tools for building reply keyboards
├── metrics/ 
│   └── metrics.go          # Prometheus integration for metrics collection
├── middleware/ 
│   ├── advanced.go         # Advanced middleware (security, tracing, logging)
│   └── middleware.go       # Core middleware (authentication, timing, recovery)
├── payments/ 
│   ├── payments.go         # Payment-related functions (invoices, shipping queries, etc.)
│   └── payments_test.go    # Tests and examples for the payments module
├── proxy/ 
│   └── proxy.go            # Optional support for proxy servers
└── webhooks/
    └── webhook.go          # Webhook support for receiving updates
```

## Installation

To install the library, run:

```bash
go get github.com/VVolf8/go-telegram-bot
```

Ensure that you have Go version 1.24 or higher installed.

## Quick Start

Below is an example of a simple bot using the core features of the library:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/VVolf8/go-telegram-bot/core"
    "github.com/VVolf8/go-telegram-bot/keyboards"
)

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        fmt.Println("TELEGRAM_BOT_TOKEN is not set")
        return
    }

    logger := core.NewLogger(core.DebugLevel)
    bot := core.NewBotClient(token, logger, nil)
    router := core.NewRouter(logger)

    // Handle the /start command using an inline keyboard
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

    // Start polling for updates
    poller := core.NewPoller(bot, router, logger)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go func() {
        if err := poller.Start(ctx); err != nil {
            logger.Error("Error starting poller", core.Field{"error", err})
        }
    }()

    logger.Info("Bot is running. Press Ctrl+C to exit.")
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    logger.Info("Shutting down...")
    poller.Stop()
}
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

## Using the Payments Module

The updated version includes a payments module that allows you to work with Telegram API payments. Below is an example test bot for the payments module:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/VVolf8/go-telegram-bot/core"
    "github.com/VVolf8/go-telegram-bot/payments"
)

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        fmt.Println("TELEGRAM_BOT_TOKEN is not set")
        return
    }

    logger := core.NewLogger(core.DebugLevel)
    bot := core.NewBotClient(token, logger, nil)

    // Initialize PaymentService (implementation provided in payments.NewPaymentService)
    paymentService := payments.NewPaymentService(token, logger, nil)

    router := core.NewRouter(logger)

    // Test sending an invoice
    router.HandleCommand("/test_invoice", func(update core.Update) error {
        invoice := payments.Invoice{
            ChatID:         update.Message.Chat.ID,
            Title:          "Test Invoice",
            Description:    "Test invoice description for payment processing",
            Payload:        "test_payload",
            ProviderToken:  "provider_token_example", // replace with a real provider token
            StartParameter: "start_param",
            Currency:       "USD",
            Prices: []payments.Price{
                {Label: "Test Price", Amount: 1000},
            },
        }
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        err := paymentService.SendInvoice(ctx, invoice)
        if err != nil {
            return err
        }
        return bot.SendMessage(ctx, update.Message.Chat.ID, "Invoice sent successfully!")
    })

    // Test answering a shipping query
    router.HandleCommand("/test_shipping", func(update core.Update) error {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        err := paymentService.AnswerShippingQuery(ctx, "shipping_query_id_example", true, "")
        if err != nil {
            return err
        }
        return bot.SendMessage(ctx, update.Message.Chat.ID, "Shipping query processed successfully!")
    })

    // Test answering a pre‑checkout query
    router.HandleCommand("/test_precheckout", func(update core.Update) error {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        err := paymentService.AnswerPreCheckoutQuery(ctx, "precheckout_query_id_example", true, "")
        if err != nil {
            return err
        }
        return bot.SendMessage(ctx, update.Message.Chat.ID, "Pre‑checkout query processed successfully!")
    })

    // Test handling a successful payment
    router.HandleCommand("/test_success_payment", func(update core.Update) error {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        err := paymentService.HandleSuccessfulPayment(ctx, update)
        if err != nil {
            return err
        }
        return bot.SendMessage(ctx, update.Message.Chat.ID, "Successful payment processed!")
    })

    // Start polling for updates
    poller := core.NewPoller(bot, router, logger)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go func() {
        if err := poller.Start(ctx); err != nil {
            logger.Error("Error starting poller", core.Field{"error", err})
        }
    }()

    logger.Info("Test bot for payments is running. Awaiting commands...")
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    logger.Info("Received shutdown signal, stopping bot...")
    poller.Stop()
}
```

## Additional Sections
- **Metrics:**
The metrics module collects statistics about the bot's performance. To expose metrics, run an HTTP server:

```go
package main

import (
    "github.com/VVolf8/go-telegram-bot/metrics"
)

func main() {
    // Run the metrics server on port 2112
    if err := metrics.ExposeMetricsHandler(":2112"); err != nil {
        panic(err)
    }
}
```

- **Middleware:**
Use middleware (e.g., Logging, Auth, Timing, Tracing, and Recovery) to wrap your handlers and add pre- and post-processing logic.

- **Proxy and Webhooks:**
If you need to work via a proxy or use webhook mode, refer to the proxy and webhooks modules.

## License
Licensed under the Apache License 2.0 with an additional clause:
This software CANNOT be used in projects that require paid access for end users without prior written permission from the author. For commercial usage, please contact the author.

## Conclusion
go-telegram-bot provides a comprehensive suite of tools for developing Telegram bots—from basic messaging to handling payments and monitoring metrics. Its modular architecture and well-thought-out design enable developers to quickly configure and extend bot functionality to meet any requirements.
