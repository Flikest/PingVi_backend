package servicehttp

import (
	"fmt"
	"net/http"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceSSO) Logup(ctx *gin.Context) {
	body := dto.CreateUserRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := s.Repository.CreateUser(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creting user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (s *ServiceSSO) Login(ctx *gin.Context) {
	body := dto.LogInRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
	}

	response, status, err := s.Repository.Login(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with loginning user: ", "error", err)
		ctx.JSON(status, err)
		return
	}

	ctx.JSON(status, response)
}

func (s *ServiceSSO) Logout(ctx *gin.Context) {
	body := dto.LogoutRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	uuid, err := s.Repository.Logout(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with logoutiing user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusUnauthorized, uuid)
}

func (s *ServiceSSO) UpdateUser(ctx *gin.Context) {
	body := dto.UpdateUserRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	response, err := s.Repository.UpdateUser(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *ServiceSSO) UpdatePassword(ctx *gin.Context) {
	body := dto.UpdatePasswordRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
	}

	user, status, err := s.Repository.UpdatePassword(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating password: ", "error", err)
		ctx.JSON(status, err)
		return
	}

	ctx.JSON(status, user)
}

func (s *ServiceSSO) Verify2FA(ctx *gin.Context) {
	body := dto.TwoFaRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	response, status, err := s.Repository.Verify2FA(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with verefy 2 fa: ", "error", err)
		ctx.JSON(status, err)
		return
	}

	ctx.JSON(status, response)
}

func (s *ServiceSSO) EnableTwoFactorAuth(ctx *gin.Context) {
	body := struct {
		UserID uuid.UUID `json:"user_id"`
	}{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := s.Repository.EnableTwoFactorAuth(ctx.Request.Context(), body.UserID)
	if err != nil {
		s.Log.Error("error with enabling 2fa: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *ServiceSSO) DisableTwoFactorAuth(ctx *gin.Context) {
	body := struct {
		UserID uuid.UUID `json:"user_id"`
	}{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := s.Repository.DisableTwoFactorAuth(ctx.Request.Context(), body.UserID)
	if err != nil {
		s.Log.Error("error with disabling 2 fa: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *ServiceSSO) SendTemporaryLink(ctx *gin.Context) {
	body := struct {
		Email string `json:"email"`
	}{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	err := s.Repository.SendTemporaryLink(ctx.Request.Context(), body.Email)
	if err != nil {
		s.Log.Error("error with sending temporary link", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("the link has been sent to your email: %s", body.Email))
}

func (s *ServiceSSO) RecoverPassword(ctx *gin.Context) {
	body := dto.RecoverPasswordRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	status, err := s.Repository.RecoverPassword(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with recovering: ", "error", err)
		ctx.JSON(status, err)
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, "https://pingvi.com/login")
}

func (s *ServiceSSO) SelectUserByID(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		s.Log.Error("error with parsing uuid from param: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	user, err := s.Repository.SelectUserByID(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with selecting user by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *ServiceSSO) SelectAllUserSession(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		s.Log.Error("error with parsing uuid from param: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	sessions, err := s.Repository.SelectAllUserSessions(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with selecting session: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, sessions)
}

func (s *ServiceSSO) DeleteUser(ctx *gin.Context) {
	body := struct {
		UserID uuid.UUID `json:"user_id"`
	}{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	id, err := s.Repository.DeleteUser(ctx.Request.Context(), body.UserID)
	if err != nil {
		s.Log.Error("error with deleting user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}
