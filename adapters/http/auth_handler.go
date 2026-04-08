package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type AuthHandler struct {
	requestOTPUC     *usecases.RequestOTPUseCase
	verifyOTPUC      *usecases.VerifyOTPUseCase
	loginPasswordUC  *usecases.LoginPasswordUseCase
	refreshTokenUC   *usecases.RefreshTokenUseCase
	logoutUC         *usecases.LogoutUseCase
	getCurrentUserUC *usecases.GetCurrentUserUseCase
}

func NewAuthHandler(
	requestOTPUC *usecases.RequestOTPUseCase,
	verifyOTPUC *usecases.VerifyOTPUseCase,
	loginPasswordUC *usecases.LoginPasswordUseCase,
	refreshTokenUC *usecases.RefreshTokenUseCase,
	logoutUC *usecases.LogoutUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
) *AuthHandler {
	return &AuthHandler{
		requestOTPUC:     requestOTPUC,
		verifyOTPUC:      verifyOTPUC,
		loginPasswordUC:  loginPasswordUC,
		refreshTokenUC:   refreshTokenUC,
		logoutUC:         logoutUC,
		getCurrentUserUC: getCurrentUserUC,
	}
}

type requestOTPRequest struct {
	Phone string `json:"phone"`
}

func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("auth_handler.RequestOTP.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Phone == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Phone is required")
		return
	}

	_, err := h.requestOTPUC.Execute(r.Context(), usecases.RequestOTPInput{
		Phone: req.Phone,
	})
	if err != nil {
		slog.Error("auth_handler.RequestOTP.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to send OTP")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent"})
}

type verifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

type authResponse struct {
	AccessToken  string               `json:"accessToken"`
	RefreshToken string               `json:"refreshToken"`
	User         *usecases.UserOutput `json:"user"`
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("auth_handler.VerifyOTP.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Phone == "" || req.OTP == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Phone and OTP are required")
		return
	}

	output, err := h.verifyOTPUC.Execute(r.Context(), usecases.VerifyOTPInput{
		Phone: req.Phone,
		OTP:   req.OTP,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidOTP):
			writeJSONError(w, http.StatusUnauthorized, "invalid_otp", "Invalid OTP code")
		case errors.Is(err, usecases.ErrOTPExpired):
			writeJSONError(w, http.StatusUnauthorized, "otp_expired", "OTP has expired")
		case errors.Is(err, usecases.ErrUserInactive):
			writeJSONError(w, http.StatusForbidden, "user_inactive", "User account is inactive")
		case errors.Is(err, usecases.ErrWebAccessDenied):
			writeJSONError(w, http.StatusForbidden, "web_access_denied", "Web portal is for administrators only. Please use the mobile app.")
		default:
			slog.Error("auth_handler.VerifyOTP.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to verify OTP")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		User:         output.User,
	})
}

type loginPasswordRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *AuthHandler) LoginPassword(w http.ResponseWriter, r *http.Request) {
	var req loginPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("auth_handler.LoginPassword.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Phone == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Phone and password are required")
		return
	}

	output, err := h.loginPasswordUC.Execute(r.Context(), usecases.LoginPasswordInput{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidCredentials):
			writeJSONError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid phone or password")
		case errors.Is(err, usecases.ErrPasswordNotSet):
			writeJSONError(w, http.StatusUnauthorized, "password_not_set", "Password not set for this account")
		case errors.Is(err, usecases.ErrUserInactive):
			writeJSONError(w, http.StatusForbidden, "user_inactive", "User account is inactive")
		case errors.Is(err, usecases.ErrWebAccessDenied):
			writeJSONError(w, http.StatusForbidden, "web_access_denied", "Web portal is for administrators only. Please use the mobile app.")
		default:
			slog.Error("auth_handler.LoginPassword.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to login")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		User:         output.User,
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("auth_handler.RefreshToken.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Refresh token is required")
		return
	}

	output, err := h.refreshTokenUC.Execute(r.Context(), usecases.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrInvalidRefreshToken) {
			writeJSONError(w, http.StatusUnauthorized, "invalid_refresh_token", "Invalid or expired refresh token")
			return
		}
		slog.Error("auth_handler.RefreshToken.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to refresh token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refreshTokenResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("auth_handler.Logout.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Refresh token is required")
		return
	}

	_ = h.logoutUC.Execute(r.Context(), usecases.LogoutInput{
		RefreshToken: req.RefreshToken,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	output, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{
		UserID: claims.UserID,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
			return
		}
		slog.Error("auth_handler.GetCurrentUser.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to get user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
