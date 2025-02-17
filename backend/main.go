package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"gitlab.com/gomidi/midi"
)

var (
	dataStore = make(map[string]string)
	mutex     = &sync.Mutex{}
	authToken = "securetoken"
)

// simple placeholder for auth
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer "+authToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// these are simple in-memory data store for demonstration purposes
func getData(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	c.JSON(http.StatusOK, dataStore)
}

func createData(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	mutex.Lock()
	for k, v := range req {
		dataStore[k] = v
	}
	mutex.Unlock()
	c.JSON(http.StatusCreated, gin.H{"message": "Created successfully"})
}

func updateData(c *gin.Context) {
	key := c.Param("key")
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := dataStore[key]; exists {
		dataStore[key] = req["value"]
		c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
	}
}

func deleteData(c *gin.Context) {
	key := c.Param("key")
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := dataStore[key]; exists {
		delete(dataStore, key)
		c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
	}
}

func parseMIDI(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := midi.Reader(file)

	for {
		event, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			return err
		}
		if noteOnEvent, ok := event.(*midi.NoteOnEvent); ok {
			fmt.Printf("Note: %v, Velocity: %v\n", noteOnEvent.Note, noteOnEvent.Velocity)
		}
	}

	return nil
}

// upload file function
func uploadFile(c *gin.Context) {
	file, err := c.FormFile("midi")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
		return
	}

	if file.Header.Get("Content-Type") != "audio/midi" && file.Header.Get("Content-Type") != "audio/x-midi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type, only MIDI files allowed"})
		return
	}

	filePath := "./uploads/" + file.Filename
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "path": filePath})
}

func main() {
	r := gin.Default()

	auth := r.Group("/data", authMiddleware())
	{
		auth.GET("", getData)
		auth.POST("", createData)
		auth.PUT("/:key", updateData)
		auth.DELETE("/:key", deleteData)
	}

	r.Run(":8080")
}
