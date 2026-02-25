package deployment

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

func (r *PostgresRepository) GetByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error) {
	query := `
		SELECT
			id,
			project_id,
			status,
			commit_hash,
			commit_message,
			storage_path,
			public_url,
			created_at,
			updated_at
		FROM deployments
		WHERE id = $1
	`

	deployment := &Deployment{}
	err := r.db.QueryRow(ctx, query, d.ID).Scan(
		&deployment.ID,
		&deployment.ProjectID,
		&deployment.Status,
		&deployment.CommitHash,
		&deployment.CommitMessage,
		&deployment.StoragePath,
		&deployment.PublicURL,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *PostgresRepository) GetPaged(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error) {
	offset := (params.PageNumber - 1) * params.PageSize

	query := `
		WITH deployments_cte AS (
			SELECT
				id,
				project_id,
				status,
				commit_hash,
				commit_message,
				storage_path,
				public_url,
				created_at,
				updated_at,
				COUNT(*) OVER () AS total_count
			FROM deployments
			WHERE project_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		)
		SELECT
			id,
			project_id,
			status,
			commit_hash,
			commit_message,
			storage_path,
			public_url,
			created_at,
			updated_at,
			total_count
		FROM deployments_cte
	`

	rows, err := r.db.Query(ctx, query, params.ProjectID, params.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*Deployment
	var totalCount int64

	for rows.Next() {
		deployment := &Deployment{}
		err := rows.Scan(
			&deployment.ID,
			&deployment.ProjectID,
			&deployment.Status,
			&deployment.CommitHash,
			&deployment.CommitMessage,
			&deployment.StoragePath,
			&deployment.PublicURL,
			&deployment.CreatedAt,
			&deployment.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}
		deployments = append(deployments, deployment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment rows: %w", err)
	}

	// Calculate total pages
	pageCount := (totalCount + int64(params.PageSize) - 1) / int64(params.PageSize)

	return &DeploymentPaged{
		Items:      deployments,
		TotalCount: totalCount,
		PageSize:   params.PageSize,
		PageNumber: params.PageNumber,
		PageCount:  pageCount,
	}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, d CreateDeployment) (string, error) {
	var ID *string

	query := `
		INSERT INTO deployments (project_id, status, commit_hash, commit_message, storage_path)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		d.ProjectID,
		StatusQueued,
		d.CommitHash,
		d.CommitMessage,
		d.StoragePath,
	).Scan(&ID)

	if err != nil {
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}

	return *ID, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, d UpdateDeploymentStatus) error {
	query := `
		UPDATE deployments
		SET status = $1,
			public_url = COALESCE($2, public_url),
			updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query,
		d.Status,
		d.PublicURL,
		d.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, d DeleteDeployment) error {
	query := `
		DELETE FROM deployments
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, d.ID)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}
