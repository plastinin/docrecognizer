package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

// Поддерживаемые MIME типы
var supportedContentTypes = map[string]bool{
	"image/png":       true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/webp":      true,
	"image/tiff":      true,
	"application/pdf": true,
}

// Маппинг расширений на MIME типы
var extToContentType = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".tiff": "image/tiff",
	".tif":  "image/tiff",
	".pdf":  "application/pdf",
}

// ValidateContentType проверяет поддерживается ли тип файла
func ValidateContentType(contentType string) error {
	// Убираем параметры типа charset
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(strings.ToLower(ct))

	if !supportedContentTypes[ct] {
		return ErrUnsupportedFileType
	}
	return nil
}

// ContentTypeFromFileName определяет MIME тип по имени файла
func ContentTypeFromFileName(fileName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	ct, ok := extToContentType[ext]
	if !ok {
		return "", ErrUnsupportedFileType
	}
	return ct, nil
}

// IsImage проверяет, является ли файл изображением
func IsImage(contentType string) bool {
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(strings.ToLower(ct))
	return strings.HasPrefix(ct, "image/")
}

// IsPDF проверяет, является ли файл PDF
func IsPDF(contentType string) bool {
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(strings.ToLower(ct))
	return ct == "application/pdf"
}