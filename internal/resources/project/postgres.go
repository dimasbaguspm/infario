package project

import (
	"context"
	"fmt"

	"github.com/dimasbaguspm/infario/pkgs/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db}
}

func (r *PostgresRepository) GetPaged(ctx context.Context, params GetPagedProject) (*ProjectPaged, error) {
	offset := params.Offset()

	query := `
		WITH projects_cte AS (
			SELECT
				id,
				name,
				created_at,
				updated_at,
				deleted_at,
				COUNT(*) OVER () AS total_count
			FROM projects
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		)
		SELECT
			id,
			name,
			created_at,
			updated_at,
			deleted_at,
			total_count
		FROM projects_cte
	`

	rows, err := r.db.Query(ctx, query, params.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	var totalCount int64

	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.DeletedAt,
			&totalCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project rows: %w", err)
	}

	// Calculate total pages
	pageCount := (totalCount + int64(params.PageSize) - 1) / int64(params.PageSize)

	return (*ProjectPaged)(&response.Collection[*Project]{
		Items:      projects,
		TotalCount: totalCount,
		PageSize:   params.PageSize,
		PageNumber: params.PageNumber,
		PageCount:  pageCount,
	}), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, p GetSingleProject) (*Project, error) {
	query := `
		SELECT
			id,
			name,
			created_at,
			updated_at,
			deleted_at
		FROM projects
		WHERE id = $1
			AND deleted_at IS NULL
	`

	d := &Project{}
	err := r.db.QueryRow(ctx, query, p.ID).Scan(
		&d.ID,
		&d.Name,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return d, nil
}

func (r *PostgresRepository) Create(ctx context.Context, p CreateProject) (string, error) {
	var ID *string

	query := `
		INSERT INTO projects (name)
		VALUES ($1)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, p.Name).Scan(&ID)
	if err != nil {
		return "", fmt.Errorf("failed to create project: %w", err)
	}

	return *ID, nil
}

func (r *PostgresRepository) Update(ctx context.Context, p UpdateProject) error {
	query := `
		UPDATE projects
		SET name = COALESCE(NULLIF($1, ''), name),
			updated_at = NOW()
		WHERE id = $2
			AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, p.Name, p.ID)
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
