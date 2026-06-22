package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	activatemessaging "github.com/Flikest/PingVi_backend/pkg/activate_messaging"
	"github.com/Flikest/PingVi_backend/pkg/geocoder"
	passwordhash "github.com/Flikest/PingVi_backend/pkg/password_hash"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	verificationcode "github.com/Flikest/PingVi_backend/pkg/verification_code"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// FIXME: переписать все на слоеную архитектуру, сейчас слой овечает и за бизнес логику и за рабуту с бд

func (r *RepositorySSO) createSession(ctx context.Context, tx pgx.Tx, userID uuid.UUID, device string, locale string) (response dto.LogInResponse, status int, err error) {
	session := dto.Session{}

	var isUserNotSession bool

	selectedSessionDataQuery := `
		SELECT id, access_token, refresh_token, created_at, access_expires_at, refresh_expires_at, locale
	 	FROM sessions 
	 	WHERE user_id=$1 AND user_device=$2
	`
	if err := tx.QueryRow(ctx, selectedSessionDataQuery, userID, device).Scan(
		&session.ID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.CreatedAt,
		&session.AccessExpiresAt,
		&session.RefreshExpiresAt,
		&session.Locale); err != nil && err != pgx.ErrNoRows {
		r.Log.Error("error with scanning session data: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	} else if err == pgx.ErrNoRows {
		isUserNotSession = true
	}

	accessToken, err := tokens.CreateAccessToken(userID, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		r.Log.Error("error with creating access token", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	refreshToken, err := tokens.CreateRefreshToken(128)
	if err != nil {
		r.Log.Error("error with creating refresh token: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	now := time.Now()

	if isUserNotSession == true {
		insertSessionQuery := `
			INSERT INTO
			sessions (id, user_id, access_token, refresh_token, user_device, created_at, access_expires_at, refresh_expires_at, locale)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		uuid, err := uuid.NewV7()
		if err != nil {
			r.Log.Error("error with generating uuid: ", "error", err)
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		_, err = tx.Exec(ctx, insertSessionQuery,
			uuid,
			userID,
			accessToken,
			refreshToken,
			device,
			now,
			now.Add(time.Minute*15),
			now.AddDate(0, 2, 0),
			locale)
		if err != nil {
			r.Log.Error("error with creating user session: ", "error", err)
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		return dto.LogInResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, http.StatusCreated, nil
	}

	updateSessionQuery := `
		UPDATE sessions
		SET access_token=$2, refresh_token=$3, created_at=$4, access_expires_at=$5, refresh_expires_at=$6, locale=$7
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateSessionQuery,
		session.ID,
		accessToken,
		refreshToken,
		now,
		now.Add(time.Minute*15),
		now.AddDate(0, 2, 0),
		locale)
	if err != nil {
		r.Log.Error("error with updating user session: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	return dto.LogInResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK, nil
}

func (r *RepositorySSO) CreateUser(ctx context.Context, u dto.CreateUserRequest) (user dto.User, err error) {
	insertUserQuery := `
		INSERT INTO 
		users (id, name, email, phone_number, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *
	`

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("failed to create transaction", "error", err)
		return dto.User{}, err
	}
	defer tx.Rollback(ctx)

	userID, err := uuid.NewV7()
	if err != nil {
		r.Log.Error("error with generating uuid v7 in create user function: ", "error", err)
		return dto.User{}, err
	}

	row := tx.QueryRow(ctx, insertUserQuery, userID, u.Name, u.Email, u.PhoneNumber, u.Password)

	if err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.Email,
		&user.ProfileStatus,
		&user.Description,
		&user.RiskProfile,
		&user.IsPrivate,
		&user.TwoFactorEnabled,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt); err != nil {
		r.Log.Error("error with adding database response: ", "error", err)
		return dto.User{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Error("failed to committing transaction", "error", err)
		return dto.User{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

func (r *RepositorySSO) Login(ctx context.Context, lr dto.LogInRequest) (response dto.LogInResponse, httpStatus int, err error) {
	if lr.Email == "" || lr.Password == "" {
		r.Log.Error("all fields must be filled in")
		return dto.LogInResponse{}, http.StatusBadRequest, fmt.Errorf("all fields must be filled in")
	}

	if lr.Lat == "" || lr.Long == "" || lr.Device == "" {
		r.Log.Error("the long and lat fields must be filled in")
		return dto.LogInResponse{}, http.StatusBadRequest, fmt.Errorf("the long and lat and device fields must be filled in")
	}

	locale, err := geocoder.Geocoder(fmt.Sprintf("%s, %s", lr.Long, lr.Lat), "ru")
	if err != nil {
		r.Log.Error("error with getting user locale: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error with creating transaction: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}
	defer tx.Rollback(ctx)

	userData := struct {
		ID               uuid.UUID
		PasswordHash     string
		TwoFactorEnabled bool
	}{}

	selectedUserDataQuery := "SELECT id, password_hash, two_factor_enabled FROM USERS WHERE email=$1"
	if err := tx.QueryRow(ctx, selectedUserDataQuery, lr.Email).Scan(
		&userData.ID,
		&userData.PasswordHash,
		&userData.TwoFactorEnabled); err != nil {
		if err == pgx.ErrNoRows {
			r.Log.Error("such user does not exist: ", "error", err)
			return dto.LogInResponse{}, http.StatusBadRequest, fmt.Errorf("such user does not exist")
		}
		r.Log.Error("error with scanning user data: ", "error", err)
		return dto.LogInResponse{}, http.StatusBadRequest, err
	}

	check := passwordhash.Checking(lr.Password, userData.PasswordHash)
	if check != true {
		r.Log.Error("error with checking password", "error", fmt.Errorf("incorrect password"))
		return dto.LogInResponse{}, http.StatusBadRequest, fmt.Errorf("incorrect password")
	} else if check && userData.TwoFactorEnabled {
		jwt2FAToken, err := tokens.Create2FaToken(userData.ID, fmt.Sprintf("%s %s", locale.Country, locale.City), lr.Device, []byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			r.Log.Error("error with creating 2fa token")
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		code := verificationcode.GenerateVerificationCode()

		create2FaSessionQuery := `
			INSERT INTO verification_codes (id, user_id, verification_code, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id) 
			DO UPDATE SET verification_code = EXCLUDED.verification_code, created_at=EXCLUDED.created_at, expires_at = EXCLUDED.expires_at
		`
		uuid, err := uuid.NewV7()
		if err != nil {
			r.Log.Error("error with creating uuid v7: ", "error", err)
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		_, err = tx.Exec(ctx, create2FaSessionQuery, uuid, userData.ID, code, time.Now(), time.Now().Add(time.Minute*5))
		if err != nil {
			r.Log.Error("error with inserting data on verification_codes table:", "error", err)
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		if err := tx.Commit(ctx); err != nil {
			r.Log.Error("error with commiting transaction: ", "error", err)
			return dto.LogInResponse{}, http.StatusInternalServerError, err
		}

		status, err := activatemessaging.PushVerifyMessageToEmail(ctx, code, lr.Email)
		if err != nil {
			r.Log.Error("error with pushing verify message", "error", err)
			return dto.LogInResponse{}, status, err
		}

		return dto.LogInResponse{
			AccessToken:  "",
			RefreshToken: "",
			Jwt2FAToken:  jwt2FAToken,
		}, http.StatusCreated, nil
	}

	response, status, err := r.createSession(ctx, tx, userData.ID, lr.Device, fmt.Sprintf("%s %s", locale.Country, locale.City))
	if err != nil {
		r.Log.Error("error with create session: ", "error", err)
		return dto.LogInResponse{}, status, err
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	return response, status, err
}

func (r *RepositorySSO) Logout(ctx context.Context, lr dto.LogoutRequest) (userID uuid.UUID, err error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error with creating transaction: ", "error", err)
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        DELETE FROM sessions 
        WHERE user_id = $1 AND user_device = $2
    `

	result, err := tx.Exec(ctx, query, lr.UserID, lr.Device)
	if err != nil {
		r.Log.Error("error with deleting session: ", "error", err)
		return uuid.Nil, fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.Log.Warn("no session found to delete", "user_id", lr.UserID, "device", lr.Device)
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return lr.UserID, nil
}

func (r *RepositorySSO) UpdateUser(ctx context.Context, ur dto.UpdateUserRequest) (user dto.User, err error) {
	if ur.ID == uuid.Nil {
		return dto.User{}, fmt.Errorf("user ID is required")
	}

	if ur.Email == "" {
		return dto.User{}, fmt.Errorf("email is required")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error with creating transaction: ", "error", err)
		return dto.User{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err = tx.QueryRow(ctx, checkQuery, ur.ID).Scan(&exists)
	if err != nil {
		return dto.User{}, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return dto.User{}, fmt.Errorf("user with ID %v not found", ur.ID)
	}

	var emailExists bool
	emailCheckQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`
	err = tx.QueryRow(ctx, emailCheckQuery, ur.Email, ur.ID).Scan(&emailExists)
	if err != nil {
		return dto.User{}, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if emailExists {
		return dto.User{}, fmt.Errorf("email %s is already taken by another user", ur.Email)
	}

	query := `
        UPDATE users
        SET name = $2, 
            avatar = $3, 
            email = $4,
			phone_number = $5,
            profile_status = $6, 
            description = $7, 
            risk_profile = $8, 
            is_private = $9, 
            updated_at = $10
        WHERE id = $1
        RETURNING *
    `

	err = tx.QueryRow(ctx, query,
		ur.ID,
		ur.Name,
		ur.Avatar,
		ur.Email,
		ur.PhoneNumber,
		ur.ProfileStatus,
		ur.Description,
		ur.RiskProfile,
		ur.IsPrivate,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.Email,
		&user.ProfileStatus,
		&user.Description,
		&user.RiskProfile,
		&user.IsPrivate,
		&user.IsActivated,
		&user.TwoFactorEnabled,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		r.Log.Error("error with updating user: ", "error", err)
		return dto.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return dto.User{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

func (r *RepositorySSO) UpdatePassword(ctx context.Context, upr dto.UpdatePasswordRequest) (user dto.User, status int, err error) {
	if upr.UserID == uuid.Nil || len(upr.NewPassword) < 8 || upr.OldPassword == "" {
		r.Log.Error("invalid input")
		return dto.User{}, http.StatusBadRequest, fmt.Errorf("invalid input: user ID required, new password must be at least 8 characters, old password required")
	}

	if upr.OldPassword == upr.NewPassword {
		r.Log.Error("new password must be different from old password")
		return dto.User{}, http.StatusBadRequest, fmt.Errorf("new password must be different from old password")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error with creating transaction: ", "error", err)
		return dto.User{}, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	selectUserOldPass := `
        SELECT password_hash FROM users WHERE id = $1
    `

	var oldPass string
	if err := tx.QueryRow(ctx, selectUserOldPass, upr.UserID).Scan(&oldPass); err != nil {
		if err == pgx.ErrNoRows {
			r.Log.Error("user not found", "user_id", upr.UserID)
			return dto.User{}, http.StatusBadRequest, fmt.Errorf("user not found")
		}
		r.Log.Error("error with selecting old password from user table: ", "error", err)
		return dto.User{}, http.StatusInternalServerError, fmt.Errorf("failed to get user password: %w", err)
	}

	if !passwordhash.Checking(upr.OldPassword, oldPass) {
		r.Log.Error("incorrect password", "user_id", upr.UserID)
		return dto.User{}, http.StatusBadRequest, fmt.Errorf("incorrect old password")
	}

	newPasswordHash, err := passwordhash.Hashing(upr.NewPassword)
	if err != nil {
		r.Log.Error("error with hashing password: ", "error", err)
		return dto.User{}, http.StatusInternalServerError, fmt.Errorf("failed to hash new password: %w", err)
	}

	updateQuery := `
        UPDATE users
        SET password_hash = $2, updated_at = $3
        WHERE id = $1
        RETURNING *
    `

	if err := tx.QueryRow(ctx, updateQuery, upr.UserID, newPasswordHash, time.Now()).Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.Email,
		&user.ProfileStatus,
		&user.Description,
		&user.RiskProfile,
		&user.IsPrivate,
		&user.IsActivated,
		&user.TwoFactorEnabled,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		r.Log.Error("error with updating user password: ", "error", err)
		return dto.User{}, http.StatusInternalServerError, fmt.Errorf("failed to update password: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return dto.User{}, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, http.StatusOK, nil
}

func (r *RepositorySSO) Verify2FA(ctx context.Context, tfr dto.TwoFaRequest) (response dto.LogInResponse, status int, err error) {
	if tfr.Code == "" || tfr.Jwt2FAToken == "" {
		r.Log.Error("verification code not entered")
		return dto.LogInResponse{}, http.StatusBadRequest, errors.New("verification code not entered")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error with creating transaction: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}
	defer tx.Rollback(ctx)

	claims, err := tokens.Verify2FA(tfr.Jwt2FAToken, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		r.Log.Error("error with verify jwt2fa token: ", "error", err)
		return dto.LogInResponse{}, http.StatusUnauthorized, errors.New("invalid or expired session")
	}

	selectedCodeQuery := `
		SELECT verification_code, expires_at FROM verification_codes WHERE user_id=$1
	`
	row := tx.QueryRow(ctx, selectedCodeQuery, claims.ID)

	var code string
	var expiresAt time.Time
	if err := row.Scan(&code, &expiresAt); err != nil {
		if err == pgx.ErrNoRows {
			r.Log.Error("No records with this user ID were found", "error", err)
			return dto.LogInResponse{}, http.StatusUnauthorized, errors.New("verification code not found or expired")
		}
		r.Log.Error("error with scaning verification_code from verification_codes table: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	if tfr.Code != code {
		r.Log.Error("The user entered an incorrect 2FA code")
		return dto.LogInResponse{}, http.StatusBadRequest, errors.New("the user entered an incorrect 2FA code")
	}

	if time.Now().After(expiresAt) {
		r.Log.Error("The 2FA code has expired")
		return dto.LogInResponse{}, http.StatusUnauthorized, errors.New("the verification code has expired")
	}

	deleteCodeQuery := "DELETE FROM verification_codes WHERE user_id=$1"
	_, err = tx.Exec(ctx, deleteCodeQuery, claims.ID)
	if err != nil {
		r.Log.Error("error with deleting from verification_codes table: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	response, status, err = r.createSession(ctx, tx, claims.ID, claims.Device, claims.Locale)
	if err != nil {
		r.Log.Error("error with create session: ", "error", err)
		return dto.LogInResponse{}, status, err
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("error with committing transaction: ", "error", err)
		return dto.LogInResponse{}, http.StatusInternalServerError, err
	}

	return response, status, nil
}

func (r *RepositorySSO) EnableTwoFactorAuth(ctx context.Context, userID uuid.UUID) (user dto.User, err error) {
	query := `
        UPDATE users 
        SET two_factor_enabled = TRUE
        WHERE id = $1
        RETURNING *
    `

	err = r.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.Email,
		&user.ProfileStatus,
		&user.Description,
		&user.RiskProfile,
		&user.IsPrivate,
		&user.IsActivated,
		&user.TwoFactorEnabled,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.Log.Warn("user not found", "userID", userID)
			return dto.User{}, fmt.Errorf("user with id %s not found", userID)
		}

		r.Log.Error("error scanning data from users table", "error", err, "userID", userID)
		return dto.User{}, err
	}

	r.Log.Info("2FA enabled successfully", "userID", userID)
	return user, nil
}

func (r *RepositorySSO) DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID) (user dto.User, err error) {
	query := `
        UPDATE users 
        SET two_factor_enabled = FALSE
        WHERE id = $1
        RETURNING *
    `

	err = r.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.Email,
		&user.ProfileStatus,
		&user.Description,
		&user.RiskProfile,
		&user.IsPrivate,
		&user.IsActivated,
		&user.TwoFactorEnabled,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.Log.Warn("user not found", "userID", userID)
			return dto.User{}, fmt.Errorf("user with id %s not found", userID)
		}

		r.Log.Error("error scanning data from users table", "error", err, "userID", userID)
		return dto.User{}, err
	}

	r.Log.Info("2FA disable successfully", "userID", userID)
	return user, nil
}

func (r *RepositorySSO) SendTemporaryLink(ctx context.Context, email string) (err error) {
	token, err := tokens.GenerateTemporaryToken()
	if err != nil {
		r.Log.Error("error creating temporary token:", "error", err)
		return err
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error creating transaction:", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var userID uuid.UUID
	queryGetUser := `SELECT id FROM users WHERE email = $1`
	err = tx.QueryRow(ctx, queryGetUser, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.Log.Info("user with such email does not exist", "email", email)
			return nil
		}
		r.Log.Error("error selecting user:", "error", err)
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		r.Log.Error("error generating uuid:", "error", err)
		return err
	}

	queryInsert := `
	INSERT INTO password_reset_tokens (id, user_id, token, expires_at, used)
	VALUES ($1, $2, $3, $4, $5)
    `
	_, err = tx.Exec(ctx, queryInsert, id, userID, token, time.Now().Add(15*time.Minute), false)
	if err != nil {
		r.Log.Error("error inserting token:", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Error("error committing transaction:", "error", err)
		return err
	}

	link := fmt.Sprintf("https://pingvi.com/recover_password/%s", token)

	_, err = activatemessaging.PushPasswordRecoveryMessage(ctx, link, email)
	if err != nil {
		r.Log.Error("erorr with pushinig recovery password message: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositorySSO) RecoverPassword(ctx context.Context, req dto.RecoverPasswordRequest) (int, error) {
	if req.Token == "" {
		r.Log.Error("token is empty")
		return http.StatusBadRequest, errors.New("token is required")
	}

	if req.NewPassword == "" {
		r.Log.Error("new password is empty")
		return http.StatusBadRequest, errors.New("new password is required")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("error creating transaction:", "error", err)
		return http.StatusInternalServerError, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	querySelect := `
        SELECT id, user_id, token, used, expires_at 
        FROM password_reset_tokens 
        WHERE token = $1
    `
	var prt dto.PasswordResetToken
	err = tx.QueryRow(ctx, querySelect, req.Token).Scan(
		&prt.ID,
		&prt.UserID,
		&prt.Token,
		&prt.Used,
		&prt.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.Log.Error("token not found")
			return http.StatusBadRequest, errors.New("invalid token")
		}
		r.Log.Error("error selecting token:", "error", err)
		return http.StatusInternalServerError, err
	}

	if prt.Used {
		r.Log.Error("token already used")
		return http.StatusBadRequest, errors.New("token already used")
	}

	if time.Now().After(prt.ExpiresAt) {
		r.Log.Error("token expired")
		return http.StatusRequestTimeout, errors.New("token expired")
	}

	hashNewPass, err := passwordhash.Hashing(req.NewPassword)
	if err != nil {
		r.Log.Error("error hashing password:", "error", err)
		return http.StatusInternalServerError, err
	}

	queryUpdate := `
        UPDATE users
        SET password = $1
        WHERE id = $2
    `
	_, err = tx.Exec(ctx, queryUpdate, hashNewPass, prt.UserID)
	if err != nil {
		r.Log.Error("error updating user password:", "error", err)
		return http.StatusInternalServerError, err
	}

	queryMarkUsed := `
        UPDATE password_reset_tokens
        SET used = TRUE
        WHERE id = $1
    `
	_, err = tx.Exec(ctx, queryMarkUsed, prt.ID)
	if err != nil {
		r.Log.Error("error marking token as used:", "error", err)
		return http.StatusInternalServerError, err
	}

	if err = tx.Commit(ctx); err != nil {
		r.Log.Error("error committing transaction:", "error", err)
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (r *RepositorySSO) SelectUserNameByID(ctx context.Context, userID uuid.UUID) (name string, err error) {
	if userID == uuid.Nil {
		r.Log.Error("invalid input: user id is required")
		return "", fmt.Errorf("invalid input: user id os required")
	}

	query := `
		SELECT name FROM users WHERE id=$1
	`

	if err := r.DB.QueryRow(ctx, query, userID).Scan(&name); err != nil {
		r.Log.Error("error with selecting username by id: ", "error", err)
		return "", err
	}

	return name, nil
}

func (r *RepositorySSO) SelectUserByID(ctx context.Context, userID uuid.UUID) (user dto.User, err error) {
	if userID == uuid.Nil {
		r.Log.Error("invalid input: ", "error", err)
		return dto.User{}, fmt.Errorf("invalid input: user id is required")
	}

	query := `
		SELECT *
		FROM users
		WHERE id = $1
	`

	if err := r.DB.QueryRow(ctx, query, userID).Scan(&user); err != nil {
		r.Log.Error("error with selecting user by id: ", "error", err)
		return dto.User{}, err
	}

	return user, nil
}

func (r *RepositorySSO) SelectAllUserSessions(ctx context.Context, userID uuid.UUID) (session []dto.Session, err error) {
	if userID == uuid.Nil {
		r.Log.Error("invalid input: user_id is required")
		return nil, fmt.Errorf("invalid input: userid is required")
	}

	query := `
        SELECT *
        FROM sessions 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		r.Log.Error("error with selecting all user sessions: ", "error", err)
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []dto.Session

	for rows.Next() {
		var session dto.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.UserDevice,
			&session.AccessToken,
			&session.RefreshToken,
			&session.CreatedAt,
			&session.AccessExpiresAt,
			&session.RefreshExpiresAt,
			&session.Locale,
			&session.LocaleImgUrl,
		)
		if err != nil {
			r.Log.Error("error with scanning session: ", "error", err)
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		r.Log.Error("error after iterating rows: ", "error", err)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return sessions, nil
}

func (r *RepositorySSO) DeleteUser(ctx context.Context, userID uuid.UUID) (ID uuid.UUID, err error) {
	if userID == uuid.Nil {
		r.Log.Error("invalid input: user_id is required")
		return uuid.Nil, fmt.Errorf("user ID is required")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		r.Log.Error("failed to create transaction", "error", err)
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err = tx.QueryRow(ctx, checkQuery, userID).Scan(&exists)
	if err != nil {
		r.Log.Error("error checking user existence", "error", err)
		return uuid.Nil, fmt.Errorf("failed to check user: %w", err)
	}

	if !exists {
		r.Log.Warn("user not found", "user_id", userID)
		return uuid.Nil, fmt.Errorf("user with ID %v not found", userID)
	}

	result, err := tx.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		r.Log.Error("error deleting user sessions", "error", err)
		return uuid.Nil, fmt.Errorf("failed to delete user sessions: %w", err)
	}

	sessionsDeleted := result.RowsAffected()
	r.Log.Info("deleted user sessions", "user_id", userID, "count", sessionsDeleted)

	result, err = tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		r.Log.Error("error deleting user", "error", err)
		return uuid.Nil, fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.Log.Warn("user not found during deletion", "user_id", userID)
		return uuid.Nil, fmt.Errorf("user with ID %v not found", userID)
	}

	if err := tx.Commit(ctx); err != nil {
		r.Log.Error("failed to commit transaction", "error", err)
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Log.Info("user deleted successfully", "user_id", userID)
	return userID, nil
}
