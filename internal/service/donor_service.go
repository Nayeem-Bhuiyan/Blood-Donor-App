package service

import (
	"errors"

	"bloodbank/internal/model"
	"bloodbank/internal/repository"
)

var ErrDonorNotFound = errors.New("donor not found")

type DonorService struct {
	repository repository.DonorRepository
}

func NewDonorService(repository repository.DonorRepository) *DonorService {
	return &DonorService{repository: repository}
}

func (service *DonorService) List() ([]model.Donor, error) {
	return service.repository.List()
}

func (service *DonorService) Create(donor model.Donor) (model.Donor, error) {
	donors, err := service.repository.List()
	if err != nil {
		return model.Donor{}, err
	}
	donor.ID = nextID(donors)
	model.UpdateDonationStatus(&donor)
	donors = append(donors, donor)
	if err := service.repository.Save(donors); err != nil {
		return model.Donor{}, err
	}
	return donor, nil
}

func (service *DonorService) Get(id int) (model.Donor, error) {
	donors, err := service.repository.List()
	if err != nil {
		return model.Donor{}, err
	}
	for _, donor := range donors {
		if donor.ID == id {
			return donor, nil
		}
	}
	return model.Donor{}, ErrDonorNotFound
}

func (service *DonorService) Update(id int, updated model.Donor) (model.Donor, error) {
	donors, err := service.repository.List()
	if err != nil {
		return model.Donor{}, err
	}
	for index := range donors {
		if donors[index].ID == id {
			updated.ID = id
			model.UpdateDonationStatus(&updated)
			donors[index] = updated
			if err := service.repository.Save(donors); err != nil {
				return model.Donor{}, err
			}
			return updated, nil
		}
	}
	return model.Donor{}, ErrDonorNotFound
}

func nextID(donors []model.Donor) int {
	maxID := 0
	for _, donor := range donors {
		if donor.ID > maxID {
			maxID = donor.ID
		}
	}
	return maxID + 1
}
