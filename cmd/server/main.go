package main

import (
	"log"
	"net/http"
	"path/filepath"

	"bloodbank/internal/config"
	"bloodbank/internal/handler"
	"bloodbank/internal/repository"
	"bloodbank/internal/service"
)

func main() {
	settings := config.Load()
	if err := ensureDirectories(settings); err != nil {
		log.Fatal(err)
	}

	donorRepository := repository.NewJSONDonorRepository(filepath.Join(settings.DataDir, "donors.json"))
	fileRepository, err := repository.NewLocalFileRepository(settings.UploadsDir)
	if err != nil {
		log.Fatal(err)
	}
	uploadService := service.NewUploadService(fileRepository)
	uploadHandler := handler.NewUploadHandler(uploadService)
	donorHandler := handler.NewDonorHandler(service.NewDonorService(donorRepository), uploadHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/donors", donorHandler.Collection)
	mux.HandleFunc("/api/donors/", donorHandler.Item)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(settings.UploadsDir))))
	mux.Handle("/", http.FileServer(http.Dir(settings.StaticDir)))

	log.Printf("Blood Bank app is running on port %s", settings.Port)
	log.Fatal(http.ListenAndServe(":"+settings.Port, mux))
}

func ensureDirectories(settings config.Config) error {
	return repository.EnsureDataDirectory(settings.DataDir, settings.UploadsDir)
}
