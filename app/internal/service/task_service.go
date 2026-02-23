package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/hamfa/task-manager/internal/model"
	"github.com/hamfa/task-manager/internal/repository"
)

// TaskService handles business logic for tasks
type TaskService struct {
	postgresRepo *repository.PostgresRepository
	mongoRepo    *repository.MongoRepository
	redisCache   *repository.RedisCache
	logger       *zap.Logger
}

// NewTaskService creates a new task service
func NewTaskService(
	pg *repository.PostgresRepository,
	mongo *repository.MongoRepository,
	redis *repository.RedisCache,
	logger *zap.Logger,
) *TaskService {
	return &TaskService{
		postgresRepo: pg,
		mongoRepo:    mongo,
		redisCache:   redis,
		logger:       logger,
	}
}

// Create creates a new task and logs the activity
func (s *TaskService) Create(ctx context.Context, req model.TaskCreateRequest) (*model.Task, error) {
	task, err := s.postgresRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("service: create task: %w", err)
	}

	// Cache the new task
	if cacheErr := s.redisCache.SetTask(ctx, task); cacheErr != nil {
		s.logger.Warn("failed to cache new task", zap.Error(cacheErr))
	}

	// Log activity to MongoDB (non-blocking)
	go func() {
		if logErr := s.mongoRepo.LogActivity(context.Background(), task.ID, "created",
			fmt.Sprintf("Task '%s' created with priority %s", task.Title, task.Priority)); logErr != nil {
			s.logger.Warn("failed to log activity", zap.Error(logErr))
		}
	}()

	return task, nil
}

// GetByID retrieves a task by ID with cache-aside pattern
func (s *TaskService) GetByID(ctx context.Context, id string) (*model.Task, error) {
	// Try cache first
	cached, err := s.redisCache.GetTask(ctx, id)
	if err != nil {
		s.logger.Warn("cache lookup failed", zap.Error(err))
	}
	if cached != nil {
		s.logger.Debug("cache hit", zap.String("task_id", id))
		return cached, nil
	}

	// Cache miss — fetch from PostgreSQL
	task, err := s.postgresRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: get task: %w", err)
	}

	// Populate cache
	if cacheErr := s.redisCache.SetTask(ctx, task); cacheErr != nil {
		s.logger.Warn("failed to cache task", zap.Error(cacheErr))
	}

	return task, nil
}

// List retrieves paginated tasks
func (s *TaskService) List(ctx context.Context, page, perPage int, status string) (*model.TaskListResponse, error) {
	tasks, total, err := s.postgresRepo.List(ctx, page, perPage, status)
	if err != nil {
		return nil, fmt.Errorf("service: list tasks: %w", err)
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	return &model.TaskListResponse{
		Data:    tasks,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// Update modifies a task and invalidates its cache
func (s *TaskService) Update(ctx context.Context, id string, req model.TaskUpdateRequest) (*model.Task, error) {
	task, err := s.postgresRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("service: update task: %w", err)
	}

	// Invalidate and recache
	if cacheErr := s.redisCache.InvalidateTask(ctx, id); cacheErr != nil {
		s.logger.Warn("failed to invalidate cache", zap.Error(cacheErr))
	}
	if cacheErr := s.redisCache.SetTask(ctx, task); cacheErr != nil {
		s.logger.Warn("failed to recache task", zap.Error(cacheErr))
	}

	// Log activity
	go func() {
		if logErr := s.mongoRepo.LogActivity(context.Background(), id, "updated",
			fmt.Sprintf("Task '%s' updated", task.Title)); logErr != nil {
			s.logger.Warn("failed to log activity", zap.Error(logErr))
		}
	}()

	return task, nil
}

// Delete removes a task
func (s *TaskService) Delete(ctx context.Context, id string) error {
	// Get task info before delete for logging
	task, _ := s.postgresRepo.GetByID(ctx, id)

	if err := s.postgresRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service: delete task: %w", err)
	}

	// Invalidate cache
	if cacheErr := s.redisCache.InvalidateTask(ctx, id); cacheErr != nil {
		s.logger.Warn("failed to invalidate cache", zap.Error(cacheErr))
	}

	// Log activity
	go func() {
		title := id
		if task != nil {
			title = task.Title
		}
		if logErr := s.mongoRepo.LogActivity(context.Background(), id, "deleted",
			fmt.Sprintf("Task '%s' deleted", title)); logErr != nil {
			s.logger.Warn("failed to log activity", zap.Error(logErr))
		}
	}()

	return nil
}

// GetActivities returns activity logs for a task
func (s *TaskService) GetActivities(ctx context.Context, taskID string, limit int64) ([]model.ActivityLog, error) {
	return s.mongoRepo.GetActivities(ctx, taskID, limit)
}
