package webhooks

import (
        "bytes"
        "context"
        "encoding/json"
        "fmt"
        "io/ioutil"
        "net/http"
        "time"

        "github.com/VVolf8/go-telegram-bot/core"
)

// WebhookManager – интерфейс для работы с вебхуками Telegram.
type WebhookManager interface {
        // SetWebhook устанавливает вебхук для бота по указанному URL.
        SetWebhook(ctx context.Context, webhookURL string) error
        // DeleteWebhook удаляет текущий вебхук.
        DeleteWebhook(ctx context.Context) error
        // ListenAndServe запускает HTTP-сервер для приёма обновлений через вебхук.
        // updateHandler вызывается для каждого полученного обновления.
        ListenAndServe(ctx context.Context, addr string, updateHandler func(ctx context.Context, update core.Update)) error
}

// webhookManager – реализация WebhookManager.
type webhookManager struct {
        token      string
        apiURL     string
        httpClient *http.Client
        logger     core.Logger
}

// NewWebhookManager создаёт новый экземпляр WebhookManager с использованием переданного токена и логгера.
func NewWebhookManager(token string, logger core.Logger) WebhookManager {
        if logger == nil {
                logger = core.NewDefaultLogger()
        }
        return &webhookManager{
                token:      token,
                apiURL:     fmt.Sprintf("https://api.telegram.org/bot%s", token),
                httpClient: &http.Client{},
                logger:     logger,
        }
}

// SetWebhook устанавливает вебхук для бота.
func (w *webhookManager) SetWebhook(ctx context.Context, webhookURL string) error {
        endpoint := fmt.Sprintf("%s/setWebhook", w.apiURL)
        payload := map[string]interface{}{
                "url": webhookURL,
        }
        body, err := json.Marshal(payload)
        if err != nil {
                w.logger.Error("Failed to marshal setWebhook payload", core.Field{"error", err})
                return err
        }

        req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
        if err != nil {
                w.logger.Error("Failed to create setWebhook request", core.Field{"error", err})
                return err
        }
        req.Header.Set("Content-Type", "application/json")

        var resp *http.Response
        core.WithRecovery(w.logger, func() {
                resp, err = w.httpClient.Do(req)
        })
        if err != nil {
                w.logger.Error("Error executing setWebhook request", core.Field{"error", err})
                return err
        }
        defer resp.Body.Close()

        respBody, _ := ioutil.ReadAll(resp.Body)
        if resp.StatusCode != http.StatusOK {
                w.logger.Error("Non-OK response from setWebhook",
                        core.Field{"status", resp.Status},
                        core.Field{"body", string(respBody)},
                )
                return fmt.Errorf("setWebhook failed with status: %s", resp.Status)
        }

        w.logger.Info("Webhook set successfully", core.Field{"webhook_url", webhookURL})
        return nil
}

// DeleteWebhook удаляет текущий вебхук.
func (w *webhookManager) DeleteWebhook(ctx context.Context) error {
        endpoint := fmt.Sprintf("%s/deleteWebhook", w.apiURL)
        req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
        if err != nil {
                w.logger.Error("Failed to create deleteWebhook request", core.Field{"error", err})
                return err
        }

        var resp *http.Response
        core.WithRecovery(w.logger, func() {
                resp, err = w.httpClient.Do(req)
        })
        if err != nil {
                w.logger.Error("Error executing deleteWebhook request", core.Field{"error", err})
                return err
        }
        defer resp.Body.Close()

        respBody, _ := ioutil.ReadAll(resp.Body)
        if resp.StatusCode != http.StatusOK {
                w.logger.Error("Non-OK response from deleteWebhook",
                        core.Field{"status", resp.Status},
                        core.Field{"body", string(respBody)},
                )
                return fmt.Errorf("deleteWebhook failed with status: %s", resp.Status)
        }

        w.logger.Info("Webhook deleted successfully")
        return nil
}

// ListenAndServe запускает HTTP-сервер для приёма обновлений через вебхук.
// updateHandler вызывается для каждого обновления, полученного в POST-запросе.
func (w *webhookManager) ListenAndServe(ctx context.Context, addr string, updateHandler func(ctx context.Context, update core.Update)) error {
        // Создаем мультиплексор для обработки запросов.
        mux := http.NewServeMux()
        mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
                // Обрабатываем только POST-запросы.
                if req.Method != http.MethodPost {
                        rw.WriteHeader(http.StatusMethodNotAllowed)
                        return
                }

                body, err := ioutil.ReadAll(req.Body)
                if err != nil {
                        w.logger.Error("Failed to read webhook request body", core.Field{"error", err})
                        rw.WriteHeader(http.StatusBadRequest)
                        return
                }
                defer req.Body.Close()

                var update core.Update
                if err := json.Unmarshal(body, &update); err != nil {
                        w.logger.Error("Failed to unmarshal webhook update", core.Field{"error", err})
                        rw.WriteHeader(http.StatusBadRequest)
                        return
                }

                w.logger.Info("Webhook update received", core.Field{"update_id", update.UpdateID})

                // Вызываем обработчик обновления с защитой от паники.
                core.WithRecovery(w.logger, func() {
                        updateHandler(req.Context(), update)
                })

                // Отправляем ответ Telegram.
                rw.WriteHeader(http.StatusOK)
                rw.Write([]byte("OK"))
        })

        server := &http.Server{
                Addr:    addr,
                Handler: mux,
        }

        // Запускаем сервер в отдельной горутине.
        go func() {
                w.logger.Info("Starting webhook server", core.Field{"addr", addr})
                if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                        w.logger.Error("Webhook server error", core.Field{"error", err})
                }
        }()

        // Ожидаем отмены контекста для корректного завершения работы сервера.
        <-ctx.Done()
        w.logger.Info("Shutting down webhook server")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return server.Shutdown(shutdownCtx)
}
