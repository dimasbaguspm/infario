package domain

import "time"

type Project struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletdAt  *time.Time `json:"deleted_at,omitempty"`
}

type GetSingleProject struct {
	ID string `json:"id" validate:"required"`
}

type CreateProject struct {
	Name string `json:"name" validate:"required"`
}

type UpdateProject struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name,omitempty"`
}

type DeleteProject struct {
	ID string `json:"id" validate:"required"`
}
