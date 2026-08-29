package model

type DonorListResponse struct {
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int     `json:"total"`
	TotalPages int     `json:"totalPages"`
	Items      []Donor `json:"items"`
}
