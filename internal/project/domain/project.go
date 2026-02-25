package domain

import "time"

type Project struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	GitURL        string     `json:"git_url"`
	GitProvider   string     `json:"git_provider"` // github, gitlab, etc.
	PrimaryBranch string     `json:"primary_branch"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletdAt      *time.Time `json:"deleted_at,omitempty"`
}

type GetSingleProject struct {
	ID string `json:"id" validate:"required"`
}

type CreateProject struct {
	Name          string `json:"name" validate:"required"`
	GitURL        string `json:"git_url" validate:"required"`
	GitProvider   string `json:"git_provider,omitempty"` // github, gitlab, etc.
	PrimaryBranch string `json:"primary_branch,omitempty"`
}

type UpdateProject struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty"`
	GitURL        string `json:"git_url,omitempty"`
	GitProvider   string `json:"git_provider,omitempty"` // github, gitlab, etc.
	PrimaryBranch string `json:"primary_branch,omitempty"`
}

type DeleteProject struct {
	ID string `json:"id" validate:"required"`
}
