package handlers

import (
	"net/http"

	"archive_service/internal/models"
	"archive_service/internal/taskmanager"

	"github.com/gin-gonic/gin"
)

type taskHandler struct {
	manager *taskmanager.Manager
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// createTask 	godoc
// @Summary     Create a new task
// @Tags        tasks
// @Produce     json
// @Success     200 {object} models.Task
// @Failure     503 {object} handlers.ErrorResponse
// @Router      /tasks/create [post]
func (h *taskHandler) createTask(ctx *gin.Context) {
	var task *models.Task
	var err error
	task, err = h.manager.CreateTask()
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, ErrorResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, task)
}

// addFile 		godoc
// @Summary     Add file URL to a task
// @Tags        tasks
// @Produce     json
// @Param       task_id query string true "Task ID"
// @Param       url     query string true "File URL"
// @Success     200 {object} models.Task
// @Failure     400 {object} handlers.ErrorResponse
// @Router      /tasks/add [post]
func (h *taskHandler) addFile(ctx *gin.Context) {
	taskID := ctx.Query("task_id")
	url := ctx.Query("url")
	if taskID == "" || url == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "task_id and url are required"})
		return
	}
	task, err := h.manager.AddFile(taskID, url)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, task)
}

// getStatus godoc
// @Summary     Get task status
// @Tags        tasks
// @Produce     json
// @Param       task_id query string true "Task ID"
// @Success     200 {object} models.Task
// @Failure     400 {object} handlers.ErrorResponse
// @Failure     404 {object} handlers.ErrorResponse
// @Router      /tasks/status [get]
func (h *taskHandler) getStatus(ctx *gin.Context) {
	taskID := ctx.Query("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "task_id is required"})
		return
	}
	task, ok := h.manager.Get(taskID)
	if !ok {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "task not found"})
		return
	}
	ctx.JSON(http.StatusOK, task)
}
