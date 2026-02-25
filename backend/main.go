package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/user/distributed-knights/backend/db"
	"github.com/user/distributed-knights/backend/handlers"
	"github.com/user/distributed-knights/backend/messenger"
)

func main() {
	// Initialize Database (The Chronicles)
	db.InitDB()

	// Initialize NATS (The Messenger's Guild)
	messenger.InitNATS()

	if os.Getenv("WORKER_MODE") == "true" {
		messenger.StartWorker()
		select {} // Block forever
	}

	// Also start a local worker for convenience in single-container mode
	go messenger.StartWorker()

	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.POST("/analyze", handlers.AnalyzeHandler)
	r.GET("/campaigns", handlers.ListCampaignsHandler)
	r.GET("/campaigns/:id", handlers.GetCampaignHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
