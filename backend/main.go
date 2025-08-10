package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const DEFAULT_HSET_GROUP = "default"

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

type SetRequest map[string]interface{}

type GetResponse struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type KeysResponse struct {
	Keys  []string `json:"keys,omitempty"`
	Error string   `json:"error,omitempty"`
}

// Redis client
var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func main() {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       11, // use default DB
	})

	// Test Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Printf("Redis operations will fail. Please ensure Redis is running on localhost:6379")
	} else {
		log.Println("Successfully connected to Redis")
	}

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

	// Return the URL content
	c.JSON(http.StatusOK, ParseResponse{
		URL:     targetURL,
		Content: string(body),
	})
}

// setRedisData handles POST /api/set requests
// Stores the entire request body as a string value with the key from URL parameter
func setRedisData(c *gin.Context) {
	// Get key from URL parameter
	key := c.Query("key")
	group := c.Query("group")
	if group == "" {
		group = DEFAULT_HSET_GROUP
	}
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	// Read the raw request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body: " + err.Error()})
		return
	}

	// Store the raw body as a string in Redis
	value := string(bodyBytes)
	// HSET group key value
	err = redisClient.HSet(ctx, group, key, value).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data in Redis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// getRedisData handles GET /api/get requests
// Uses Redis GET command to retrieve the stored value
func getRedisData(c *gin.Context) {
	// Get key from query parameter
	key := c.Query("key")
	group := c.Query("group")
	if group == "" {
		group = DEFAULT_HSET_GROUP
	}
	if key == "" {
		c.JSON(http.StatusBadRequest, GetResponse{Error: "key parameter is required"})
		return
	}

	// Get data from Redis
	data, err := redisClient.HGet(ctx, group, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, GetResponse{Error: "key not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, GetResponse{Error: "Failed to get data from Redis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetResponse{Data: data})
}

// getRedisKeys handles GET /api/hkeys requests
// Uses Redis KEYS command for hash keys
func getRedisKeys(c *gin.Context) {
	// Get all keys from Redis
	// Using KEYS * pattern - note this can be slow on large datasets
	// In production, consider using SCAN instead
	group := c.Query("group")
	if group == "" {
		group = DEFAULT_HSET_GROUP
	}
	keys, err := redisClient.HKeys(ctx, group).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, KeysResponse{Error: "Failed to get keys from Redis: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, KeysResponse{Keys: keys})
}
