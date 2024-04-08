package graph

type (
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	Task struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
		Content   string `json:"content"`
	}
)
