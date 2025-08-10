package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ParseResponse struct {
	URL     string `json:"url"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UploadResponse struct {
	Filename string `json:"filename,omitempty"`
	Path     string `json:"path,omitempty"`
	Error    string `json:"error,omitempty"`
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
		api.POST("/upload", uploadFile)
	}

	// Start server
	log.Println("Starting server on :8081...")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func uploadFile(c *gin.Context) {
	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Error: "No file uploaded: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Create the upload directory if it doesn't exist
	if err := os.MkdirAll("upload", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Error: "Failed to create upload directory: " + err.Error(),
		})
		return
	}

	// Create a file in the upload directory with the original filename
	filename := header.Filename
	filePath := filepath.Join("upload", filename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Error: "Failed to create file: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Error: "Failed to save file: " + err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, UploadResponse{
		Filename: filename,
		Path:     "/upload/" + filename,
	})
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
