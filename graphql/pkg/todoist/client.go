package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseURL = "https://api.todoist.com/rest"

type (
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	Task struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
		Content   string `json:"content"`
		Completed bool   `json:"completed"`
	}

	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	Client struct {
		token      string
		httpClient HTTPClient
	}
)

func NewClient(token string, httpClient HTTPClient) *Client {
	return &Client{
		token:      token,
		httpClient: httpClient,
	}
}

func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	fmt.Println("GetProjects")
	req, err := http.NewRequest("GET", baseURL+"/v2/projects", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var projects []Project
	if err = json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return projects, nil
}

func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	fmt.Println("GetProject", projectID)
	req, err := http.NewRequest("GET", baseURL+"/v2/projects/"+projectID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var project Project
	if err = json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &project, nil
}

func (c *Client) GetTasks(ctx context.Context, projectID string) ([]Task, error) {
	fmt.Println("GetTasks", projectID)
	req, err := http.NewRequest("GET", baseURL+"/v2/tasks", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	q := req.URL.Query()
	if projectID != "" {
		q.Add("project_id", projectID)
	}
	req.URL.RawQuery = q.Encode()

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var tasks []Task
	if err = json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return tasks, nil
}

func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	fmt.Println("GetTask", taskID)
	req, err := http.NewRequest("GET", baseURL+"/v2/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var task Task
	if err = json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &task, nil
}
