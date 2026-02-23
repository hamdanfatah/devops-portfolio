package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hamfa/task-manager/internal/model"
)

// MongoRepository handles MongoDB operations for activity logging
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("activity_logs"),
	}
}

// LogActivity records an activity log entry
func (r *MongoRepository) LogActivity(ctx context.Context, taskID, action, details string) error {
	log := model.ActivityLog{
		TaskID:    taskID,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// GetActivities retrieves activity logs for a specific task
func (r *MongoRepository) GetActivities(ctx context.Context, taskID string, limit int64) ([]model.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	filter := bson.D{{Key: "task_id", Value: taskID}}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []model.ActivityLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	return logs, nil
}

// GetRecentActivities retrieves the most recent activity logs across all tasks
func (r *MongoRepository) GetRecentActivities(ctx context.Context, limit int64) ([]model.ActivityLog, error) {
	if limit <= 0 {
		limit = 20
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []model.ActivityLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	return logs, nil
}

// Ping checks the MongoDB connection
func (r *MongoRepository) Ping(ctx context.Context) error {
	return r.collection.Database().Client().Ping(ctx, nil)
}
