package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin HTTP client for the local Ollama daemon management API.
// The base URL is resolved lazily so configuration changes (made through the
// AI providers settings) take effect without rebuilding the client.
type Client struct {
	baseURL func() string
	http    *http.Client
}

func NewClient(baseURL func() string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) root() string {
	return strings.TrimRight(c.baseURL(), "/")
}

type versionResponse struct {
	Version string `json:"version"`
}

// Version returns the daemon version, doubling as a reachability probe.
func (c *Client) Version(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.root()+"/api/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama: version returned status %d", resp.StatusCode)
	}

	var parsed versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return parsed.Version, nil
}

type tagsResponse struct {
	Models []struct {
		Name       string    `json:"name"`
		Model      string    `json:"model"`
		Size       int64     `json:"size"`
		Digest     string    `json:"digest"`
		ModifiedAt time.Time `json:"modified_at"`
		Details    struct {
			Family            string `json:"family"`
			ParameterSize     string `json:"parameter_size"`
			QuantizationLevel string `json:"quantization_level"`
		} `json:"details"`
	} `json:"models"`
}

// ListModels returns the models currently installed on the daemon.
func (c *Client) ListModels(ctx context.Context) ([]ModelDto, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.root()+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama: tags returned status %d", resp.StatusCode)
	}

	var parsed tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	models := make([]ModelDto, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		models = append(models, ModelDto{
			Name:              m.Name,
			Size:              m.Size,
			Digest:            m.Digest,
			ModifiedAt:        m.ModifiedAt,
			Family:            m.Details.Family,
			ParameterSize:     m.Details.ParameterSize,
			QuantizationLevel: m.Details.QuantizationLevel,
		})
	}
	return models, nil
}

// DeleteModel removes an installed model from the daemon.
func (c *Client) DeleteModel(ctx context.Context, name string) error {
	body, err := json.Marshal(map[string]string{"name": name, "model": name})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.root()+"/api/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrModelNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama: delete returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}
