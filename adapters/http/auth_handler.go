package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type AuthHandler struct {
	requestOTPUC     *usecases.RequestOTPUseCase
	verifyOTPUC      *usecases.VerifyOTPUseCase
	loginPasswordUC  *usecases.LoginPasswordUseCase
	refreshTokenUC   *usecases.RefreshTokenUseCase
	logoutUC         *usecases.LogoutUseCase
	getCurrentUserUC *usecases.GetCurrentUserUseCase
	rateLimitService *usecases.RateLimitService
	i18nService      ports.I18nService
}

func NewAuthHandler(
	requestOTPUC *usecases.RequestOTPUseCase,
	verifyOTPUC *usecases.VerifyOTPUseCase,
	loginPasswordUC *usecases.LoginPasswordUseCase,
	refreshTokenUC *usecases.RefreshTokenUseCase,
	logoutUC *usecases.LogoutUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
	rateLimitService *usecases.RateLimitService,
	i18nService ports.I18nService,
) *AuthHandler {
	return &AuthHandler{
		requestOTPUC:     requestOTPUC,
		verifyOTPUC:      verifyOTPUC,
		loginPasswordUC:  loginPasswordUC,
		refreshTokenUC:   refreshTokenUC,
		logoutUC:         logoutUC,
		getCurrentUserUC: getCurrentUserUC,
		rateLimitService: rateLimitService,
		i18nService:      i18nService,
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

	subjectKey := authRateLimitSubject(r, req.Phone)
	result, err := h.rateLimitService.AllowOTPRequest(r.Context(), subjectKey)
	if err != nil {
		slog.Error("auth_handler.RequestOTP.rate_limit", "error", err)
		writeLocalizedError(w, http.StatusInternalServerError, domain.NewLocalizedError("internal_error", "error.general.internal_error", nil), h.i18nService, r.Context())
		return
	}
	if !result.Allowed {
		writeRateLimitedError(w, r, domain.ErrOTPRequestRateLimited, result.RetryAfter, h.i18nService)
		return
	}

	_, err = h.requestOTPUC.Execute(r.Context(), usecases.RequestOTPInput{
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

	subjectKey := authRateLimitSubject(r, req.Phone)
	checkResult, err := h.rateLimitService.CheckOTPVerify(r.Context(), subjectKey)
	if err != nil {
		slog.Error("auth_handler.VerifyOTP.rate_limit_check", "error", err)
		writeLocalizedError(w, http.StatusInternalServerError, domain.NewLocalizedError("internal_error", "error.general.internal_error", nil), h.i18nService, r.Context())
		return
	}
	if !checkResult.Allowed {
		writeRateLimitedError(w, r, domain.ErrOTPVerifyRateLimited, checkResult.RetryAfter, h.i18nService)
		return
	}

	output, err := h.verifyOTPUC.Execute(r.Context(), usecases.VerifyOTPInput{
		Phone: req.Phone,
		OTP:   req.OTP,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrInvalidOTP) {
			result, rateLimitErr := h.rateLimitService.RegisterOTPVerifyFailure(r.Context(), subjectKey)
			if rateLimitErr != nil {
				slog.Error("auth_handler.VerifyOTP.register_failure", "error", rateLimitErr)
			} else if !result.Allowed {
				writeRateLimitedError(w, r, domain.ErrOTPVerifyRateLimited, result.RetryAfter, h.i18nService)
				return
			}
		}
		var statusCode int
		switch {
		case errors.Is(err, usecases.ErrInvalidOTP):
			statusCode = http.StatusUnauthorized
			err = domain.NewLocalizedError("invalid_otp", "error.auth.invalid_otp", nil)
		case errors.Is(err, usecases.ErrOTPExpired):
			statusCode = http.StatusUnauthorized
			err = domain.NewLocalizedError("otp_expired", "error.auth.otp_expired", nil)
		case errors.Is(err, usecases.ErrUserInactive):
			statusCode = http.StatusForbidden
			err = domain.NewLocalizedError("user_inactive", "error.auth.account_disabled", nil)
		case errors.Is(err, usecases.ErrWebAccessDenied):
			statusCode = http.StatusForbidden
			err = domain.NewLocalizedError("web_access_denied", "error.auth.unauthorized", nil)
		default:
			slog.Error("auth_handler.VerifyOTP.execute_usecase", "error", err)
			statusCode = http.StatusInternalServerError
			err = domain.NewLocalizedError("internal_error", "error.general.internal_error", nil)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	if err := h.rateLimitService.ResetOTPVerify(r.Context(), subjectKey); err != nil {
		slog.Error("auth_handler.VerifyOTP.reset_rate_limit", "error", err)
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

	subjectKey := authRateLimitSubject(r, req.Phone)
	checkResult, err := h.rateLimitService.CheckLogin(r.Context(), subjectKey)
	if err != nil {
		slog.Error("auth_handler.LoginPassword.rate_limit_check", "error", err)
		writeLocalizedError(w, http.StatusInternalServerError, domain.NewLocalizedError("internal_error", "error.general.internal_error", nil), h.i18nService, r.Context())
		return
	}
	if !checkResult.Allowed {
		writeRateLimitedError(w, r, domain.ErrLoginRateLimited, checkResult.RetryAfter, h.i18nService)
		return
	}

	output, err := h.loginPasswordUC.Execute(r.Context(), usecases.LoginPasswordInput{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		if usecases.IsCredentialRateLimitError(err) {
			result, rateLimitErr := h.rateLimitService.RegisterLoginFailure(r.Context(), subjectKey)
			if rateLimitErr != nil {
				slog.Error("auth_handler.LoginPassword.register_failure", "error", rateLimitErr)
			} else if !result.Allowed {
				writeRateLimitedError(w, r, domain.ErrLoginRateLimited, result.RetryAfter, h.i18nService)
				return
			}
		}
		var statusCode int
		switch {
		case errors.Is(err, usecases.ErrInvalidCredentials):
			statusCode = http.StatusUnauthorized
			err = domain.NewLocalizedError("invalid_credentials", "error.auth.invalid_credentials", nil)
		case errors.Is(err, usecases.ErrPasswordNotSet):
			statusCode = http.StatusUnauthorized
			err = domain.NewLocalizedError("password_not_set", "error.auth.invalid_credentials", nil)
		case errors.Is(err, usecases.ErrUserInactive):
			statusCode = http.StatusForbidden
			err = domain.NewLocalizedError("user_inactive", "error.auth.account_disabled", nil)
		case errors.Is(err, usecases.ErrWebAccessDenied):
			statusCode = http.StatusForbidden
			err = domain.NewLocalizedError("web_access_denied", "error.auth.unauthorized", nil)
		default:
			slog.Error("auth_handler.LoginPassword.execute_usecase", "error", err)
			statusCode = http.StatusInternalServerError
			err = domain.NewLocalizedError("internal_error", "error.general.internal_error", nil)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	if err := h.rateLimitService.ResetLogin(r.Context(), subjectKey); err != nil {
		slog.Error("auth_handler.LoginPassword.reset_rate_limit", "error", err)
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
		var statusCode int
		switch {
		case errors.Is(err, usecases.ErrInvalidRefreshToken):
			statusCode = http.StatusUnauthorized
			err = domain.NewLocalizedError("invalid_refresh_token", "error.auth.token_expired", nil)
		default:
			slog.Error("auth_handler.RefreshToken.execute_usecase", "error", err)
			statusCode = http.StatusInternalServerError
			err = domain.NewLocalizedError("internal_error", "error.general.internal_error", nil)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
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

func authRateLimitSubject(r *http.Request, phone string) string {
	ip := clientIP(r)
	normalizedPhone := usecases.NormalizePhoneForRateLimit(phone)
	if ip == "" {
		return normalizedPhone
	}
	if normalizedPhone == "" {
		return ip
	}
	return ip + "|" + normalizedPhone
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
