package domain

import "time"

type Deployment struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	CommitHash   string     `json:"commit_hash"`
	Status       string     `json:"status"`
	PreviewURL   string     `json:"preview_url,omitempty"`
	StorageJSON  string     `json:"storage_json,omitempty"`
	MetadataJSON string     `json:"metadata_json,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type GetSingleDeployment struct {
	ID string `json:"id" validate:"required"`
}

type CreateDeployment struct {
	ProjectID string `json:"project_id" validate:"required"`
	Hash      string `json:"hash" validate:"required"`
}
