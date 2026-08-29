package handler

import (
	"errors"
	"net/http"

	"bloodbank/internal/service"
)

type UploadHandler struct {
	service *service.UploadService
}

func NewUploadHandler(service *service.UploadService) *UploadHandler {
	return &UploadHandler{service: service}
}

func (handler *UploadHandler) Picture(r *http.Request) (string, error) {
	file, header, err := r.FormFile("picture")
	if errors.Is(err, http.ErrMissingFile) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	return handler.service.Save(file, header)
}
