package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type UploadResponse struct {
	Filename string `json:"filename,omitempty"`
	Path     string `json:"path,omitempty"`
	Error    string `json:"error,omitempty"`
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
