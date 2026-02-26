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
			hash,
			status,
			created_at,
			expired_at,
			entry_path
		FROM deployments
		WHERE id = $1
	`

	deployment := &Deployment{}
	err := r.db.QueryRow(ctx, query, d.ID).Scan(
		&deployment.ID,
		&deployment.ProjectID,
		&deployment.Hash,
		&deployment.Status,
		&deployment.CreatedAt,
		&deployment.ExpiredAt,
		&deployment.EntryPath,
	)

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *PostgresRepository) GetPaged(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error) {
	offset := params.Offset()

	// Always join with projects table to include project_name
	query := `
		WITH deployments_cte AS (
			SELECT
				d.id,
				d.project_id,
				d.hash,
				d.status,
				d.created_at,
				d.expired_at,
				d.entry_path,
				p.name AS project_name,
				COUNT(*) OVER () AS total_count
			FROM deployments d
			LEFT JOIN projects p ON p.id = d.project_id
			WHERE ($1::uuid IS NULL OR d.project_id = $1::uuid)
			AND ($2::text IS NULL OR d.status = $2)
			ORDER BY d.created_at DESC
			LIMIT $3 OFFSET $4
		)
		SELECT
			id,
			project_id,
			hash,
			status,
			created_at,
			expired_at,
			entry_path,
			project_name,
			total_count
		FROM deployments_cte
	`

	rows, err := r.db.Query(ctx, query, params.ProjectID, params.Status, params.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer rows.Close()

	deployments := make([]Deployment, 0)
	var totalCount int64

	for rows.Next() {
		deployment := &Deployment{}
		var projectName *string
		err := rows.Scan(
			&deployment.ID,
			&deployment.ProjectID,
			&deployment.Hash,
			&deployment.Status,
			&deployment.CreatedAt,
			&deployment.ExpiredAt,
			&deployment.EntryPath,
			&projectName,
			&totalCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}
		deployment.ProjectName = projectName
		deployments = append(deployments, *deployment)
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

func (r *PostgresRepository) Upload(ctx context.Context, d UploadDeployment) (string, error) {
	var ID *string

	// Default TTL: 30 days from now
	expiredAt := "NOW() + INTERVAL '30 days'"

	query := fmt.Sprintf(`
		INSERT INTO deployments (project_id, hash, status, entry_path, expired_at)
		VALUES ($1, $2, $3, $4, %s)
		RETURNING id
	`, expiredAt)

	err := r.db.QueryRow(ctx, query,
		d.ProjectID,
		d.Hash,
		StatusPending,
		d.EntryPath,
	).Scan(&ID)

	if err != nil {
		return "", fmt.Errorf("failed to upload deployment: %w", err)
	}

	return *ID, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, d UpdateDeploymentStatus) error {
	query := `
		UPDATE deployments
		SET status = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, d.Status, d.ID)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// GetExpired retrieves all deployments that have exceeded their TTL.
func (r *PostgresRepository) GetExpired(ctx context.Context) ([]Deployment, error) {
	query := `
		SELECT
			id,
			project_id,
			hash,
			status,
			created_at,
			expired_at,
			entry_path
		FROM deployments
		WHERE expired_at IS NOT NULL
		AND expired_at <= NOW()
		AND status != $1
		ORDER BY expired_at ASC
	`

	rows, err := r.db.Query(ctx, query, StatusExpired)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired deployments: %w", err)
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		if err := rows.Scan(
			&deployment.ID,
			&deployment.ProjectID,
			&deployment.Hash,
			&deployment.Status,
			&deployment.CreatedAt,
			&deployment.ExpiredAt,
			&deployment.EntryPath,
		); err != nil {
			return nil, fmt.Errorf("failed to scan expired deployment: %w", err)
		}
		deployments = append(deployments, deployment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired deployment rows: %w", err)
	}

	return deployments, nil
}
