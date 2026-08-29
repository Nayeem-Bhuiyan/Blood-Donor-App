package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxPictureSize = 50 * 1024

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

func appRoot() string {
	if wd, err := os.Getwd(); err == nil {
		if _, statErr := os.Stat(filepath.Join(wd, "donors.json")); statErr == nil {
			return wd
		}
	}

	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		if _, statErr := os.Stat(filepath.Join(dir, "donors.json")); statErr == nil {
			return dir
		}
	}

	return "."
}

func uploadsDir() string {
	dir := filepath.Join(appRoot(), "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create upload directory: %v", err)
	}
	return dir
}

func saveUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	if header == nil || header.Filename == "" {
		return "", nil
	}

	source, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return "", err
	}

	decoded, _, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("decode picture: %w", err)
	}

	compressed, err := compressPicture(decoded)
	if err != nil {
		return "", err
	}
	if len(compressed) > maxPictureSize {
		return "", fmt.Errorf("compressed picture exceeds %d KB", maxPictureSize/1024)
	}

	safeName := fmt.Sprintf("donor-%d-%d.jpg", os.Getpid(), time.Now().UnixNano())
	targetPath := filepath.Join(uploadsDir(), safeName)

	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := out.Write(compressed); err != nil {
		return "", err
	}

	return safeName, nil
}

func compressPicture(source image.Image) ([]byte, error) {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	maxDimension := 1600
	if width > maxDimension || height > maxDimension {
		scale := float64(maxDimension) / float64(max(width, height))
		width = max(1, int(float64(width)*scale))
		height = max(1, int(float64(height)*scale))
		source = resizePicture(source, width, height)
	}

	for {
		for quality := 85; quality >= 20; quality -= 5 {
			var output bytes.Buffer
			if err := jpeg.Encode(&output, source, &jpeg.Options{Quality: quality}); err != nil {
				return nil, fmt.Errorf("encode picture: %w", err)
			}
			if output.Len() <= maxPictureSize {
				return output.Bytes(), nil
			}
		}

		width = max(1, int(float64(width)*0.8))
		height = max(1, int(float64(height)*0.8))
		source = resizePicture(source, width, height)
	}
}

func resizePicture(source image.Image, width, height int) image.Image {
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(resized, resized.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	sourceBounds := source.Bounds()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceX := sourceBounds.Min.X + x*sourceBounds.Dx()/width
			sourceY := sourceBounds.Min.Y + y*sourceBounds.Dy()/height
			resized.Set(x, y, source.At(sourceX, sourceY))
		}
	}

	return resized
}

func donorReadyForDonationFromDate(lastDonation string) string {
	if lastDonation == "" {
		return ""
	}

	parsedDate, err := time.Parse("2006-01-02", lastDonation)
	if err != nil {
		return ""
	}

	readyDate := parsedDate.AddDate(0, 0, 120)
	return readyDate.Format("2006-01-02")
}

func donorAvailabilityFromDate(lastDonation string) string {
	if lastDonation == "" {
		return "Available"
	}

	parsedDate, err := time.Parse("2006-01-02", lastDonation)
	if err != nil {
		return "Available"
	}

	readyDate := parsedDate.AddDate(0, 0, 120)
	if time.Now().After(readyDate) || time.Now().Equal(readyDate) {
		return "Available"
	}
	return "Ready for donation " + readyDate.Format("02-Jan-2006")
}

func parseDonorFromRequest(r *http.Request, existing *Donor) (Donor, error) {
	donor := Donor{}

	if existing != nil {
		donor = *existing
	}

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return Donor{}, err
		}

		donor.Name = r.FormValue("name")
		donor.BloodGroup = r.FormValue("bloodGroup")
		donor.Location = r.FormValue("location")
		donor.LastDonation = r.FormValue("lastDonation")
		donor.ReadyForDonation = donorReadyForDonationFromDate(donor.LastDonation)
		donor.Availability = donorAvailabilityFromDate(donor.LastDonation)
		donor.Gender = r.FormValue("gender")
		donor.Email = r.FormValue("email")
		donor.Address = r.FormValue("address")
		donor.Notes = r.FormValue("notes")

		ageValue := r.FormValue("age")
		if ageValue != "" {
			age, err := strconv.Atoi(ageValue)
			if err != nil {
				return Donor{}, err
			}
			donor.Age = age
		}

		donor.Phone = r.FormValue("phone")

		file, header, err := r.FormFile("picture")
		if err == nil {
			defer file.Close()
			savedName, saveErr := saveUploadedFile(file, header)
			if saveErr != nil {
				return Donor{}, saveErr
			}
			donor.Picture = savedName
		} else if err != http.ErrMissingFile {
			return Donor{}, err
		}

		return donor, nil
	}

	if err := json.NewDecoder(r.Body).Decode(&donor); err != nil {
		return Donor{}, err
	}

	donor.ReadyForDonation = donorReadyForDonationFromDate(donor.LastDonation)
	donor.Availability = donorAvailabilityFromDate(donor.LastDonation)
	return donor, nil
}

func loadDonors() ([]Donor, error) {
	dataFile := filepath.Join(appRoot(), "donors.json")
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Donor{}, nil
		}
		return nil, err
	}

	var donors []Donor
	if len(data) == 0 {
		return []Donor{}, nil
	}

	if err := json.Unmarshal(data, &donors); err != nil {
		return nil, err
	}

	for i := range donors {
		donors[i].ReadyForDonation = donorReadyForDonationFromDate(donors[i].LastDonation)
		donors[i].Availability = donorAvailabilityFromDate(donors[i].LastDonation)
	}

	return donors, nil
}

func saveDonors(donors []Donor) error {
	data, err := json.MarshalIndent(donors, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(appRoot(), "donors.json"), data, 0644)
}

func nextDonorID(donors []Donor) int {
	maxID := 0
	for _, donor := range donors {
		if donor.ID > maxID {
			maxID = donor.ID
		}
	}
	return maxID + 1
}

func donorsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		donors, err := loadDonors()
		if err != nil {
			http.Error(w, "Unable to load donor data", http.StatusInternalServerError)
			return
		}

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 {
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

		response := map[string]any{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
			"items":      donors[start:end],
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Failed to encode donors response:", err)
		}

	case http.MethodPost:
		donor, err := parseDonorFromRequest(r, nil)
		if err != nil {
			http.Error(w, "Invalid donor payload", http.StatusBadRequest)
			return
		}

		donors, err := loadDonors()
		if err != nil {
			http.Error(w, "Unable to load donor data", http.StatusInternalServerError)
			return
		}

		donor.ID = nextDonorID(donors)
		donors = append(donors, donor)

		if err := saveDonors(donors); err != nil {
			http.Error(w, "Unable to save donor data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(donor); err != nil {
			log.Println("Failed to encode created donor response:", err)
		}

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func donorByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/donors/")
	if idStr == "" || idStr == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid donor id", http.StatusBadRequest)
		return
	}

	donors, err := loadDonors()
	if err != nil {
		http.Error(w, "Unable to load donor data", http.StatusInternalServerError)
		return
	}

	var donor Donor
	index := -1
	for i, item := range donors {
		if item.ID == id {
			donor = item
			index = i
			break
		}
	}

	if index == -1 {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPut:
		updatedDonor, err := parseDonorFromRequest(r, &donor)
		if err != nil {
			http.Error(w, "Invalid donor payload", http.StatusBadRequest)
			return
		}
		updatedDonor.ID = id
		donors[index] = updatedDonor
		donor = updatedDonor

		if err := saveDonors(donors); err != nil {
			http.Error(w, "Unable to save donor data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(donor); err != nil {
			log.Println("Failed to encode updated donor response:", err)
		}

	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(donor); err != nil {
			log.Println("Failed to encode donor response:", err)
		}

	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	root := appRoot()
	uploadsRoot := filepath.Join(root, "uploads")
	if err := os.MkdirAll(uploadsRoot, 0755); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/donors", donorsHandler)
	http.HandleFunc("/api/donors/", donorByIDHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsRoot))))
	http.Handle("/", http.FileServer(http.Dir(filepath.Join(root, "static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	fmt.Printf("Blood Bank app is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
