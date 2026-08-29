package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AppRoot    string
	StaticDir  string
	DataDir    string
	UploadsDir string
	Port       string
}

func Load() Config {
	root := applicationRoot()
	dataDir := strings.TrimSpace(os.Getenv("DATA_DIR"))
	if dataDir == "" {
		if _, err := os.Stat(filepath.FromSlash("/data")); err == nil {
			dataDir = filepath.FromSlash("/data")
		} else {
			localData := filepath.Join(root, "data")
			if _, err := os.Stat(localData); err == nil {
				dataDir = localData
			} else if _, err := os.Stat(filepath.Join(root, "donors.json")); err == nil {
				dataDir = root
			} else {
				dataDir = localData
			}
		}
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8084"
	}
	uploadsDir := strings.TrimSpace(os.Getenv("UPLOADS_DIR"))
	if uploadsDir == "" {
		uploadsDir = filepath.Join(root, "uploads")
	}

	return Config{
		AppRoot:    root,
		StaticDir:  filepath.Join(root, "static"),
		DataDir:    dataDir,
		UploadsDir: uploadsDir,
		Port:       port,
	}
}

func applicationRoot() string {
	if root, err := os.Getwd(); err == nil {
		return root
	}
	if executable, err := os.Executable(); err == nil {
		return filepath.Dir(executable)
	}
	return "."
}
