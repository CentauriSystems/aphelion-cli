package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	ErrorMsg   string `json:"error"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	return fmt.Sprintf("API error: %d", e.StatusCode)
}

func NewClient() *Client {
	return &Client{
		baseURL: config.GetAPIUrl(),
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (c *Client) buildURL(endpoint string) string {
	baseURL := strings.TrimSuffix(c.baseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}

func (c *Client) request(method, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	url := c.buildURL(endpoint)
	
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token := config.GetAccessToken(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var rawMessage string
		if respBody, err := io.ReadAll(resp.Body); err == nil {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				if apiErr.Message != "" {
					rawMessage = apiErr.Message
				} else {
					rawMessage = apiErr.ErrorMsg
				}
			} else {
				rawMessage = strings.TrimSpace(string(respBody))
			}
		}
		return nil, toActionableError(resp.StatusCode, endpoint, rawMessage)
	}

	return resp, nil
}

func (c *Client) Get(endpoint string, result interface{}) error {
	resp, err := c.request("GET", endpoint, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) GetWithQuery(endpoint string, params map[string]string, result interface{}) error {
	u, err := url.Parse(c.buildURL(endpoint))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if token := config.GetAccessToken(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var rawMessage string
		if respBody, err := io.ReadAll(resp.Body); err == nil {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				if apiErr.Message != "" {
					rawMessage = apiErr.Message
				} else {
					rawMessage = apiErr.ErrorMsg
				}
			} else {
				rawMessage = strings.TrimSpace(string(respBody))
			}
		}
		return toActionableError(resp.StatusCode, endpoint, rawMessage)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Post(endpoint string, body interface{}, result interface{}) error {
	resp, err := c.request("POST", endpoint, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Delete(endpoint string) error {
	resp, err := c.request("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) Put(endpoint string, body interface{}, result interface{}) error {
	resp, err := c.request("PUT", endpoint, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Patch(endpoint string, body interface{}, result interface{}) error {
	resp, err := c.request("PATCH", endpoint, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) PostMultipart(endpoint string, fieldName string, fileName string, fileData io.Reader, fields map[string]string, result interface{}) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	reqURL := c.buildURL(endpoint)
	req, err := http.NewRequest("POST", reqURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if token := config.GetAccessToken(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var rawMessage string
		if respBody, err := io.ReadAll(resp.Body); err == nil {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				if apiErr.Message != "" {
					rawMessage = apiErr.Message
				} else {
					rawMessage = apiErr.ErrorMsg
				}
			} else {
				rawMessage = strings.TrimSpace(string(respBody))
			}
		}
		return toActionableError(resp.StatusCode, endpoint, rawMessage)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) DeleteWithBody(endpoint string, body interface{}) error {
	resp, err := c.request("DELETE", endpoint, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) HealthCheck() error {
	var result map[string]interface{}
	return c.Get("/health", &result)
}