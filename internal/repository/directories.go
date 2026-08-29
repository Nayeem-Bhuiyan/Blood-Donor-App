package repository

import "os"

func EnsureDataDirectory(dataDir, uploadsDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(uploadsDir, 0755)
}
