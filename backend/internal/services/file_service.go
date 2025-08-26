package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileService struct {
	uploadDir string
}

func NewFileService() *FileService {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}
	
	return &FileService{
		uploadDir: uploadDir,
	}
}

func (s *FileService) SaveInvoiceDocument(file multipart.File, header *multipart.FileHeader, invoiceID string) (string, error) {
	// Create invoice-specific directory
	invoiceDir := filepath.Join(s.uploadDir, "invoices", invoiceID)
	if err := os.MkdirAll(invoiceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create invoice directory: %v", err)
	}
	
	// Generate unique filename
	timestamp := time.Now().Unix()
	fileExt := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%d_%s%s", timestamp, uuid.New().String()[:8], fileExt)
	filePath := filepath.Join(invoiceDir, uniqueFilename)
	
	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()
	
	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}
	
	// Return relative path for storage in database
	relativePath := filepath.Join("invoices", invoiceID, uniqueFilename)
	return strings.ReplaceAll(relativePath, "\\", "/"), nil // Normalize path separators
}

func (s *FileService) GetFilePath(relativePath string) string {
	return filepath.Join(s.uploadDir, relativePath)
}

func (s *FileService) DeleteFile(relativePath string) error {
	fullPath := s.GetFilePath(relativePath)
	return os.Remove(fullPath)
}

func (s *FileService) FileExists(relativePath string) bool {
	fullPath := s.GetFilePath(relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}
