package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type DocumentHandler struct {
	generateUploadURLUC   *usecases.GenerateDocumentUploadURLUseCase
	generateDownloadURLUC *usecases.GenerateDocumentDownloadURLUseCase
}

func NewDocumentHandler(
	generateUploadURLUC *usecases.GenerateDocumentUploadURLUseCase,
	generateDownloadURLUC *usecases.GenerateDocumentDownloadURLUseCase,
) *DocumentHandler {
	return &DocumentHandler{
		generateUploadURLUC:   generateUploadURLUC,
		generateDownloadURLUC: generateDownloadURLUC,
	}
}

type generateUploadURLRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type generateDownloadURLRequest struct {
	ObjectKey        string `json:"objectKey"`
	DownloadFileName string `json:"downloadFileName,omitempty"`
}

func (h *DocumentHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	var req generateUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	out, err := h.generateUploadURLUC.Execute(r.Context(), usecases.GenerateDocumentUploadURLInput{
		UserUID:     claims.UserUID,
		FileName:    req.FileName,
		ContentType: req.ContentType,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidFilename):
			writeJSONError(w, http.StatusBadRequest, "invalid_filename", err.Error())
		case errors.Is(err, usecases.ErrInvalidContentType):
			writeJSONError(w, http.StatusBadRequest, "invalid_content_type", err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to generate upload URL")
		}
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *DocumentHandler) GenerateDownloadURL(w http.ResponseWriter, r *http.Request) {
	var req generateDownloadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	out, err := h.generateDownloadURLUC.Execute(r.Context(), usecases.GenerateDocumentDownloadURLInput{
		ObjectKey:        req.ObjectKey,
		DownloadFileName: req.DownloadFileName,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidObjectKey):
			writeJSONError(w, http.StatusBadRequest, "invalid_object_key", err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to generate download URL")
		}
		return
	}

	writeJSON(w, http.StatusOK, out)
}