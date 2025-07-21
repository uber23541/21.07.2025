package handlers

import (
	"archive_service/internal/taskmanager"

	"github.com/gin-gonic/gin"
)

func SetupRouter(manager *taskmanager.Manager) *gin.Engine {
	router := gin.Default()

	router.Static("/archives", manager.StorageDir())

	handler := &taskHandler{manager: manager}

	router.POST("/tasks/create", handler.createTask)
	router.POST("/tasks/add", handler.addFile)
	router.GET("/tasks/status", handler.getStatus)

	return router
}
