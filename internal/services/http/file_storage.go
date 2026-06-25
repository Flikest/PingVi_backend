package servicehttp

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	s3urlparse "github.com/Flikest/PingVi_backend/pkg/s3_url_parse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceFileStorage) UploadToFileStorage(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		s.Log.Error("invalid form", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	const maxSingleFileSize = 2000 << 20
	publicCategories := []string{"avatar", "emoji", "sticker"}
	var results []dto.FileStatus

	for _, file := range files {
		status := dto.FileStatus{Filename: file.Filename}

		if file.Size > maxSingleFileSize {
			status.Status = "failed"
			status.Error = "The file is too large. The maximum size is 2 GB"
			results = append(results, status)
			continue
		}

		ext := filepath.Ext(file.Filename)
		extLower := strings.ToLower(ext)
		if extLower == ".exe" || extLower == ".bat" || extLower == ".sh" {
			status.Status = "failed"
			status.Error = "This type of file is prohibited for security reasons"
			results = append(results, status)
			continue
		}

		fileID, err := uuid.NewV7()
		if err != nil {
			s.Log.Error("failed to generate file ID", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		category := ctx.PostForm("category")
		if category == "" {
			category = "default"
		}

		bucketName := "private"
		if slices.Contains(publicCategories, category) {
			bucketName = "public"
		}

		objectName := fmt.Sprintf("%s/%s%s", category, fileID.String(), extLower)

		src, err := file.Open()
		if err != nil {
			status.Status = "failed"
			status.Error = "Failed to open file: " + err.Error()
			results = append(results, status)
			continue
		}

		path, err := s.Repository.UploadFile(ctx.Request.Context(), bucketName, objectName, src, file.Size)
		src.Close()

		if err != nil {
			s.Log.Error("failed to upload file", "filename", file.Filename, "error", err)
			status.Status = "failed"
			status.Error = "Failed to upload file: " + err.Error()
			results = append(results, status)
			continue
		}

		results = append(results, dto.FileStatus{
			Filename: file.Filename,
			Status:   "success",
			Bucket:   path,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files": results,
	})
}

func (s *ServiceFileStorage) IssueTemporaryURL(ctx *gin.Context) {
	var body dto.IssueTemporaryURLRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		s.Log.Error("invalid request body", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if len(body.Links) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no URLs provided"})
		return
	}

	var result []dto.IssueTemporaryURLResponse

	for _, link := range body.Links {
		if strings.HasPrefix(link, "public/") {
			result = append(result, dto.IssueTemporaryURLResponse{
				TemporaryURL: link,
			})
			continue
		}

		// http://localhost:9000/private/catalog/uuid.ext
		bucketName, objectName, err := s3urlparse.ParseS3Path(link)
		if err != nil {
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "invalid filepath: " + err.Error(),
			})
			continue
		}

		if bucketName != "private" {
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "temporary URLs can only be generated for private files",
			})
			continue
		}

		temporaryURL, err := s.Repository.GenerateTemporaryURL(
			ctx.Request.Context(),
			bucketName,
			objectName,
			24*time.Hour,
			nil,
		)

		if err != nil {
			s.Log.Error("failed to generate temporary URL",
				"bucket", bucketName,
				"object", objectName,
				"error", err)
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "failed to generate temporary URL: " + err.Error(),
			})
			continue
		}

		result = append(result, dto.IssueTemporaryURLResponse{
			TemporaryURL: temporaryURL,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"result": result})
}

func (s *ServiceFileStorage) UpdateInFileStorage(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		s.Log.Error("invalid form", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	filePath := ctx.PostForm("filepath")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "filepath is required"})
		return
	}

	const maxSingleFileSize = 2000 << 20
	var results []dto.FileStatus

	for _, file := range files {
		status := dto.FileStatus{Filename: file.Filename}

		if file.Size > maxSingleFileSize {
			status.Status = "failed"
			status.Error = "The file is too large. The maximum size is 2 GB"
			results = append(results, status)
			continue
		}

		ext := filepath.Ext(file.Filename)
		extLower := strings.ToLower(ext)
		if extLower == ".exe" || extLower == ".bat" || extLower == ".sh" {
			status.Status = "failed"
			status.Error = "This type of file is prohibited for security reasons"
			results = append(results, status)
			continue
		}

		bucketName, objectName, err := s3urlparse.ParseS3Path(filePath)
		if err != nil {
			status.Status = "failed"
			status.Error = "invalid filepath: " + err.Error()
			results = append(results, status)
			continue
		}

		src, err := file.Open()
		if err != nil {
			status.Status = "failed"
			status.Error = "Failed to open file: " + err.Error()
			results = append(results, status)
			continue
		}

		path, err := s.Repository.UploadFile(ctx.Request.Context(), bucketName, objectName, src, file.Size)
		src.Close()

		if err != nil {
			s.Log.Error("failed to update file", "filename", file.Filename, "error", err)
			status.Status = "failed"
			status.Error = "Failed to update file: " + err.Error()
			results = append(results, status)
			continue
		}

		results = append(results, dto.FileStatus{
			Filename: file.Filename,
			Status:   "success",
			Bucket:   path,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files": results,
	})
}

func (s *ServiceFileStorage) DeleteFromFileStorage(ctx *gin.Context) {
	type DeleteRequest struct {
		Paths []string `json:"paths"`
	}

	var req DeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Paths) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no paths provided"})
		return
	}

	var results []dto.FileStatus

	for _, path := range req.Paths {
		u, err := url.Parse(path)
		if err != nil {
			s.Log.Error("invalid URL", "path", path, "error", err)
			results = append(results, dto.FileStatus{
				Filename: path,
				Status:   "failed",
				Error:    "invalid path",
			})
			continue
		}

		urlPath := strings.TrimPrefix(u.Path, "/")
		parts := strings.SplitN(urlPath, "/", 2)

		if len(parts) < 2 {
			results = append(results, dto.FileStatus{
				Filename: path,
				Status:   "failed",
				Error:    "invalid path format, expected bucket/object",
			})
			continue
		}

		bucketName := parts[0]
		objectName := parts[1]

		err = s.Repository.DeleteFile(ctx.Request.Context(), bucketName, objectName)
		if err != nil {
			s.Log.Error("failed to delete file", "path", path, "error", err)
			results = append(results, dto.FileStatus{
				Filename: path,
				Status:   "failed",
				Error:    "failed to delete file: " + err.Error(),
			})
			continue
		}

		results = append(results, dto.FileStatus{
			Filename: path,
			Status:   "success",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"results": results})
}
