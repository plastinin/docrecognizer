package llm

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/gen2brain/go-fitz"
)

// PDFConverter конвертирует PDF в изображения
type PDFConverter struct{}

// NewPDFConverter создаёт новый конвертер
func NewPDFConverter() *PDFConverter {
	return &PDFConverter{}
}

// ConvertFirstPage конвертирует первую страницу PDF в PNG
func (c *PDFConverter) ConvertFirstPage(pdfData []byte) ([]byte, error) {
	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	if doc.NumPage() == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Конвертируем первую страницу
	img, err := doc.Image(0)
	if err != nil {
		return nil, fmt.Errorf("failed to render page: %w", err)
	}

	// Кодируем в PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// ConvertAllPages конвертирует все страницы PDF в изображения
func (c *PDFConverter) ConvertAllPages(pdfData []byte) ([][]byte, error) {
	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	numPages := doc.NumPage()
	if numPages == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	images := make([][]byte, 0, numPages)
	for i := 0; i < numPages; i++ {
		img, err := doc.Image(i)
		if err != nil {
			return nil, fmt.Errorf("failed to render page %d: %w", i, err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode page %d: %w", i, err)
		}

		images = append(images, buf.Bytes())
	}

	return images, nil
}

// ImageToBytes конвертирует image.Image в PNG bytes
func ImageToBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadAllBytes читает все байты из reader
func ReadAllBytes(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}