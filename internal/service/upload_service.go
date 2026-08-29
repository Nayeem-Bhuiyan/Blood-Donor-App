package service

import (
	"mime/multipart"

	"bloodbank/internal/repository"
)

type UploadService struct {
	repository repository.FileRepository
}

func NewUploadService(repository repository.FileRepository) *UploadService {
	return &UploadService{repository: repository}
}

func (service *UploadService) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
	return service.repository.Save(file, header)
}
