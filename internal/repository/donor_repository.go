package repository

import (
	"encoding/json"
	"os"
	"sync"

	"bloodbank/internal/model"
)

type DonorRepository interface {
	List() ([]model.Donor, error)
	Save([]model.Donor) error
}

type JSONDonorRepository struct {
	path string
	mu   sync.Mutex
}

func NewJSONDonorRepository(path string) *JSONDonorRepository {
	return &JSONDonorRepository{path: path}
}

func (repository *JSONDonorRepository) List() ([]model.Donor, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	data, err := os.ReadFile(repository.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Donor{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []model.Donor{}, nil
	}
	var donors []model.Donor
	if err := json.Unmarshal(data, &donors); err != nil {
		return nil, err
	}
	for index := range donors {
		model.UpdateDonationStatus(&donors[index])
	}
	return donors, nil
}

func (repository *JSONDonorRepository) Save(donors []model.Donor) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	data, err := json.MarshalIndent(donors, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(repository.path, data, 0644)
}
