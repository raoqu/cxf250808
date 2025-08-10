package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type ParseResponse struct {
	URL     string `json:"url"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
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
