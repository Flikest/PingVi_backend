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

// Logup godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided credentials.
// @Description  - Validates email format and password strength
// @Description  - Hashes password before storing
// @Description  - Returns user data without sensitive information
// @Tags         sso
// @Accept       json
// @Produce      json
// @Param        request  body  dto.CreateUserRequest  true  "User registration data"  example({"email":"user@example.com","password":"SecurePass123!","name":"John Doe"})
// @Success      201  {object}  dto.UserResponse  "Created user"
// @Failure      400  {object}  map[string]interface{}  "Invalid request body"  example({"error":"invalid email format"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to create user"})
// @Router       /logup [post]
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

// Login godoc
// @Summary      Authenticate user
// @Description  Logs in a user with email and password. Returns access and refresh tokens.
// @Description  - Sets HTTP-only cookies for access_token (15min) and refresh_token (60 days)
// @Description  - Supports 2FA if enabled on the account
// @Tags         sso
// @Accept       json
// @Produce      json
// @Param        request  body  dto.LogInRequest  true  "Login credentials"  example({"email":"user@example.com","password":"SecurePass123!"})
// @Success      200  {object}  dto.LoginResponse  "Login successful with tokens"
// @Failure      400  {object}  map[string]interface{}  "Invalid credentials"  example({"error":"invalid email or password"})
// @Failure      401  {object}  map[string]interface{}  "Authentication failed"  example({"error":"2FA required"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to login"})
// @Router       /login [post]
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

// Logout godoc
// @Summary      Logout user
// @Description  Invalidates the user's refresh token and clears session.
// @Description  - Requires authentication
// @Description  - Deletes the refresh token from database
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.LogoutRequest  true  "Logout data"  example({"refresh_token":"abc123def456"})
// @Success      200  {object}  map[string]interface{}  "Logout successful"  example({"user_id":"123e4567-e89b-12d3-a456-426614174000"})
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid refresh token"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to logout"})
// @Router       /logout [post]
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

// UpdateUser godoc
// @Summary      Update user information
// @Description  Updates the authenticated user's profile information.
// @Description  - Can update: name, email, avatar, etc.
// @Description  - Email update may require re-verification
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.UpdateUserRequest  true  "User update data"  example({"name":"John Updated","email":"newemail@example.com"})
// @Success      200  {object}  dto.UserResponse  "Updated user"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid email format"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to update user"})
// @Router       / [put]
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

// UpdatePassword godoc
// @Summary      Update user password
// @Description  Updates the authenticated user's password. Requires current password verification.
// @Description  - Validates password strength
// @Description  - Hashes new password before storing
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.UpdatePasswordRequest  true  "Password update data"  example({"current_password":"OldPass123!","new_password":"NewPass456!"})
// @Success      200  {object}  dto.UserResponse  "User with updated password"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"current password is incorrect"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to update password"})
// @Router       /password [patch]
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

// Verify2FA godoc
// @Summary      Verify 2FA code
// @Description  Verifies a 2FA (Two-Factor Authentication) code for the user.
// @Description  - Requires email and 2FA code
// @Description  - Returns tokens upon successful verification
// @Tags         sso
// @Accept       json
// @Produce      json
// @Param        request  body  dto.TwoFaRequest  true  "2FA verification data"  example({"email":"user@example.com","code":"123456"})
// @Success      200  {object}  dto.LoginResponse  "2FA verified successfully with tokens"
// @Failure      400  {object}  map[string]interface{}  "Invalid code"  example({"error":"invalid 2FA code"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to verify 2FA"})
// @Router       /2fa [post]
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

// EnableTwoFactorAuth godoc
// @Summary      Enable 2FA for user
// @Description  Enables Two-Factor Authentication for the authenticated user.
// @Description  - Returns a secret key for TOTP setup
// @Description  - User must scan QR code with authenticator app
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Success      200  {object}  dto.TwoFactorAuthResponse  "2FA enabled with secret key"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to enable 2FA"})
// @Router       /enable_2fa [patch]
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

// DisableTwoFactorAuth godoc
// @Summary      Disable 2FA for user
// @Description  Disables Two-Factor Authentication for the user.
// @Description  - Requires user ID in request body
// @Description  - Removes 2FA secret from user account
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  object  true  "User ID"  example({"user_id":"123e4567-e89b-12d3-a456-426614174000"})
// @Success      200  {object}  dto.UserResponse  "2FA disabled for user"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid user id"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to disable 2FA"})
// @Router       /disable_2fa [patch]

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

// SendTemporaryLink godoc
// @Summary      Send password recovery link
// @Description  Sends a temporary link to the user's email for password recovery.
// @Description  - Link expires after 15 minutes
// @Description  - User must click the link to reset password
// @Tags         sso
// @Accept       json
// @Produce      json
// @Param        request  body  object  true  "User email"  example({"email":"user@example.com"})
// @Success      200  {object}  map[string]interface{}  "Temporary link sent"  example({"message":"the link has been sent to your email: user@example.com"})
// @Failure      400  {object}  map[string]interface{}  "Invalid email"  example({"error":"email not found"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to send email"})
// @Router       /temporary_link [post]
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

// RecoverPassword godoc
// @Summary      Recover password
// @Description  Resets the user's password using a temporary link token.
// @Description  - Token must be valid and not expired
// @Description  - Redirects to login page after successful reset
// @Tags         sso
// @Accept       json
// @Produce      json
// @Param        request  body  dto.RecoverPasswordRequest  true  "Password recovery data"  example({"token":"recovery_token_123","new_password":"NewSecurePass456!"})
// @Success      307  {string}  string  "Redirect to login page"  example("Redirecting to https://pingvi.com/login")
// @Failure      400  {object}  map[string]interface{}  "Invalid token"  example({"error":"invalid or expired token"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to recover password"})
// @Router       /recover_password [post]
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

// SelectUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieves user information by user ID.
// @Description  - Returns public user information
// @Description  - Does not return sensitive data (password hash, etc.)
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        id  path  string  true  "User ID"  example("123e4567-e89b-12d3-a456-426614174000")
// @Success      200  {object}  dto.UserResponse  "User information"
// @Failure      400  {object}  map[string]interface{}  "Invalid user ID"  example({"error":"invalid user id"})
// @Failure      404  {object}  map[string]interface{}  "User not found"  example({"error":"user not found"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to get user"})
// @Router       /{id} [get]
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

// SelectUserSessions godoc
// @Summary      Get user sessions
// @Description  Retrieves all active sessions for the authenticated user.
// @Description  - Returns session information (device, IP, last activity)
// @Description  - Useful for managing active sessions
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Success      200  {array}  dto.SessionResponse  "List of active sessions"
// @Failure      400  {object}  map[string]interface{}  "Invalid token"  example({"error":"invalid token"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to get sessions"})
// @Router       /sessions [get]
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

// RefreshTokens godoc
// @Summary      Refresh access token
// @Description  Refreshes the access token using a valid refresh token.
// @Description  - Accepts tokens from headers or cookies
// @Description  - Returns new access and refresh token pair
// @Description  - Old refresh token is invalidated
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  false  "Bearer access token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        X-Refresh-Token  header  string  false  "Refresh token"  example("abc123def456")
// @Success      200  {object}  dto.TokenResponse  "New token pair"  example({"access_token":"new_access_token","refresh_token":"new_refresh_token"})
// @Failure      400  {object}  map[string]interface{}  "Invalid tokens"  example({"error":"token mismatch"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to refresh tokens"})
// @Router       /refresh [post]
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

// DeleteUser godoc
// @Summary      Delete user account
// @Description  Permanently deletes the authenticated user's account.
// @Description  - Requires authentication
// @Description  - Deletes all user data (sessions, messages, etc.)
// @Description  - This action cannot be undone
// @Tags         sso
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Success      200  {object}  map[string]interface{}  "User deleted successfully"  example({"user_id":"123e4567-e89b-12d3-a456-426614174000"})
// @Failure      400  {object}  map[string]interface{}  "Invalid token"  example({"error":"invalid token"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to delete user"})
// @Router       / [delete]
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
