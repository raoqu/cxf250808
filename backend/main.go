package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	CreateRedisConn()

	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/parse", parseURL)
		api.POST("/upload", uploadFile)
		api.POST("/set", setRedisData)
		api.GET("/get", getRedisData)
		api.GET("/hkeys", getRedisKeys)
	}

	// Start server
	log.Println("Starting server on :6161...")
	if err := r.Run(":6161"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
