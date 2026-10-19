package infisical

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "http://localhost:80"

type KMSClient struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func NewClient(token, keyID string) *KMSClient {
	return &KMSClient{
		BaseURL: defaultBaseURL,
		Token:   token,
		HTTP:    http.DefaultClient,
	}
}

type encryptRequest struct {
	Plaintext string `json:"plaintext"`
}

type decryptRequest struct {
	Ciphertext string `json:"ciphertext"`
}

type kmsResponse struct {
	Plaintext  string `json:"plaintext,omitempty"`
	Ciphertext string `json:"ciphertext,omitempty"`
}

func (c *KMSClient) call(path string, body any, keyID string) (*kmsResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/kms/keys/%s/%s", c.BaseURL, keyID, path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("infisical %s: status %d, body: %s", path, resp.StatusCode, string(data))
	}

	var out kmsResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &out, nil
}

func (c *KMSClient) Encrypt(plaintext string, keyID string) (string, error) {
	resp, err := c.call("encrypt", encryptRequest{Plaintext: plaintext}, keyID)
	if err != nil {
		return "", err
	}
	return resp.Ciphertext, nil
}

func (c *KMSClient) Decrypt(ciphertext string, keyID string) (string, error) {
	resp, err := c.call("decrypt", decryptRequest{Ciphertext: ciphertext}, keyID)
	if err != nil {
		return "", err
	}
	return resp.Plaintext, nil
}
