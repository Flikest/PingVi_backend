package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
