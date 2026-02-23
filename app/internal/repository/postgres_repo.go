package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hamfa/task-manager/internal/model"
)

// PostgresRepository handles PostgreSQL operations for tasks
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// InitSchema creates the tasks table if it doesn't exist
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(36) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			status VARCHAR(20) DEFAULT 'pending',
			priority VARCHAR(20) DEFAULT 'medium',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
		CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// Create inserts a new task
func (r *PostgresRepository) Create(ctx context.Context, req model.TaskCreateRequest) (*model.Task, error) {
	task := &model.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if task.Priority == "" {
		task.Priority = "medium"
	}

	query := `
		INSERT INTO tasks (id, title, description, status, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, priority, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		task.ID, task.Title, task.Description, task.Status,
		task.Priority, task.CreatedAt, task.UpdatedAt,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetByID retrieves a task by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	query := `SELECT id, title, description, status, priority, created_at, updated_at FROM tasks WHERE id = $1`

	var task model.Task
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	return &task, nil
}

// List retrieves paginated tasks
func (r *PostgresRepository) List(ctx context.Context, page, perPage int, status string) ([]model.Task, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Count total
	countQuery := `SELECT COUNT(*) FROM tasks`
	args := []interface{}{}

	if status != "" {
		countQuery += ` WHERE status = $1`
		args = append(args, status)
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Fetch tasks
	listQuery := `SELECT id, title, description, status, priority, created_at, updated_at FROM tasks`
	if status != "" {
		listQuery += ` WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, perPage, offset)
	} else {
		listQuery += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, perPage, offset)
	}

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, total, nil
}

// Update modifies an existing task
func (r *PostgresRepository) Update(ctx context.Context, id string, req model.TaskUpdateRequest) (*model.Task, error) {
	// First get the existing task
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}
	existing.UpdatedAt = time.Now()

	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, priority, created_at, updated_at
	`

	var task model.Task
	err = r.pool.QueryRow(ctx, query,
		existing.Title, existing.Description, existing.Status,
		existing.Priority, existing.UpdatedAt, id,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return &task, nil
}

// Delete removes a task by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Ping checks the database connection
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
