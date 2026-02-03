package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/plastinin/docrecognizer/internal/config"
	"go.uber.org/zap"
)

// OllamaClient клиент для работы с Ollama API
type OllamaClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
	logger     *zap.Logger
}

// NewOllamaClient создаёт новый экземпляр OllamaClient
func NewOllamaClient(cfg config.OllamaConfig, logger *zap.Logger) *OllamaClient {
	return &OllamaClient{
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		baseURL: cfg.Host,
		model:   cfg.Model,
		logger:  logger,
	}
}

// ollamaRequest структура запроса к Ollama API
type ollamaRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	Images   []string `json:"images,omitempty"` // Base64 encoded images
	Stream   bool     `json:"stream"`
	Format   string   `json:"format,omitempty"` // "json" для JSON output
	Options  *ollamaOptions `json:"options,omitempty"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// ollamaResponse структура ответа от Ollama API
type ollamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Error     string `json:"error,omitempty"`
}

// RecognizeDocument распознаёт документ и извлекает данные по схеме
// RecognizeDocument распознаёт документ с помощью vision модели
func (c *OllamaClient) RecognizeDocument(ctx context.Context, imageData []byte, contentType string, schema []string) (map[string]any, error) {
	c.logger.Debug("Starting document recognition",
		zap.String("model", c.model),
		zap.Int("image_size", len(imageData)),
		zap.Strings("schema", schema),
	)

	// Формируем промпт
	prompt := c.buildPrompt(schema)

	// Кодируем изображение в base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// Формируем запрос для /api/chat (vision модели)
	reqBody := map[string]any{
		"model": c.model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": prompt,
				"images":  []string{imageBase64},
			},
		},
		"stream": false,
		"format": "json",
		"options": map[string]any{
			"temperature": 0.1,
			"num_predict": 2048,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Отправляем запрос к /api/chat
	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("Ollama request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("status_code", resp.StatusCode),
	)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ
	var chatResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", chatResp.Error)
	}

	// Логируем ответ модели для отладки
	c.logger.Debug("Raw LLM response", zap.String("response", chatResp.Message.Content))

	// Парсим JSON из ответа
	result, err := c.parseResponse(chatResp.Message.Content, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

// buildPrompt формирует промпт для распознавания документа
func (c *OllamaClient) buildPrompt(schema []string) string {
	fieldsJSON, _ := json.Marshal(schema)
	
	prompt := fmt.Sprintf(`You are a document recognition assistant. Analyze the provided document image and extract the requested information.

TASK: Extract the following fields from the document:
%s

INSTRUCTIONS:
1. Carefully analyze the document image
2. Extract values for each requested field
3. If a field is not found or not applicable, use null
4. For dates, use ISO 8601 format (YYYY-MM-DD)
5. For monetary amounts, extract the numeric value only
6. Return ONLY valid JSON, no additional text

RESPONSE FORMAT:
Return a JSON object with the requested fields as keys and extracted values.

Example for fields ["invoice_number", "date", "total_amount"]:
{"invoice_number": "INV-2024-001", "date": "2024-01-15", "total_amount": 1500.00}

Now analyze the document and extract: %s`, string(fieldsJSON), strings.Join(schema, ", "))

	return prompt
}

// parseResponse парсит ответ LLM и извлекает JSON
func (c *OllamaClient) parseResponse(response string, schema []string) (map[string]any, error) {
	// Очищаем ответ от возможных markdown блоков
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Пытаемся найти JSON в ответе
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	
	if startIdx == -1 || endIdx == -1 || startIdx > endIdx {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[startIdx : endIdx+1]

	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Проверяем, что все запрошенные поля присутствуют (добавляем null если нет)
	for _, field := range schema {
		if _, ok := result[field]; !ok {
			result[field] = nil
		}
	}

	return result, nil
}

// CheckHealth проверяет доступность Ollama
func (c *OllamaClient) CheckHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// CheckModel проверяет, что модель загружена
func (c *OllamaClient) CheckModel(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, model := range tagsResp.Models {
		if strings.HasPrefix(model.Name, strings.Split(c.model, ":")[0]) {
			c.logger.Info("Model found", zap.String("model", model.Name))
			return nil
		}
	}

	return fmt.Errorf("model %s not found, please run: ollama pull %s", c.model, c.model)
}