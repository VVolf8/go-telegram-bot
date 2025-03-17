package files

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "io/ioutil"
        "mime/multipart"
        "net/http"
        "os"
        "path/filepath"

        "github.com/VVolf8/go-telegram-bot/core"
)

// File представляет собой файл, возвращаемый Telegram API.
type File struct {
        FileID   string `json:"file_id"`
        FilePath string `json:"file_path"`
        // Дополнительные поля можно добавить при необходимости.
}

// FileManager определяет интерфейс для работы с файлами.
type FileManager interface {
        UploadFile(chatID int64, filePath, caption string) error
        DownloadFile(fileID string) ([]byte, error)
}

// fileManager – реализация FileManager.
type fileManager struct {
        token      string
        httpClient *http.Client
        logger     core.Logger
}

// NewFileManager создаёт новый FileManager с заданным токеном, логгером и HTTP-клиентом.
func NewFileManager(token string, logger core.Logger, httpClient *http.Client) FileManager {
        if httpClient == nil {
                httpClient = &http.Client{}
        }
        return &fileManager{
                token:      token,
                httpClient: httpClient,
                logger:     logger,
        }
}

// UploadFile загружает файл (например, документ) в Telegram, отправляя его через multipart/form-data.
func (fm *fileManager) UploadFile(chatID int64, filePath, caption string) error {
        endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", fm.token)

        file, err := os.Open(filePath)
        if err != nil {
                fm.logger.Error("Failed to open file", core.Field{"error", err}, core.Field{"filePath", filePath})
                return err
        }
        defer file.Close()

        // Создаем multipart-форму.
        var requestBody bytes.Buffer
        writer := multipart.NewWriter(&requestBody)

        // Добавляем поле chat_id.
        err = writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
        if err != nil {
                fm.logger.Error("Failed to write chat_id field", core.Field{"error", err})
                return err
        }

        // Если указан caption, добавляем его.
        if caption != "" {
                err = writer.WriteField("caption", caption)
                if err != nil {
                        fm.logger.Error("Failed to write caption field", core.Field{"error", err})
                        return err
                }
        }

        // Создаем поле файла.
        part, err := writer.CreateFormFile("document", filepath.Base(filePath))
        if err != nil {
                fm.logger.Error("Failed to create form file", core.Field{"error", err})
                return err
        }
        _, err = io.Copy(part, file)
        if err != nil {
                fm.logger.Error("Failed to copy file content", core.Field{"error", err})
                return err
        }

        // Завершаем запись multipart-формы.
        if err = writer.Close(); err != nil {
                fm.logger.Error("Failed to close writer", core.Field{"error", err})
                return err
        }

        req, err := http.NewRequest("POST", endpoint, &requestBody)
        if err != nil {
                fm.logger.Error("Failed to create upload request", core.Field{"error", err})
                return err
        }
        req.Header.Set("Content-Type", writer.FormDataContentType())

        var resp *http.Response
        core.WithRecovery(fm.logger, func() {
                resp, err = fm.httpClient.Do(req)
        })
        if err != nil {
                fm.logger.Error("Error during file upload", core.Field{"error", err})
                return err
        }
        defer resp.Body.Close()

        respBody, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                fm.logger.Error("Failed to read upload response", core.Field{"error", err})
                return err
        }

        if resp.StatusCode != http.StatusOK {
                fm.logger.Error("Non-OK response during file upload",
                        core.Field{"status", resp.Status},
                        core.Field{"body", string(respBody)},
                )
                return fmt.Errorf("upload failed with status: %s", resp.Status)
        }

        fm.logger.Info("File uploaded successfully", core.Field{"chat_id", chatID}, core.Field{"file", filePath})
        return nil
}

// DownloadFile скачивает файл по file_id. Сначала вызывается getFile для получения пути, затем происходит скачивание.
func (fm *fileManager) DownloadFile(fileID string) ([]byte, error) {
        // Шаг 1. Вызов getFile для получения file_path.
        getFileURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", fm.token, fileID)
        req, err := http.NewRequest("GET", getFileURL, nil)
        if err != nil {
                fm.logger.Error("Failed to create getFile request", core.Field{"error", err})
                return nil, err
        }

        var resp *http.Response
        core.WithRecovery(fm.logger, func() {
                resp, err = fm.httpClient.Do(req)
        })
        if err != nil {
                fm.logger.Error("Error executing getFile request", core.Field{"error", err})
                return nil, err
        }
        defer resp.Body.Close()

        respBody, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                fm.logger.Error("Failed to read getFile response", core.Field{"error", err})
                return nil, err
        }

        if resp.StatusCode != http.StatusOK {
                fm.logger.Error("Non-OK response from getFile",
                        core.Field{"status", resp.Status},
                        core.Field{"body", string(respBody)},
                )
                return nil, fmt.Errorf("getFile failed with status: %s", resp.Status)
        }

        var result struct {
                OK     bool `json:"ok"`
                Result File `json:"result"`
        }
        if err = json.Unmarshal(respBody, &result); err != nil {
                fm.logger.Error("Failed to unmarshal getFile response", core.Field{"error", err})
                return nil, err
        }
        if !result.OK || result.Result.FilePath == "" {
                fm.logger.Error("Telegram API returned error for getFile", core.Field{"response", string(respBody)})
                return nil, fmt.Errorf("telegram API error: %s", string(respBody))
        }

        // Шаг 2. Скачивание файла по полученному пути.
        downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", fm.token, result.Result.FilePath)
        reqDownload, err := http.NewRequest("GET", downloadURL, nil)
        if err != nil {
                fm.logger.Error("Failed to create download request", core.Field{"error", err})
                return nil, err
        }

        var downloadResp *http.Response
        core.WithRecovery(fm.logger, func() {
                downloadResp, err = fm.httpClient.Do(reqDownload)
        })
        if err != nil {
                fm.logger.Error("Error executing download request", core.Field{"error", err})
                return nil, err
        }
        defer downloadResp.Body.Close()

        fileData, err := ioutil.ReadAll(downloadResp.Body)
        if err != nil {
                fm.logger.Error("Failed to read downloaded file", core.Field{"error", err})
                return nil, err
        }

        fm.logger.Info("File downloaded successfully", core.Field{"file_id", fileID})
        return fileData, nil
}
