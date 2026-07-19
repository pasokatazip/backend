package fincode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

const (
	defaultAPIVersion = "20211001"
	maxResponseBytes  = 1 << 20
)

var ErrInvalidConfig = errors.New("invalid fincode client config")

type Config struct {
	BaseURL    string
	SecretKey  string
	APIVersion string
	HTTPClient *http.Client
}

type Client struct {
	baseURL    string
	secretKey  string
	apiVersion string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("fincode API request failed: status=%d body=%s", e.StatusCode, e.Body)
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	secretKey := strings.TrimSpace(config.SecretKey)
	if baseURL == "" || secretKey == "" {
		return nil, ErrInvalidConfig
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		return nil, ErrInvalidConfig
	}

	apiVersion := strings.TrimSpace(config.APIVersion)
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &Client{
		baseURL:    baseURL,
		secretKey:  secretKey,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}, nil
}

func (c *Client) doJSON(
	ctx context.Context,
	method string,
	path string,
	idempotencyKey string,
	requestBody any,
	responseBody any,
) error {
	var body []byte
	var err error
	if requestBody != nil {
		body, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("%w: encode fincode request: %v", domain.ErrInternal, err)
		}
	}

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("%w: create fincode request: %v", domain.ErrInternal, err)
	}

	request.Header.Set("Authorization", "Bearer "+c.secretKey)
	request.Header.Set("Api-Version", c.apiVersion)
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	if idempotencyKey != "" {
		request.Header.Set("idempotent_key", idempotencyKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: send fincode request: %v", domain.ErrExternalService, err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("%w: read fincode response: %v", domain.ErrExternalService, err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{
			StatusCode: response.StatusCode,
			Body:       strings.TrimSpace(string(responseBytes)),
		}
		return fmt.Errorf("%w: %w", domain.ErrExternalService, apiErr)
	}

	if responseBody == nil || len(responseBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(responseBytes, responseBody); err != nil {
		return fmt.Errorf("%w: decode fincode response: %v", domain.ErrExternalService, err)
	}

	return nil
}
