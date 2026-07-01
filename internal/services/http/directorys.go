package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateDirectory godoc
// @Summary      Create a new directory in a channel
// @Description  Creates a new directory in a channel. Requires 'manage_directories' permission (permission[4] == '1').
// @Tags         messenger-directories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.CreateDirectory  true  "Directory creation data"  example({"channel_id":"123e4567-e89b-12d3-a456-426614174000","name":"Documents"})
// @Success      201  {object}  dto.Directory  "Created directory"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      403  {object}  map[string]interface{}  "Insufficient permissions"  example({"error":"not enough rights"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to create directory"})
// @Router       /directorys [post]
func (s *ServiceMessenger) CreateDirectory(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.CreateDirectory
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with creating directory", "error", err)
		ctx.JSON(http.StatusBadRequest, "invalid body")
		return
	}

	body.UserID = userID

	permision, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error checking user rights: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if permision[4] != '1' {
		s.Log.Warn("the user does not have enough right")
		ctx.JSON(http.StatusBadRequest, "not enough rights")
		return
	}

	directory, err := s.Repository.CreateDirectory(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creating directory", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, directory)
}

// SelectAllDirectorys godoc
// @Summary      Get all directories in a channel
// @Description  Retrieves all directories from a specific channel.
// @Tags         messenger-directories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        channel_id  path  string  true  "Channel ID"  example("123e4567-e89b-12d3-a456-426614174000")
// @Success      200  {array}  dto.Directory  "List of directories"
// @Failure      400  {object}  map[string]interface{}  "Invalid channel ID"  example({"error":"invalid channel id"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to get directories"})
// @Router       /directorys/{channel_id} [get]
func (s *ServiceMessenger) SelectAllDirectorys(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing channel uuid, ", "error", err)
		ctx.JSON(http.StatusBadRequest, "invalid channel id")
		return
	}

	dirs, err := s.Repository.SelectAllDirectorys(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with selecting all directorys", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, dirs)
}

// UpdateDirectory godoc
// @Summary      Update a directory
// @Description  Updates an existing directory in a channel. Requires 'manage_directories' permission (permission[4] == '1').
// @Tags         messenger-directories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.UpdateDirectory  true  "Directory update data"  example({"id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"987fcdeb-51d2-12d3-a456-426614174000","name":"New Directory Name"})
// @Success      200  {object}  dto.Directory  "Updated directory"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      403  {object}  map[string]interface{}  "Insufficient permissions"  example({"error":"not enough rights"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to update directory"})
// @Router       /directorys [put]
func (s *ServiceMessenger) UpdateDirectory(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.UpdateDirectory
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, "invalid body")
		return
	}

	body.UserID = userID

	permision, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error checking user rights: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if permision[4] != '1' {
		s.Log.Warn("the user does not have enough right")
		ctx.JSON(http.StatusBadRequest, "not enough rights")
		return
	}

	directory, err := s.Repository.UpdateDirectory(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating directory: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, directory)
}

// DeleteDirectory godoc
// @Summary      Delete a directory
// @Description  Permanently deletes a directory from a channel. Requires 'manage_directories' permission (permission[4] == '1').
// @Tags         messenger-directories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.DeleteDirectory  true  "Directory deletion data"  example({"id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"987fcdeb-51d2-12d3-a456-426614174000"})
// @Success      200  {object}  map[string]interface{}  "Directory ID"  example({"directory_id":"123e4567-e89b-12d3-a456-426614174000"})
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      403  {object}  map[string]interface{}  "Insufficient permissions"  example({"error":"not enough rights"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to delete directory"})
// @Router       /directorys [delete]
func (s *ServiceMessenger) DeleteDirectory(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.DeleteDirectory
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, "invalid body")
		return
	}

	body.UserID = userID

	permision, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error checking user rights: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if permision[4] != '1' {
		s.Log.Warn("the user does not have enough right")
		ctx.JSON(http.StatusBadRequest, "not enough rights")
		return
	}

	if err := s.Repository.DeleteDirectory(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with updating directory: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, body.ID)
}
