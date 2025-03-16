package proxy

import (
	"net/http"
	"net/url"
	"time"
)

// NewHTTPClientWithProxy создаёт HTTP-клиент, использующий указанный прокси-адрес.
// proxyURL должен быть в формате "http://user:pass@proxy.example.com:port" или "http://proxy.example.com:port".
func NewHTTPClientWithProxy(proxyURL string) (*http.Client, error) {
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
		// Можно добавить дополнительные настройки, например, таймауты, пул соединений и т.д.
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	return client, nil
}
