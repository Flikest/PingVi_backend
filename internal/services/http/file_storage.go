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

// UploadToFileStorage godoc
//
//	@Summary		Upload files to storage
//	@Description	Upload one or multiple files to the file storage system. Files can be uploaded to public or private buckets based on category.
//	@Description	- Public buckets: avatar, emoji, sticker (accessible without authentication)
//	@Description	- Private buckets: default and other categories (require temporary URL for access)
//	@Description	- Maximum file size: 2GB per file
//	@Description	- Prohibited extensions: .exe, .bat, .sh (security reasons)
//	@Tags			file-storage
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true										"Bearer JWT token"						example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			files			formData	[]file					true										"Files to upload (max 2GB per file)"	collectionFormat(multi)
//	@Param			category		formData	string					false										"Category for file organization"		Enums(avatar, emoji, sticker, default)	example("avatar")
//	@Success		200				{object}	map[string]interface{}	"Returns array of file statuses"			example({"files":[{"filename":"profile.jpg","status":"success","bucket":"public/avatar/123e4567-e89b-12d3-a456-426614174000.jpg"}]})
//	@Failure		400				{object}	map[string]interface{}	"Invalid form data or no files provided"	example({"error":"no files provided"})
//	@Failure		400				{object}	map[string]interface{}	"File too large or prohibited type"			example({"files":[{"filename":"virus.exe","status":"failed","error":"This type of file is prohibited for security reasons"}]})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"						example({"error":"failed to generate file ID"})
//	@Router			/ [post]
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

// IssueTemporaryURL godoc
//
//	@Summary		Generate temporary URLs for files
//	@Description	Generate temporary access URLs for private files (valid for 24 hours).
//	@Description	- Public files (public/) return original URL without expiration
//	@Description	- Private files (private/) generate signed temporary URLs
//	@Description	- Only private files can generate temporary URLs
//	@Tags			file-storage
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true										"Bearer JWT token"									example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.IssueTemporaryURLRequest	true										"List of file URLs to generate temporary access"	example({"links":["private/catalog/123e4567-e89b-12d3-a456-426614174000.pdf","public/avatar/user-avatar.jpg"]})
//	@Success		200				{object}	map[string]interface{}			"Returns array of temporary URL responses"	example({"result":[{"temporary_url":"https://s3.amazonaws.com/private/catalog/123e4567...?X-Amz-Expires=86400&X-Amz-Signature=..."},{"temporary_url":"public/avatar/user-avatar.jpg"}]})
//	@Failure		400				{object}	map[string]interface{}			"Invalid request body or no URLs provided"	example({"error":"invalid body"})
//	@Failure		500				{object}	map[string]interface{}			"Internal server error"						example({"result":[{"error":"failed to generate temporary URL: invalid bucket"}]})
//	@Router			/temporary_url [post]
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

// UpdateInFileStorage godoc
//
//	@Summary		Update existing files
//	@Description	Replace existing files in storage with new ones.
//	@Description	- Requires the complete filepath of the file to update
//	@Description	- Filepath format: {bucket}/{category}/{uuid}.{ext}
//	@Description	- Maximum file size: 2GB per file
//	@Description	- Prohibited extensions: .exe, .bat, .sh
//	@Tags			file-storage
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true														"Bearer JWT token"							example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			filepath		formData	string					true														"Path of the file to update"				example("private/catalog/123e4567-e89b-12d3-a456-426614174000.pdf")
//	@Param			files			formData	[]file					true														"New files to upload (max 2GB per file)"	collectionFormat(multi)
//	@Success		200				{object}	map[string]interface{}	"Returns array of file statuses"							example({"files":[{"filename":"new_document.pdf","status":"success","bucket":"private/catalog/123e4567-e89b-12d3-a456-426614174000.pdf"}]})
//	@Failure		400				{object}	map[string]interface{}	"Invalid form data, no files provided, or missing filepath"	example({"error":"filepath is required"})
//	@Failure		400				{object}	map[string]interface{}	"Invalid filepath format"									example({"files":[{"filename":"doc.pdf","status":"failed","error":"invalid filepath: expected bucket/object"}]})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"										example({"error":"failed to update file"})
//	@Router			/ [put]
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

// DeleteFromFileStorage godoc
//
//	@Summary		Delete files from storage
//	@Description	Delete one or multiple files from the storage system by providing their paths.
//	@Description	- File paths can be full URLs or paths
//	@Description	- Format: {bucket}/{object} or full S3 URL
//	@Tags			file-storage
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true									"Bearer JWT token"				example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.DeleteRequest			true									"List of file paths to delete"	example({"paths":["public/avatar/user-avatar.jpg","private/catalog/123e4567-e89b-12d3-a456-426614174000.pdf"]})
//	@Success		200				{object}	map[string]interface{}	"Returns array of deletion statuses"	example({"results":[{"filename":"public/avatar/user-avatar.jpg","status":"success"},{"filename":"private/catalog/123e4567-e89b-12d3-a456-426614174000.pdf","status":"success"}]})
//	@Failure		400				{object}	map[string]interface{}	"Invalid request or no paths provided"	example({"error":"no paths provided"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"					example({"results":[{"filename":"invalid-path","status":"failed","error":"invalid path format, expected bucket/object"}]})
//	@Router			/ [delete]
func (s *ServiceFileStorage) DeleteFromFileStorage(ctx *gin.Context) {

	var req dto.DeleteRequest
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
