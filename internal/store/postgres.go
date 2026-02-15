package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/web3-frozen/demo-api/internal/model"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id          TEXT PRIMARY KEY,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT 'todo',
		priority    TEXT NOT NULL DEFAULT 'medium',
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
	`
	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *PostgresStore) List(ctx context.Context) ([]model.Task, error) {
	rows, err := s.pool.Query(ctx,
		"SELECT id, title, description, status, priority, created_at, updated_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	return tasks, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*model.Task, error) {
	var t model.Task
	err := s.pool.QueryRow(ctx,
		"SELECT id, title, description, status, priority, created_at, updated_at FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStore) Create(ctx context.Context, req model.CreateTaskRequest) (*model.Task, error) {
	t := model.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo",
		Priority:    req.Priority,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_, err := s.pool.Exec(ctx,
		"INSERT INTO tasks (id, title, description, status, priority, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		t.ID, t.Title, t.Description, t.Status, t.Priority, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStore) Update(ctx context.Context, id string, req model.UpdateTaskRequest) (*model.Task, error) {
	existing, err := s.Get(ctx, id)
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
	existing.UpdatedAt = time.Now().UTC()

	_, err = s.pool.Exec(ctx,
		"UPDATE tasks SET title=$1, description=$2, status=$3, priority=$4, updated_at=$5 WHERE id=$6",
		existing.Title, existing.Description, existing.Status, existing.Priority, existing.UpdatedAt, id)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
