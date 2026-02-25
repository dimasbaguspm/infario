package domain

import "time"

type Deployment struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	Hash       string     `json:"hash"`
	Status     string     `json:"status"`
	PreviewURL string     `json:"preview_url,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type GetSingleDeployment struct {
	ID string `json:"id" validate:"required"`
}

type CreateDeployment struct {
	ProjectID string `json:"project_id" validate:"required"`
	Hash      string `json:"hash" validate:"required"`
}

type UpdateDeployment struct {
	ID         string `json:"id" validate:"required"`
	Status     string `json:"status,omitempty"`
	PreviewURL string `json:"preview_url,omitempty"`
}

type DeleteDeployment struct {
	ID string `json:"id" validate:"required"`
}
