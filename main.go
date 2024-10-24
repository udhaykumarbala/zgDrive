package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"zgdrive/model"
	"zgdrive/services"

	_ "zgdrive/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

// @title ZGDrive API
// @version 1.0
// @description This is the API for ZGDrive file management system.
// @host localhost:8080
// @BasePath /

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	ctx := context.Background()
	newFilesChan := make(chan model.File)
	downloadedFilesChan := make(chan model.File)
	zgService, err := services.NewZgService()
	if err != nil {
		fmt.Println("Error creating ZgService:", err)
		return
	}

	dbservice := services.NewDBService("./files.db")
	if dbservice == nil {
		log.Fatal("Failed to initialize database service")
	}

	go func() {
		for {
			newFile := <-newFilesChan
			fmt.Println("New file:", newFile)
			tx, err := zgService.UploadFile(ctx, newFile.Filename)
			if err != nil {
				fmt.Println("Error uploading file:", err)
				return
			}
			fmt.Println("Transaction hash:", tx)

			if tx != "" {
				dbservice.UpdateTxId(ctx, newFile.ID, tx)
			}
		}
	}()

	go func() {
		for {
			downloadedFile := <-downloadedFilesChan
			fmt.Println("Downloaded file:", downloadedFile)
			dbservice.NewDownloadedFile(ctx, downloadedFile)

			// download file from zgdrive
			isDone, err := zgService.DownloadFile(ctx, downloadedFile.Filename, downloadedFile.Hash)
			if err != nil {
				fmt.Println("Error downloading file:", err)
				return
			}
			if isDone {
				dbservice.SetProcessing(ctx, downloadedFile.ID)
				// move file to downloaded directory
				filePath := fmt.Sprintf("./%s", downloadedFile.Filename)
				downloadedPath := fmt.Sprintf("./downloads/%s", downloadedFile.Filename)
				err = os.Rename(filePath, downloadedPath)
				if err != nil {
					fmt.Println("Error moving file:", err)
					return
				}
			}

		}
	}()

	go func() {
		for {
			files, err := dbservice.GetExpiredDownloadedFiles(ctx, 1*time.Hour)
			if err != nil {
				fmt.Println("Error getting expired downloaded files:", err)
				return
			}
			for _, file := range files {
				fmt.Println("Expired file:", file)
				filePath := fmt.Sprintf("./downloads/%s", file.Filename)
				err = os.Remove(filePath)
				if err != nil {
					fmt.Println("Error deleting file:", err)
					return
				}
				dbservice.RemoveDownloadedFile(ctx, file.ID)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	go func() {
		for {
			files, err := dbservice.GetUnuploadedFiles(ctx)
			if err != nil {
				fmt.Println("Error getting unuploaded files:", err)
				return
			}

			for _, file := range files {
				isDone, err := zgService.CheckFileStatus(ctx, file.Hash)
				if err != nil {
					fmt.Println("Error checking file status:", err)
					return
				}
				fmt.Printf("File name: %s, status: %t\n", file.Filename, isDone)
				if isDone {
					dbservice.SetUploaded(ctx, file.Filename)
					filePath := fmt.Sprintf("./%s", file.Filename)
					err = os.Remove(filePath)
					if err != nil {
						fmt.Println("Error deleting file:", err)
						return
					}
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	router := gin.Default()

	// Enable CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://zgdrive.local", "http://localhost:5173", "*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(config))

	// HealthCheck godoc
	// @Summary Health check
	// @Description Check if the API is running
	// @Produce plain
	// @Success 200 {string} string "OK"
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// @Summary Upload a file
	// @Description Upload a file to the system
	// @Accept multipart/form-data
	// @Produce plain
	// @Param file formData file true "File to upload"
	// @Success 200 {object} gin.H "File uploaded successfully. Transaction hash: {hash}"
	// @Failure 400 {object} gin.H "Error getting file"
	// @Failure 500 {object} gin.H "Error saving file or adding to database"
	// @Router /upload [post]
	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// store file in local directory
		err = c.SaveUploadedFile(file, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		hash, err := services.FileHash(file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		uploadedFile, err := dbservice.AddFile(ctx, file.Filename, hash, file.Size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		newFilesChan <- uploadedFile

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully. Transaction hash: " + hash})
	})

	// @Summary List all files
	// @Description Get a list of all files in the system
	// @Produce json
	// @Success 200 {array} model.File
	// @Failure 500 {string} string "Error listing files"
	// @Router /list [get]
	router.GET("/list", func(c *gin.Context) {
		files, err := dbservice.ListFiles(ctx)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error listing files: %v", err)
			return
		}
		c.JSON(http.StatusOK, files)
	})

	// @Summary Download a file
	// @Description Initiate download of a file by its ID
	// @Produce plain
	// @Param fileId path int true "File ID"
	// @Success 200 {object} gin.H "File downloaded successfully. File name: {filename}"
	// @Failure 400 {object} gin.H "Invalid file id"
	// @Failure 500 {object} gin.H "Error getting file by id or checking download status"
	// @Router /download/{fileId} [get]
	router.GET("/download/:fileId", func(c *gin.Context) {
		fileId := c.Param("fileId")
		fmt.Println("File ID:", fileId)
		fileIdInt, err := strconv.ParseInt(fileId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "fileId": fileIdInt, "status": "error"})
			return
		}

		file, err := dbservice.GetFileById(ctx, fileIdInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "fileId": fileIdInt, "status": "error"})
			return
		}

		isDone, err := dbservice.CheckIsFileAlreadyDownloaded(ctx, file.Hash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "fileId": file.ID, "status": "error"})
			return
		}
		if isDone {
			c.JSON(http.StatusOK, gin.H{"message": "File already downloaded or processing. File name: " + file.Filename, "status": "downloaded", "fileId": file.ID})
			return
		}

		downloadedFilesChan <- file

		c.JSON(http.StatusOK, gin.H{"message": "File downloaded successfully. File name: " + file.Filename, "status": "downloading", "fileId": file.ID})
	})

	// @Summary Check download status
	// @Description Check download status of a file by its ID
	// @Produce json
	// @Param fileId path int true "File ID"
	// @Success 200 {object} gin.H "Download status: {status}"
	// @Router /downloadStatus/{fileId} [get]
	router.GET("/downloadStatus/:fileId", func(c *gin.Context) {
		fileId := c.Param("fileId")
		fileIdInt, err := strconv.ParseInt(fileId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "fileId": fileIdInt, "status": "error"})
			return
		}

		file, err := dbservice.GetFileById(ctx, fileIdInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "fileId": fileIdInt, "status": "error"})
			return
		}

		isDone, err := dbservice.CheckDownloadStatus(ctx, file.Hash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "fileId": fileIdInt, "status": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"fileId": fileIdInt, "status": isDone})
	})

	// @Summary List all downloaded files
	// @Description Get a list of all downloaded files
	// @Produce json
	// @Success 200 {array} model.File
	// @Failure 500 {object} gin.H "Error listing downloaded files"
	// @Router /downloaded [get]
	router.GET("/downloaded", func(c *gin.Context) {
		files, err := dbservice.ListDownloadedFiles(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, files)
	})

	// @Summary Download a file
	// @Description Download a file by its ID, only if it is downloaded
	// @Produce plain
	// @Param fileId path int true "File ID"
	// @Success 200 {object} gin.H "File downloaded successfully. File name: {filename}"
	// @Failure 400 {object} gin.H "Invalid file id"
	// @Failure 500 {object} gin.H "Error getting file by id"
	router.GET("/downloaded/:fileId", func(c *gin.Context) {
		fileId := c.Param("fileId")
		fileIdInt, err := strconv.ParseInt(fileId, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file, err := dbservice.GetFileById(ctx, fileIdInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("File:", file)

		// serve file from downloaded directory
		filePath := fmt.Sprintf("./downloads/%s", file.Filename)
		c.File(filePath)
	})

	// Add Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}
