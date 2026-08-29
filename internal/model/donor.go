package model

import "time"

type Donor struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	BloodGroup       string `json:"bloodGroup"`
	Age              int    `json:"age"`
	Phone            string `json:"phone"`
	Location         string `json:"location"`
	LastDonation     string `json:"lastDonation"`
	ReadyForDonation string `json:"readyForDonation"`
	Availability     string `json:"availability"`
	Gender           string `json:"gender"`
	Email            string `json:"email"`
	Address          string `json:"address"`
	Notes            string `json:"notes"`
	Picture          string `json:"picture"`
}

func UpdateDonationStatus(donor *Donor) {
	donor.ReadyForDonation = ReadyForDonationFromDate(donor.LastDonation)
	donor.Availability = AvailabilityFromDate(donor.LastDonation)
}

func ReadyForDonationFromDate(lastDonation string) string {
	if lastDonation == "" {
		return ""
	}
	parsedDate, err := time.Parse("2006-01-02", lastDonation)
	if err != nil {
		return ""
	}
	return parsedDate.AddDate(0, 0, 120).Format("2006-01-02")
}

func AvailabilityFromDate(lastDonation string) string {
	if lastDonation == "" {
		return "Available"
	}
	parsedDate, err := time.Parse("2006-01-02", lastDonation)
	if err != nil {
		return "Available"
	}
	readyDate := parsedDate.AddDate(0, 0, 120)
	if !time.Now().Before(readyDate) {
		return "Available"
	}
	return "Ready for donation " + readyDate.Format("02-Jan-2006")
}
