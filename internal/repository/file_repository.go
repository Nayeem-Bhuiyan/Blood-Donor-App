package repository

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

const maxPictureSize = 50 * 1024

type FileRepository interface {
	Save(multipart.File, *multipart.FileHeader) (string, error)
}

type LocalFileRepository struct {
	directory string
}

func NewLocalFileRepository(directory string) (*LocalFileRepository, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}
	return &LocalFileRepository{directory: directory}, nil
}

func (repository *LocalFileRepository) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
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

	name := fmt.Sprintf("donor-%d-%d.jpg", os.Getpid(), time.Now().UnixNano())
	if err := os.WriteFile(filepath.Join(repository.directory, name), compressed, 0644); err != nil {
		return "", err
	}
	return name, nil
}

func compressPicture(source image.Image) ([]byte, error) {
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	if width > 1600 || height > 1600 {
		scale := 1600.0 / float64(max(width, height))
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
	bounds := source.Bounds()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceX := bounds.Min.X + x*bounds.Dx()/width
			sourceY := bounds.Min.Y + y*bounds.Dy()/height
			resized.Set(x, y, source.At(sourceX, sourceY))
		}
	}
	return resized
}
