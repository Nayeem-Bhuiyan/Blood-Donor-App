package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"bloodbank/internal/model"
	"bloodbank/internal/service"
)

type DonorHandler struct {
	service       *service.DonorService
	uploadHandler *UploadHandler
}

func NewDonorHandler(service *service.DonorService, uploadHandler *UploadHandler) *DonorHandler {
	return &DonorHandler{service: service, uploadHandler: uploadHandler}
}

func (handler *DonorHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.list(w, r)
	case http.MethodPost:
		donor, err := handler.parseDonor(r, nil)
		if err != nil {
			http.Error(w, "Invalid donor payload", http.StatusBadRequest)
			return
		}
		created, err := handler.service.Create(donor)
		if err != nil {
			http.Error(w, "Unable to save donor data", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *DonorHandler) Item(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/donors/"))
	if err != nil {
		http.Error(w, "Invalid donor id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		donor, err := handler.service.Get(id)
		if errors.Is(err, service.ErrDonorNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Unable to load donor data", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, donor)
	case http.MethodPut:
		existing, err := handler.service.Get(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		updated, err := handler.parseDonor(r, &existing)
		if err != nil {
			http.Error(w, "Invalid donor payload", http.StatusBadRequest)
			return
		}
		donor, err := handler.service.Update(id, updated)
		if err != nil {
			http.Error(w, "Unable to save donor data", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, donor)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *DonorHandler) list(w http.ResponseWriter, r *http.Request) {
	donors, err := handler.service.List()
	if err != nil {
		http.Error(w, "Unable to load donor data", http.StatusInternalServerError)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 6
	}
	total := len(donors)
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, model.DonorListResponse{Page: page, Limit: limit, Total: total, TotalPages: totalPages, Items: donors[start:end]})
}

func (handler *DonorHandler) parseDonor(r *http.Request, existing *model.Donor) (model.Donor, error) {
	donor := model.Donor{}
	if existing != nil {
		donor = *existing
	}
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return model.Donor{}, err
		}
		donor.Name = r.FormValue("name")
		donor.BloodGroup = r.FormValue("bloodGroup")
		donor.Location = r.FormValue("location")
		donor.LastDonation = r.FormValue("lastDonation")
		donor.Gender = r.FormValue("gender")
		donor.Email = r.FormValue("email")
		donor.Address = r.FormValue("address")
		donor.Notes = r.FormValue("notes")
		donor.Phone = r.FormValue("phone")
		if ageValue := r.FormValue("age"); ageValue != "" {
			age, err := strconv.Atoi(ageValue)
			if err != nil {
				return model.Donor{}, err
			}
			donor.Age = age
		}
		picture, err := handler.uploadHandler.Picture(r)
		if err != nil {
			return model.Donor{}, err
		}
		if picture != "" {
			donor.Picture = picture
		}
		return donor, nil
	}
	if err := json.NewDecoder(r.Body).Decode(&donor); err != nil {
		return model.Donor{}, err
	}
	return donor, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
