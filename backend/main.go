package main

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type ParseResponse struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

func main() {
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
	}

	// Start server
	log.Println("Starting server on :8081...")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func parseURL(c *gin.Context) {
	// Get URL from query parameter
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Error: "URL parameter is required",
		})
		return
	}

	// Validate URL
	_, err := url.ParseRequestURI(targetURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			URL:   targetURL,
			Error: "Invalid URL format",
		})
		return
	}

	// Fetch the URL content
	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			URL:   targetURL,
			Error: "Failed to fetch URL: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, ParseResponse{
			URL:   targetURL,
			Error: "Received non-200 response: " + resp.Status,
		})
		return
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			URL:   targetURL,
			Error: "Failed to read response body: " + err.Error(),
		})
		return
	}

	// Return the content
	c.JSON(http.StatusOK, ParseResponse{
		URL:     targetURL,
		Content: string(body),
	})
}
