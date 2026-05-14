package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// localizedErrorResponse is the JSON body for localized error responses.
type localizedErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeRateLimitedError(w http.ResponseWriter, r *http.Request, err error, retryAfter time.Duration, i18nService ports.I18nService) {
	seconds := int(retryAfter.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeLocalizedError(w, http.StatusTooManyRequests, err, i18nService, r.Context())
}

// writeLocalizedError inspects err for a *domain.LocalizedError, translates it using
// i18nService (reading locale from ctx), and writes a JSON error response.
// If err is not a LocalizedError it falls back to err.Error() as the message.
func writeLocalizedError(w http.ResponseWriter, status int, err error, i18nService ports.I18nService, ctx context.Context) {
	var le *domain.LocalizedError
	if errors.As(err, &le) {
		var msg string
		if len(le.Params) > 0 {
			msg = i18nService.TWithParams(ctx, le.Key, le.Params)
		} else {
			msg = i18nService.T(ctx, le.Key)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(localizedErrorResponse{Error: le.Code, Message: msg})
		return
	}
	// Fallback for non-localized errors
	writeJSONError(w, status, "internal_error", err.Error())
}
