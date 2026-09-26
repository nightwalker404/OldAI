package vllmclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClientRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ClientResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Client struct {
	Model   string
	BaseURL string
}

func NewClient(model, baseURL string) *Client {
	return &Client{
		Model:   model,
		BaseURL: baseURL + "/v1/chat/completions",
	}
}

func (c *Client) CreateRequest(messages []Message, req ClientRequest) (ClientResponse, error) {
	body := ClientRequest{
		Model:    c.Model,
		Messages: messages,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(c.BaseURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return ClientResponse{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClientResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var clientResp ClientResponse
	err = json.Unmarshal(responseBody, &clientResp)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return clientResp, nil
}
