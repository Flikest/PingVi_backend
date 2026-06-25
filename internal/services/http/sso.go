package servicehttp

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
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

	ctx.SetCookie(
		"access_token",
		response.AccessToken,
		60*15,
		"/",
		"localhost",
		false,
		true,
	)

	ctx.SetCookie(
		"refresh_token",
		response.RefreshToken,
		3600*24*60,
		"/",
		"localhost",
		false,
		true,
	)

	ctx.JSON(status, response)
}

func (s *ServiceSSO) Logout(ctx *gin.Context) {
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

	body := dto.LogoutRequest{}
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

	uuid, err := s.Repository.Logout(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with logoutiing user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusUnauthorized, uuid)
}

func (s *ServiceSSO) UpdateUser(ctx *gin.Context) {
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

	body := dto.UpdateUserRequest{}
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.ID = userID

	response, err := s.Repository.UpdateUser(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *ServiceSSO) UpdatePassword(ctx *gin.Context) {
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

	body := dto.UpdatePasswordRequest{}

	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
	}

	body.UserID = userID

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

	user, err := s.Repository.EnableTwoFactorAuth(ctx.Request.Context(), userID)
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

func (s *ServiceSSO) SelectUserSessions(ctx *gin.Context) {
	var access string

	access = ctx.Request.Header.Get("Authorization")
	if access == "" {
		accessToken, err := ctx.Cookie("access_token")
		if err != nil {
			s.Log.Error("error with get access_token cookie: ", "error", err)
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		access = accessToken
	}

	payload, err := tokens.Verify(access, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, err := s.Repository.SelectAllSessionsByUserID(ctx.Request.Context(), payload.ID)
	if err != nil {
		s.Log.Error("error with selecting sessions by user id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, sessions)
}

func (s *ServiceSSO) RefreshTokens(ctx *gin.Context) {
	var oldAccessToken, oldRefreshToken string

	oldAccessToken, oldRefreshToken = ctx.Request.Header.Get("Authorization"), ctx.Request.Header.Get("X-Refresh-Token")

	payload, err := tokens.Verify(oldAccessToken, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with vetify pyload from access token: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	if oldAccessToken == "" && oldRefreshToken == "" {
		access, err := ctx.Cookie("access_token")
		if err != nil {
			s.Log.Error("error with get access_token cookie: ", "error", err)
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		refres, err := ctx.Cookie("refresh_token")
		if err != nil {
			s.Log.Error("error with get refresh_token cookie: ", "error", err)
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		oldAccessToken, oldRefreshToken = access, refres
	}

	now := time.Now()

	userTokens, err := s.Repository.SelectUserTokens(ctx.Request.Context(), payload.ID)
	if err != nil {
		s.Log.Error("error with select user tokens: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userTokens.RefreshToken != oldRefreshToken {
		s.Log.Warn("token mismatch")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token mismatch"})
		return
	}

	accessToken, err := tokens.CreateAccessToken(payload.ID, []byte(os.Getenv("JWT_SECRET")))

	refreshToken, err := tokens.CreateRefreshToken(128)
	if err != nil {
		s.Log.Error("error with create refresh token: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.Repository.UpdateTokens(ctx.Request.Context(), payload.ID, accessToken, refreshToken, now, now.Add(time.Minute*15), now.AddDate(0, 2, 0)); err != nil {
		s.Log.Error("error with update tokens: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.SetCookie(
		"access_token",
		accessToken,
		60*15,
		"/",
		"localhost",
		false,
		true,
	)

	ctx.SetCookie(
		"refresh_token",
		accessToken,
		3600*24*60,
		"/",
		"localhost",
		false,
		true,
	)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (s *ServiceSSO) DeleteUser(ctx *gin.Context) {
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

	id, err := s.Repository.DeleteUser(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with deleting user: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}
