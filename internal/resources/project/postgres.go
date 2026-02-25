package project

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db}
}

func (r *PostgresRepository) Create(ctx context.Context, p CreateProject) (string, error) {
	var ID *string

	query := `
		INSERT INTO projects (name, git_url, provider, primary_branch)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		p.Name,
		p.GitURL,
		p.GitProvider,
		p.PrimaryBranch,
	).Scan(&ID)

	if err != nil {
		return "", fmt.Errorf("failed to create project: %w", err)
	}

	return *ID, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, p GetSingleProject) (*Project, error) {
	query := `
		SELECT
			id,
			name,
			git_url,
			git_provider,
			primary_branch,
			created_at,
			updated_at
		FROM projects
		WHERE id = $1
			AND deleted_at IS NULL
	`

	d := &Project{}
	err := r.db.QueryRow(ctx, query, p.ID).Scan(
		&d.ID,
		&d.Name,
		&d.GitURL,
		&d.GitProvider,
		&d.PrimaryBranch,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return d, nil
}

func (r *PostgresRepository) Update(ctx context.Context, p UpdateProject) error {
	query := `
		UPDATE projects
		SET name = COALESCE($1, name),
			git_url = COALESCE($2, git_url),
			git_provider = COALESCE($3, git_provider),
			primary_branch = COALESCE($4, primary_branch),
			updated_at = NOW()
		WHERE id = $5
			AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		p.Name,
		p.GitURL,
		p.GitProvider,
		p.PrimaryBranch,
		p.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, p DeleteProject) error {
	query := `
		UPDATE projects
			SET deleted_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, p.ID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}
