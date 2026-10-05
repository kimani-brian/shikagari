package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/config"
)

// Allowed image MIME types
var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// Allowed e-logbook MIME types — NTSA issues the digital logbook as an image
// or a PDF, so both are accepted for seller verification.
var allowedELogbookMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// UploadService handles image file uploads to local storage.
// Swap the internals for an S3/GCS implementation in production.
type UploadService struct {
	cfg *config.Config
}

// NewUploadService constructs an UploadService.
func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

// UploadImage validates and saves an uploaded image file.
// Returns the public URL of the saved file.
func (s *UploadService) UploadImage(
	file multipart.File,
	header *multipart.FileHeader,
	folder string,
) (string, error) {
	// ── 1. Validate file size ─────────────────────────────────────────────────
	maxBytes := s.cfg.Upload.MaxFileSizeMB * 1024 * 1024
	if header.Size > maxBytes {
		return "", fmt.Errorf("file size exceeds the %dMB limit", s.cfg.Upload.MaxFileSizeMB)
	}

	// ── 2. Validate MIME type ─────────────────────────────────────────────────
	mimeType := header.Header.Get("Content-Type")
	if !allowedMIMETypes[mimeType] {
		return "", errors.New("only JPEG, PNG, and WebP images are accepted")
	}

	// ── 3. Build a unique filename ────────────────────────────────────────────
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = mimeToExtension(mimeType)
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.New().String(), ext)

	// ── 4. Ensure the destination directory exists ────────────────────────────
	destDir := filepath.Join(s.cfg.Upload.Dir, folder)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", errors.New("failed to prepare upload directory")
	}

	// ── 5. Write the file ─────────────────────────────────────────────────────
	destPath := filepath.Join(destDir, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", errors.New("failed to save uploaded file")
	}
	defer destFile.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return "", errors.New("failed to write file data")
			}
		}
		if err != nil {
			break
		}
	}

	// ── 6. Build and return the public URL ────────────────────────────────────
	publicURL := fmt.Sprintf("/uploads/%s/%s", folder, filename)
	return publicURL, nil
}

// UploadELogbook validates and stores a seller's NTSA e-logbook.
// Accepts JPEG, PNG, WebP or PDF up to the configured upload size limit.
// Returns the public URL of the stored file.
func (s *UploadService) UploadELogbook(
	file multipart.File,
	header *multipart.FileHeader,
) (string, error) {
	maxBytes := s.cfg.Upload.MaxFileSizeMB * 1024 * 1024
	if header.Size > maxBytes {
		return "", fmt.Errorf("file size exceeds the %dMB limit", s.cfg.Upload.MaxFileSizeMB)
	}

	mimeType := header.Header.Get("Content-Type")
	if !allowedELogbookMIMETypes[mimeType] {
		return "", errors.New("NTSA e-logbook must be a JPEG, PNG, WebP, or PDF file")
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = mimeToExtension(mimeType)
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.New().String(), ext)

	destDir := filepath.Join(s.cfg.Upload.Dir, "elogbooks")
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", errors.New("failed to prepare upload directory")
	}

	destPath := filepath.Join(destDir, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", errors.New("failed to save uploaded file")
	}
	defer destFile.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return "", errors.New("failed to write file data")
			}
		}
		if readErr != nil {
			break
		}
	}

	return fmt.Sprintf("/uploads/elogbooks/%s", filename), nil
}

// mimeToExtension returns a file extension for a given MIME type.
func mimeToExtension(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	default:
		return ".jpg"
	}
}
