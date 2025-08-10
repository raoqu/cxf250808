package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const DEFAULT_HSET_GROUP = "default"

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

func CreateRedisConn() error {
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
		return err
	} else {
		log.Println("Successfully connected to Redis")
	}
	return nil
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
