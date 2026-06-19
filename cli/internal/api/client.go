package api

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
)

const maxResponseBodyBytes = 1 << 20

var ErrConflict = errors.New("fluxcore api conflict")

type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

type Project struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	ID            uint      `json:"id"`
	ProjectID     uint      `json:"project_id"`
	Name          string    `json:"name"`
	LocalPath     string    `json:"local_path"`
	RemoteURL     string    `json:"remote_url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateProjectInput struct {
	Name string
}

type CreateRepositoryInput struct {
	Name          string
	LocalPath     string
	RemoteURL     string
	DefaultBranch string
}

type Error struct {
	StatusCode int
	Code       string
	Message    string
}

func (err Error) Error() string {
	if err.Code != "" && err.Message != "" {
		return fmt.Sprintf("api error %d %s: %s", err.StatusCode, err.Code, err.Message)
	}
	return fmt.Sprintf("api error %d", err.StatusCode)
}

func NewClient(baseURL string, token string) (*Client, error) {
	return NewClientWithHTTPClient(baseURL, token, http.DefaultClient)
}

func NewClientWithHTTPClient(baseURL string, token string, httpClient *http.Client) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	token = strings.TrimSpace(token)
	if baseURL == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse server URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("server URL must use http or https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("server URL host is required")
	}
	if token == "" {
		return nil, fmt.Errorf("API token is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    parsed,
		token:      token,
		httpClient: httpClient,
	}, nil
}

func (client *Client) CreateOrGetProject(ctx context.Context, input CreateProjectInput) (Project, error) {
	project, err := client.CreateProject(ctx, input)
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, ErrConflict) {
		return Project{}, err
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return Project{}, err
	}
	for _, item := range projects {
		if item.Name == input.Name {
			return item, nil
		}
	}
	return Project{}, fmt.Errorf("project %q already exists but could not be found", input.Name)
}

func (client *Client) CreateProject(ctx context.Context, input CreateProjectInput) (Project, error) {
	var response struct {
		Project Project `json:"project"`
	}
	err := client.doJSON(ctx, http.MethodPost, "/api/projects", map[string]string{
		"name": input.Name,
	}, &response)
	return response.Project, err
}

func (client *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var response struct {
		Projects []Project `json:"projects"`
	}
	if err := client.doJSON(ctx, http.MethodGet, "/api/projects", nil, &response); err != nil {
		return nil, err
	}
	return response.Projects, nil
}

func (client *Client) CreateOrGetRepository(ctx context.Context, projectID uint, input CreateRepositoryInput) (Repository, error) {
	repository, err := client.CreateRepository(ctx, projectID, input)
	if err == nil {
		return repository, nil
	}
	if !errors.Is(err, ErrConflict) {
		return Repository{}, err
	}

	repositories, err := client.ListRepositories(ctx, projectID)
	if err != nil {
		return Repository{}, err
	}
	for _, item := range repositories {
		if item.LocalPath == input.LocalPath {
			return item, nil
		}
	}
	return Repository{}, fmt.Errorf("repository %q already exists but could not be found", input.Name)
}

func (client *Client) CreateRepository(ctx context.Context, projectID uint, input CreateRepositoryInput) (Repository, error) {
	var response struct {
		Repository Repository `json:"repository"`
	}
	err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/projects/%d/repositories", projectID), map[string]string{
		"name":           input.Name,
		"local_path":     input.LocalPath,
		"remote_url":     input.RemoteURL,
		"default_branch": input.DefaultBranch,
	}, &response)
	return response.Repository, err
}

func (client *Client) ListRepositories(ctx context.Context, projectID uint) ([]Repository, error) {
	var response struct {
		Repositories []Repository `json:"repositories"`
	}
	if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/projects/%d/repositories", projectID), nil, &response); err != nil {
		return nil, err
	}
	return response.Repositories, nil
}

func (client *Client) doJSON(ctx context.Context, method string, path string, requestBody interface{}, responseBody interface{}) error {
	var body io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		body = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, method, client.endpoint(path), body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+client.token)
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	limitedBody := io.LimitReader(response.Body, maxResponseBodyBytes+1)
	data, err := io.ReadAll(limitedBody)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if len(data) > maxResponseBodyBytes {
		return fmt.Errorf("response body exceeds maximum size")
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return parseError(response.StatusCode, data)
	}

	if responseBody == nil {
		return nil
	}
	if err := json.Unmarshal(data, responseBody); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func (client *Client) endpoint(path string) string {
	clone := *client.baseURL
	clone.Path = strings.TrimRight(clone.Path, "/") + path
	clone.RawQuery = ""
	clone.Fragment = ""
	return clone.String()
}

func parseError(statusCode int, data []byte) error {
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		apiErr := Error{
			StatusCode: statusCode,
			Message:    http.StatusText(statusCode),
		}
		if statusCode == http.StatusConflict {
			return fmt.Errorf("%w: %w", ErrConflict, apiErr)
		}
		return apiErr
	}

	apiErr := Error{
		StatusCode: statusCode,
		Code:       envelope.Error.Code,
		Message:    envelope.Error.Message,
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("%w: %w", ErrConflict, apiErr)
	}
	return apiErr
}
