package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hamfa/task-manager/internal/model"
	"github.com/hamfa/task-manager/internal/service"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	service *service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{service: svc}
}

// RegisterRoutes registers all task routes
func (h *TaskHandler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)
		tasks.GET("/:id/activities", h.GetTaskActivities)
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body model.TaskCreateRequest true "Task to create"
// @Success 201 {object} model.TaskResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req model.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	task, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create task",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, model.TaskResponse{Data: *task})
}

// GetTask godoc
// @Summary Get a task by ID
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} model.TaskResponse
// @Failure 404 {object} model.ErrorResponse
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")

	task, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, model.TaskResponse{Data: *task})
}

// ListTasks godoc
// @Summary List all tasks
// @Tags tasks
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} model.TaskListResponse
// @Router /api/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.Query("status")

	result, err := h.service.List(c.Request.Context(), page, perPage, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list tasks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateTask godoc
// @Summary Update a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body model.TaskUpdateRequest true "Task updates"
// @Success 200 {object} model.TaskResponse
// @Failure 400,404 {object} model.ErrorResponse
// @Router /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var req model.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	task, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, model.TaskResponse{Data: *task})
}

// DeleteTask godoc
// @Summary Delete a task
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 204
// @Failure 404 {object} model.ErrorResponse
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTaskActivities returns activity logs for a task
func (h *TaskHandler) GetTaskActivities(c *gin.Context) {
	id := c.Param("id")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)

	activities, err := h.service.GetActivities(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get activities",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activities})
}
