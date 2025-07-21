// swag init -g server/main.go -d . --output docs

package main

import (
	"log"

	"archive_service/internal/config"
	"archive_service/internal/handlers"
	"archive_service/internal/taskmanager"

	docs "archive_service/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title       Archive Service API
// @version     1.0
// @description Create ZIP archives from public URLs.
// @BasePath    /
func main() {
	docs.SwaggerInfo.BasePath = "/"

	cfg := config.LoadConfig()
	manager := taskmanager.New(&cfg)
	router := handlers.SetupRouter(manager)

	// /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("service started on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("startup error: %v", err)
	}
}
